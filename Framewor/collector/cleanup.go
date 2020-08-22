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

package collector

import (
	"fmt"
	"log"
	"time"

	"github.com/dustin/go-humanize"

	"github.com/dreadl0ck/netcap"
	"github.com/dreadl0ck/netcap/decoder"
	"github.com/dreadl0ck/netcap/resolvers"
	"github.com/dreadl0ck/netcap/utils"
)

// cleanup before leaving. closes all buffers and displays stats.
func (c *Collector) cleanup(force bool) {
	c.statMutex.Lock()
	c.shutdown = true
	c.statMutex.Unlock()

	// stop all workers.
	// this will block until all workers are stopped
	// all packets left in the packet queues will be processed
	c.stopWorkers()

	waitForCollector := func() chan struct{} {
		ch := make(chan struct{})

		go func() {
			c.statMutex.Lock()
			c.wg.Wait()
			c.statMutex.Unlock()

			ch <- struct{}{}
		}()

		return ch
	}

	c.printStdOut("\nwaiting for main collector wait group...")
	select {
	case <-waitForCollector():
		c.printlnStdOut(" done!")
	case <-time.After(netcap.DefaultReassemblyTimeout):
		c.printStdOut(" timeout after ", netcap.DefaultReassemblyTimeout)
	}

	if c.config.ReassembleConnections {
		// teardown the TCP stream reassembly and print stats
		decoder.CleanupReassembly(!force, c.assemblers)
	}

	// flush all gopacket decoders
	for _, decoders := range c.goPacketDecoders {
		for _, e := range decoders {
			name, size := e.Destroy()
			if size != 0 {
				c.totalBytesWritten += size
				c.files[name] = humanize.Bytes(uint64(size))
			}
		}
	}

	// flush all custom decoders
	for _, d := range c.customDecoders {
		name, size := d.Destroy()
		if size != 0 {
			c.totalBytesWritten += size
			c.files[name] = humanize.Bytes(uint64(size))
		}
	}

	resolvers.SaveFingerprintDB()

	// close the encoder logs
	for _, e := range utils.CloseLogFiles() {
		fmt.Println("failed to close logfile handle:", e)
	}

	c.mu.Lock()
	if c.isLive {
		c.statMutex.Lock()
		c.numPackets = c.current
		c.statMutex.Unlock()
	}
	c.mu.Unlock()

	// sync pcap file
	if err := c.closePcapFiles(); err != nil {
		log.Fatal("failed to close pcap files: ", err)
	}

	c.closeErrorLogFile()
	c.stats()

	if c.config.DecoderConfig.Debug {
		c.printErrors()
	}

	c.printlnStdOut("execution time", time.Since(c.start))

	if logFileHandle != nil {
		err := logFileHandle.Close()
		if err != nil {
			c.printStdOut("failed to close logfile:", err)
		}
	}
}
