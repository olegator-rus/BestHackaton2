/*
 * NETCAP - Traffic Analysis Framework
 * Copyright (c) 2017-2020 Philipp Mieden <dreadl0ck [at] protonmail [dot] ch>
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

// Package collector provides a mechanism to collect network packets from a network interface on macOS, linux and windows
package collector

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/dreadl0ck/gopacket"
	"github.com/dustin/go-humanize"
	"github.com/evilsocket/islazy/tui"
	"github.com/mgutz/ansi"

	"github.com/dreadl0ck/netcap"
	"github.com/dreadl0ck/netcap/decoder"
	"github.com/dreadl0ck/netcap/reassembly"
	"github.com/dreadl0ck/netcap/utils"
)

// errInvalidOutputDirectory indicates that a file path was supplied instead of a directory.
var errInvalidOutputDirectory = errors.New("expected a directory, but got a file for output path")

// Collector provides an interface to collect data from PCAP or a network interface.
// this structure has an optimized field order to avoid excessive padding.
type Collector struct {
	workers                  []chan *packet
	start                    time.Time
	assemblers               []*reassembly.Assembler
	customDecoders           []decoder.CustomDecoderAPI
	progressString           string
	next                     int
	unkownPcapWriterAtomic   *atomicPcapGoWriter
	unknownPcapFile          *os.File
	errorsPcapWriterBuffered *bufio.Writer
	errorsPcapWriterAtomic   *atomicPcapGoWriter
	errorsPcapFile           *os.File
	errorLogFile             *os.File
	unknownProtosAtomic      *decoder.AtomicCounterMap
	allProtosAtomic          *decoder.AtomicCounterMap
	current                  int64
	numWorkers               int
	numPacketsLast           int64
	totalBytesWritten        int64
	files                    map[string]string
	inputSize                int64
	unkownPcapWriterBuffered *bufio.Writer
	numPackets               int64
	config                   *Config
	errorMap                 *decoder.AtomicCounterMap
	goPacketDecoders         map[gopacket.LayerType][]*decoder.GoPacketDecoder
	wg                       sync.WaitGroup
	mu                       sync.Mutex
	statMutex                sync.Mutex
	shutdown                 bool
	isLive                   bool
}

// New returns a new Collector instance.
func New(config Config) *Collector {
	if config.OutDirPermission == 0 {
		config.OutDirPermission = os.FileMode(outDirPermissionDefault)
	}

	return &Collector{
		next:                1,
		unknownProtosAtomic: decoder.NewAtomicCounterMap(),
		allProtosAtomic:     decoder.NewAtomicCounterMap(),
		errorMap:            decoder.NewAtomicCounterMap(),
		files:               map[string]string{},
		config:              &config,
		start:               time.Now(),
	}
}

// stopWorkers halts all workers.
func (c *Collector) stopWorkers() {
	// wait until all packets have been decoded
	c.mu.Lock()
	for i, w := range c.workers {
		select {
		case w <- nil:
		case <-time.After(5 * time.Second):
			fmt.Println("worker", i, "seems stuck, skipping...")
		}
	}
	c.mu.Unlock()
}

// handleSignals catches signals and runs the cleanup
// SIGQUIT is not catched, to allow debugging by producing a stack and goroutine trace.
func (c *Collector) handleSignals() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// start signal handler and cleanup routine
	go func() {
		sig := <-sigs

		fmt.Println("\nreceived signal:", sig)
		fmt.Println("exiting")

		go func() {
			sign := <-sigs
			fmt.Println("force quitting, signal:", sign)
			os.Exit(0)
		}()

		c.cleanup(true)
		os.Exit(0)
	}()
}

// to decode incoming packets in parallel
// they are passed to several worker goroutines in round robin style.
func (c *Collector) handlePacket(p *packet) {
	// make it work for 1 worker only, can be used for debugging
	if c.numWorkers == 1 {
		c.workers[0] <- p

		return
	}

	// send the packetInfo to the encoder routine
	c.workers[c.next] <- p

	// increment or reset next
	if c.config.Workers == c.next+1 {
		// reset
		c.next = 0
	} else {
		c.next++
	}
}

// to decode incoming packets in parallel
// they are passed to several worker goroutines in round robin style.
func (c *Collector) handlePacketTimeout(p *packet) {
	select {
	// send the packetInfo to the encoder routine
	case c.workers[c.next] <- p:
	case <-time.After(3 * time.Second):
		pkt := gopacket.NewPacket(p.data, c.config.BaseLayer, gopacket.Default)

		var (
			nf gopacket.Flow
			tf gopacket.Flow
		)

		if nl := pkt.NetworkLayer(); nl != nil {
			nf = nl.NetworkFlow()
		}

		if tl := pkt.TransportLayer(); tl != nil {
			tf = tl.TransportFlow()
		}

		fmt.Println("handle packet timeout", nf, tf)
	}

	// increment or reset next
	if c.config.Workers == c.next+1 {
		// reset
		c.next = 0
	} else {
		c.next++
	}
}

// print errors to stdout in red.
func (c *Collector) printErrors() {
	c.errorMap.Lock()
	if len(c.errorMap.Items) > 0 {
		fmt.Println("")

		for msg, count := range c.errorMap.Items {
			fmt.Println(ansi.Red, "[ERROR]", msg, "COUNT:", count, ansi.Reset)
		}

		fmt.Println("")
	}
	c.errorMap.Unlock()
}

// closes the logfile for errors.
func (c *Collector) closeErrorLogFile() {
	c.errorMap.Lock()

	// append  stats
	var stats string
	for msg, count := range c.errorMap.Items {
		stats += fmt.Sprintln("[ERROR]", msg, "COUNT:", count)
	}

	c.errorMap.Unlock()

	c.mu.Lock()

	_, err := c.errorLogFile.WriteString(stats)
	if err != nil {
		utils.DebugLog.Println("failed to write stats into error log:", err)

		return
	}

	// sync
	err = c.errorLogFile.Sync()
	if err != nil {
		utils.DebugLog.Println("failed to sync error log:", err)

		return
	}

	// close file handle
	err = c.errorLogFile.Close()
	if err != nil {
		utils.DebugLog.Println("failed to close error log:", err)

		return
	}

	c.mu.Unlock()
}

// stats prints collector statistics.
func (c *Collector) stats() {
	var target io.Writer
	if c.config.Quiet {
		target = logFileHandle
	} else {
		target = io.MultiWriter(os.Stderr, logFileHandle)
	}

	var rows [][]string

	c.unknownProtosAtomic.Lock()

	for k, v := range c.allProtosAtomic.Items {
		if k == "Payload" {
			rows = append(rows, []string{k, fmt.Sprint(v), share(v, c.numPackets)})

			continue
		}

		if _, ok := c.unknownProtosAtomic.Items[k]; ok {
			rows = append(rows, []string{"*" + k, fmt.Sprint(v), share(v, c.numPackets)})
		} else {
			rows = append(rows, []string{k, fmt.Sprint(v), share(v, c.numPackets)})
		}
	}

	numUnknown := len(c.unknownProtosAtomic.Items)

	c.unknownProtosAtomic.Unlock()
	tui.Table(target, []string{"Layer", "NumRecords", "Share"}, rows)

	// print legend if there are unknown protos
	// -1 for "Payload" layer
	if numUnknown-1 > 0 {
		if !c.config.Quiet {
			fmt.Println("* protocol supported by gopacket, but not implemented in netcap")
		}
	}

	if len(c.customDecoders) > 0 {
		rows = [][]string{}
		for _, d := range c.customDecoders {
			rows = append(rows, []string{d.GetName(), strconv.FormatInt(d.NumRecords(), 10), share(d.NumRecords(), c.numPackets)})
		}

		tui.Table(target, []string{"CustomDecoder", "NumRecords", "Share"}, rows)
	}

	res := "\n-> total bytes of data written to disk: " + humanize.Bytes(uint64(c.totalBytesWritten)) + "\n"

	if c.unkownPcapWriterAtomic != nil {
		if c.unkownPcapWriterAtomic.count > 0 {
			res += "-> " + share(c.unkownPcapWriterAtomic.count, c.numPackets) + " of packets (" + strconv.FormatInt(c.unkownPcapWriterAtomic.count, 10) + ") written to unknown.pcap\n"
		}
	}

	if c.errorsPcapWriterAtomic != nil {
		if c.errorsPcapWriterAtomic.count > 0 {
			res += "-> " + share(c.errorsPcapWriterAtomic.count, c.numPackets) + " of packets (" + strconv.FormatInt(c.errorsPcapWriterAtomic.count, 10) + ") written to errors.pcap\n"
		}
	}

	if _, err := fmt.Fprintln(target, res); err != nil {
		fmt.Println("failed to print stats:", err)
	}

	if c.config.DecoderConfig.SaveConns {
		_, _ = fmt.Fprintln(target, "saved TCP connections:", decoder.NumSavedTCPConns())
		_, _ = fmt.Fprintln(target, "saved UDP connections:", decoder.NumSavedUDPConns())
	}
}

// updates the progress indicator and writes to stdout.
func (c *Collector) printProgress() {
	// increment atomic packet counter
	atomic.AddInt64(&c.current, 1)

	// must be locked, otherwise a race occurs when sending a SIGINT
	//  and triggering wg.Wait() in another goroutine...
	c.statMutex.Lock()

	// increment wait group for packet processing
	c.wg.Add(1)

	// dont print message when collector is about to shutdown
	if c.shutdown {
		c.statMutex.Unlock()

		return
	}
	c.statMutex.Unlock()

	if c.current%1000 == 0 {
		if !c.config.Quiet {
			// using a strings.Builder for assembling string for performance
			// TODO: could be refactored to use a byte slice with a fixed length instead
			// TODO: add Builder to collector and flush it every cycle to reduce allocations
			// also only print flows and collections when the corresponding decoders are active
			var b strings.Builder

			b.Grow(65)
			b.WriteString("decoding packets... (")
			b.WriteString(utils.Progress(c.current, c.numPackets))
			b.WriteString(")")
			// b.WriteString(strconv.Itoa(decoder.Flows.Size()))
			// b.WriteString(" connections: ")
			// b.WriteString(strconv.Itoa(decoder.Connections.Size()))
			b.WriteString(" profiles: ")
			b.WriteString(strconv.Itoa(decoder.Profiles.Size()))
			b.WriteString(" packets: ")
			b.WriteString(strconv.Itoa(int(c.current)))

			// print
			clearLine()

			_, _ = os.Stdout.WriteString(b.String())
		}
	}
}

// updates the progress indicator and writes to stdout periodically.
func (c *Collector) printProgressInterval() chan struct{} {
	stop := make(chan struct{})

	// TODO: adjust progress refresh interval based on input file size?
	const interval = 5

	go func() {
		for {
			select {
			case <-stop:
				return
			case <-time.After(interval * time.Second):

				// must be locked, otherwise a race occurs when sending a SIGINT
				// and triggering wg.Wait() in another goroutine...
				c.statMutex.Lock()

				// dont print message when collector is about to shutdown
				if c.shutdown {
					c.statMutex.Unlock()
					return
				}
				c.statMutex.Unlock()

				var (
					curr = atomic.LoadInt64(&c.current)
					num  = atomic.LoadInt64(&c.numPackets)
					last = atomic.LoadInt64(&c.numPacketsLast)
					pps  = (curr - last) / interval
				)

				atomic.StoreInt64(&c.numPacketsLast, curr)

				if !c.config.Quiet { // print
					clearLine()

					_, _ = fmt.Fprintf(os.Stdout,
						c.progressString,
						utils.Progress(curr, num),
						// decoder.Flows.Size(), // TODO: fetch this info from stats?
						// decoder.Connections.Size(), // TODO: fetch this info from stats?
						decoder.Profiles.Size(),
						decoder.ServiceStore.Size(),
						int(curr),
						pps,
					)
				}
			}
		}
	}()

	return stop
}

// assemble the progress string once, to reduce recurring allocations.
func (c *Collector) buildProgressString() {
	c.progressString = "decoding packets... (%s) profiles: %d services: %d total packets: %d pkts/sec %d"
}

// GetNumPackets returns the current number of processed packets.
func (c *Collector) GetNumPackets() int64 {
	return atomic.LoadInt64(&c.current)
}

// FreeOSMemory forces freeing memory.
func (c *Collector) freeOSMemory() {
	for range time.After(time.Duration(c.config.FreeOSMem) * time.Minute) {
		debug.FreeOSMemory()
	}
}

// PrintConfiguration dumps the current collector config to stdout.
func (c *Collector) PrintConfiguration() {
	// ensure the logfile handle gets opened
	err := c.initLogging()
	if err != nil {
		log.Fatal("failed to open logfile:", err)
	}

	var target io.Writer
	if c.config.Quiet {
		target = logFileHandle
	} else {
		target = io.MultiWriter(os.Stdout, logFileHandle)
	}

	cdata, err := json.MarshalIndent(c.config, " ", "  ")
	if err != nil {
		log.Fatal(err)
	}
	// always write the entire configuration into the logfile
	_, _ = logFileHandle.Write(cdata)

	netcap.FPrintLogo(target)

	if c.config.DecoderConfig.Debug {
		// in debug mode: dump config to stdout
		target = io.MultiWriter(os.Stdout, logFileHandle)
	} else {
		// default: write configuration into netcap.log
		target = logFileHandle
		fmt.Println() // add newline
	}

	netcap.FPrintBuildInfo(target)

	// print build information
	_, _ = fmt.Fprintln(target, "> PID:", os.Getpid())

	// print configuration as table
	tui.Table(target, []string{"Setting", "Value"}, [][]string{
		{"Workers", strconv.Itoa(c.config.Workers)},
		{"MemBuffer", strconv.FormatBool(c.config.DecoderConfig.Buffer)},
		{"MemBufferSize", strconv.Itoa(c.config.DecoderConfig.MemBufferSize) + " bytes"},
		{"Compression", strconv.FormatBool(c.config.DecoderConfig.Compression)},
		{"PacketBuffer", strconv.Itoa(c.config.PacketBufferSize) + " packets"},
		{"PacketContext", strconv.FormatBool(c.config.DecoderConfig.AddContext)},
		{"Payloads", strconv.FormatBool(c.config.DecoderConfig.IncludePayloads)},
		{"FileStorage", c.config.DecoderConfig.FileStorage},
	})
	_, _ = fmt.Fprintln(target) // add a newline
}

// initLogging can be used to open the logfile before calling Init()
// this is used to be able to dump the collector configuration into the netcap.log in quiet mode
// following calls to Init() will not open the filehandle again.
func (c *Collector) initLogging() error {
	// prevent reopen
	if logFileHandle != nil {
		return nil
	}

	if c.config.DecoderConfig.Out != "" {
		if stat, err := os.Stat(c.config.DecoderConfig.Out); err != nil {

			err = os.MkdirAll(c.config.DecoderConfig.Out, os.FileMode(outDirPermissionDefault))
			if err != nil {
				fmt.Println(err)
			}

			_, err = os.Stat(c.config.DecoderConfig.Out)
			if err != nil {
				return err
			}
		} else if !stat.IsDir() {
			return errInvalidOutputDirectory
		}
	}

	var err error

	logFileHandle, err = os.OpenFile(filepath.Join(c.config.DecoderConfig.Out, "netcap.log"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, c.config.OutDirPermission)
	if err != nil {
		return err
	}

	return nil
}

// Stop will halt packet collection and wait for all processing to finish.
func (c *Collector) Stop() {
	c.cleanup(false)
}

// ForceStop will halt packet collection immediately without waiting for processing to finish.
func (c *Collector) forceStop() {
	c.cleanup(true)
}
