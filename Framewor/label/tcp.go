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

package label

import (
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gogo/protobuf/proto"
	"gopkg.in/cheggaaa/pb.v1"

	"github.com/dreadl0ck/netcap"
	"github.com/dreadl0ck/netcap/types"
	"github.com/dreadl0ck/netcap/utils"
)

// labelTCP labels type NC_TCP.
func labelTCP(wg *sync.WaitGroup, file string, alerts []*suricataAlert, outDir, separator, selection string) *pb.ProgressBar {
	var (
		fname           = filepath.Join(outDir, "TCP.ncap.gz")
		total, errCount = netcap.Count(fname)
		labelsTotal     = 0
		progress        = pb.New(int(total)).Prefix(utils.Pad(utils.TrimFileExtension(file), 25))
		outFileName     = filepath.Join(outDir, "TCP_labeled.csv")
	)
	if errCount != nil {
		log.Fatal("failed to count audit records:", errCount)
	}

	go func() {
		r, err := netcap.Open(fname, netcap.DefaultBufferSize)
		if err != nil {
			panic(err)
		}

		// read netcap header
		header, errFileHeader := r.ReadHeader()
		if errFileHeader != nil {
			log.Fatal(errFileHeader)
		}
		if header.Type != types.Type_NC_TCP {
			panic("file does not contain TCP records: " + header.Type.String())
		}

		// outfile handle
		f, err := os.Create(outFileName)
		if err != nil {
			panic(err)
		}

		var (
			tcp = new(types.TCP)
			fl  types.AuditRecord
			pm  proto.Message
			ok  bool
		)
		pm = tcp

		types.Select(tcp, selection)

		if fl, ok = pm.(types.AuditRecord); !ok {
			panic("type does not implement types.AuditRecord interface")
		}

		// write header
		_, err = f.WriteString(strings.Join(fl.CSVHeader(), separator) + separator + "result" + "\n")
		if err != nil {
			panic(err)
		}

	read:
		for {
			err = r.Next(tcp)
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				break
			} else if err != nil {
				panic(err)
			}

			if UseProgressBars {
				progress.Increment()
			}

			var finalLabel string

			// Unidirectional TCP packets
			// checks if packet has a source or destination port matching an alert
			for _, a := range alerts {
				// must be a TCP packet
				if a.Proto == "TCP" &&

					// AND timestamp must match
					a.Timestamp == tcp.Timestamp &&

					// AND destination port must match
					a.DstPort == int(tcp.DstPort) &&

					// AND source port must match
					a.SrcPort == int(tcp.SrcPort) {
					if CollectLabels {
						// only if it is not already part of the label
						if !strings.Contains(finalLabel, a.Classification) {
							if finalLabel == "" {
								finalLabel = a.Classification
							} else {
								finalLabel += " | " + a.Classification
							}
						}

						continue
					}

					// add label
					_, _ = f.WriteString(strings.Join(tcp.CSVRecord(), separator) + separator + a.Classification + "\n")
					labelsTotal++

					goto read
				}
			}

			if len(finalLabel) != 0 {

				if strings.HasPrefix(finalLabel, " |") {
					log.Fatal("invalid label: ", finalLabel)
				}

				// add final label
				_, _ = f.WriteString(strings.Join(tcp.CSVRecord(), separator) + separator + finalLabel + "\n")
				labelsTotal++

				goto read
			}

			// label as normal
			_, _ = f.WriteString(strings.Join(tcp.CSVRecord(), separator) + separator + "normal\n")
		}
		finish(wg, r, f, labelsTotal, outFileName, progress)
	}()

	return progress
}
