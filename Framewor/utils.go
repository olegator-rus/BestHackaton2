/*
 * NETCAP - Traffic Analysis Framework
 * Copyright (c) 2017 Philipp Mieden <dreadl0ck [at] protonmail [dot] ch>
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package netcap

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	gzip "github.com/klauspost/pgzip"

	"github.com/dreadl0ck/netcap/types"
	"github.com/dreadl0ck/netcap/utils"
	"github.com/evilsocket/islazy/tui"
	"github.com/gogo/protobuf/proto"
)

var logo = `                       / |
 _______    ______   _10 |_     _______   ______    ______
/     / \  /    / \ / 01/  |   /     / | /    / \  /    / \
0010100 /|/011010 /|101010/   /0101010/  001010  |/100110  |
01 |  00 |00    00 |  10 | __ 00 |       /    10 |00 |  01 |
10 |  01 |01001010/   00 |/  |01 \_____ /0101000 |00 |__10/|
10 |  00 |00/    / |  10  00/ 00/    / |00    00 |00/   00/
00/   10/  0101000/    0010/   0010010/  0010100/ 1010100/
                                                  00 |
Network Protocol Analysis Framework               00 |
created by Philipp Mieden, 2018                   00/
` + Version

// PrintLogo prints the netcap logo
func PrintLogo() {
	utils.ClearScreen()
	fmt.Println(logo)
}

// PrintBuildInfo displays build information related to netcap
func PrintBuildInfo() {

	PrintLogo()

	fmt.Println("\n> Date:", time.Now().UTC())
	fmt.Println("> NETCAP build commit:", Commit)
	fmt.Println("> go runtime version:", runtime.Version())

	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println("> running with:", runtime.NumCPU(), "cores")

	b, ok := debug.ReadBuildInfo()
	if ok {
		for _, d := range b.Deps {
			if path.Base(d.Path) == "gopacket" {
				fmt.Println("> gopacket:", d.Path, "version:", d.Version)
			}
		}
	}
}

// DumpConfig contains all possible settings for dumping an audit records
type DumpConfig struct {
	Path          string
	Separator     string
	TabSeparated  bool
	Structured    bool
	Table         bool
	Selection     string
	UTC           bool
	Fields        bool
	JSON          bool
	MemBufferSize int
}

// Dump reads the specified netcap file
// and dumps the output according to the configuration to stdout
func Dump(c DumpConfig) {

	var (
		count  = 0
		r, err = Open(c.Path, c.MemBufferSize)
	)
	if err != nil {
		log.Fatal("failed to open audit record file: ", err)
	}
	defer r.Close()

	if c.Separator == "\\t" || c.TabSeparated {
		c.Separator = "\t"
	}

	var (
		header = r.ReadHeader()
		record = InitRecord(header.Type)
		// rows for table print
		rows [][]string
	)

	types.Select(record, c.Selection)
	types.UTC = c.UTC

	if !c.Structured && !c.Table {

		if p, ok := record.(types.AuditRecord); ok {
			fmt.Println(strings.Join(p.CSVHeader(), c.Separator))
		} else {
			fmt.Printf("type: %#v\n", record)
			log.Fatal("type does not implement the types.AuditRecord interface")
		}

		if c.Fields {
			os.Exit(0)
		}
	}

	for {
		err := r.Next(record)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		} else if err != nil {
			panic(err)
		}
		count++

		if c.Structured {
			os.Stdout.WriteString(header.Type.String())
			os.Stdout.WriteString("\n")
			os.Stdout.WriteString(proto.MarshalTextString(record))
			os.Stdout.WriteString("\n")
			continue
		}

		if p, ok := record.(types.AuditRecord); ok {
			if c.JSON {
				marshaled, err := json.MarshalIndent(p, "  ", " ")
				if err != nil {
					log.Fatal("failed to marshal json:", err)
				}
				os.Stdout.WriteString(string(marshaled))
				os.Stdout.WriteString("\n")
				continue
			}
			if c.Table {
				rows = append(rows, p.CSVRecord())

				if count%100 == 0 {
					tui.Table(os.Stdout, p.CSVHeader(), rows)
					rows = [][]string{}
				}
				continue
			}
			os.Stdout.WriteString(strings.Join(p.CSVRecord(), c.Separator) + "\n")
		} else {
			fmt.Printf("type: %#v\n", record)
			log.Fatal("type does not implement the types.AuditRecord interface")
		}

	}

	if c.Table {
		if p, ok := record.(types.AuditRecord); ok {
			tui.Table(os.Stdout, p.CSVHeader(), rows)
			fmt.Println()
		} else {
			fmt.Printf("type: %#v\n", record)
			log.Fatal("type does not implement the types.AuditRecord interface")
		}
	}

	fmt.Println(count, "records.")
}

// CloseFile closes the netcap file handle
// and removes files that do only contain a header but no audit records
func CloseFile(outDir string, file *os.File, typ string) (name string, size int64) {

	i, err := file.Stat()
	if err != nil {
		fmt.Println("[ERROR] closing file:", err, "type", typ)
		return "", 0
	}

	var (
		errSync  = file.Sync()
		errClose = file.Close()
	)
	if errSync != nil || errClose != nil {
		fmt.Println("error while closing", i.Name(), "errSync", errSync, "errClose", errClose)
	}

	return i.Name(), RemoveAuditRecordFileIfEmpty(filepath.Join(outDir, i.Name()))
}

// CreateFile is a wrapper to create new audit record file
func CreateFile(name, ext string) *os.File {
	f, err := os.Create(name + ext)
	if err != nil {
		panic(err)
	}
	return f
}

// RemoveAuditRecordFileIfEmpty removes the audit record file if it does not contain audit records
func RemoveAuditRecordFileIfEmpty(name string) (size int64) {

	if strings.HasSuffix(name, ".csv") || strings.HasSuffix(name, ".csv.gz") {
		f, err := os.Open(name)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		var r *bufio.Reader
		if strings.HasSuffix(name, ".csv.gz") {
			gr, err := gzip.NewReader(f)
			if err != nil {
				panic(err)
			}
			r = bufio.NewReader(gr)
		} else {
			r = bufio.NewReader(f)
		}

		count := 0
		for {
			_, _, err := r.ReadLine()
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			} else if err != nil {
				panic(err)
			}
			count++
			if count > 1 {
				break
			}
		}

		if count < 2 {
			// remove file
			err = os.Remove(name)
			if err != nil {
				fmt.Println("failed to remove file", err)
			}

			// return file size of zero
			return 0
		}

		// dont remove file
		// return final file size
		s, err := os.Stat(name)
		if err != nil {
			fmt.Println("failed to stat file:", name, err)
			return
		}
		return s.Size()
	}

	// Check if audit record file contains records
	// Open, read header and the first audit record and return
	r, err := Open(name, DefaultBufferSize)
	if err != nil {

		// suppress errors for OSPF because the file handle will be closed twice
		// since both v2 and v3 have the same gopacket.LayerType == "OSPF"
		if !strings.HasPrefix(name, "OSPF") {
			fmt.Println("unable to open file:", name, "error", err)
		}
		return 0
	}
	defer r.Close()

	var (
		header = r.ReadHeader()
		record = InitRecord(header.Type)
	)

	err = r.Next(record)
	if err != nil {
		// remove file
		err = os.Remove(name)
		if err != nil {
			fmt.Println("failed to remove file", err)

			// return file size of zero
			return 0
		}
		return
	}

	// dont remove file, it contains audit records
	// return final file size
	s, err := os.Stat(name)
	if err != nil {
		fmt.Println("failed to stat file:", name, err)
		return
	}
	return s.Size()
}

// NewHeader creates and returns a new netcap audit file header
func NewHeader(t types.Type, source, version string, includesPayloads bool) *types.Header {

	// init header
	header := new(types.Header)
	header.Type = t
	header.Created = utils.TimeToString(time.Now())
	header.InputSource = source
	header.Version = version
	header.ContainsPayloads = includesPayloads

	return header
}
