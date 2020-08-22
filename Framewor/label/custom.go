/*
 * NETCAP - Traffic Analysis Framework
 * Copyright (c) 2019 Philipp Mieden <dreadl0ck [at] protonmail [dot] ch>
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
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	gzip "github.com/klauspost/pgzip"

	"github.com/dreadl0ck/netcap"
	"github.com/dreadl0ck/netcap/types"
	"github.com/dreadl0ck/netcap/utils"
	"github.com/evilsocket/islazy/tui"
	pb "gopkg.in/cheggaaa/pb.v1"
)

//type Custom struct {
//	AttackNumber   int
//	StartTime      int64
//	EndTime        int64
//	AttackDuration time.Duration
//	AttackPoints   []string
//	Adresses       []string
//	AttackName     string
//	AttackType     string
//	Intent         string
//	ActualChange   string
//	Notes          string
//}

type AttackInfo struct {
	Num      int
	Name     string
	Start    time.Time
	End      time.Time
	IPs      []string
	Proto    string
	Notes    string
	Category string
}

func ParseAttackInfos(path string) (labelMap map[string]*AttackInfo, labels []*AttackInfo) {

	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	// alerts that have a duplicate timestamp
	var duplicates = []*AttackInfo{}

	// ts:alert
	labelMap = make(map[string]*AttackInfo)

	for _, record := range records[1:] {

		num, err := strconv.Atoi(record[0])
		if err != nil {
			log.Fatal(err)
		}

		start, err := time.Parse("2006/1/2 15:04:05", record[2])
		if err != nil {
			log.Fatal(err)
		}

		end, err := time.Parse("2006/1/2 15:04:05", record[3])
		if err != nil {
			log.Fatal(err)
		}

		//duration, err := time.ParseDuration(record[4])
		//if err != nil {
		//	log.Fatal(err)
		//}

		toArr := func(input string) []string {
			return strings.Split(strings.Trim(input, "\""), ";")
		}

		custom := &AttackInfo{
			Num:      num,              // int
			Start:    start,            // time.Time
			End:      end,              // time.Time
			IPs:      toArr(record[4]), // []string
			Name:     record[1],        // string
			Proto:    record[5],        // string
			Notes:    record[6],        // string
			Category: record[7],        // string
		}

		// ensure no alerts with empty name are collected
		if custom.Name == "" || custom.Name == " " {
			fmt.Println("skipping entry with empty name", custom)
			continue
		}

		// count total occurrences of classification
		ClassificationMap[custom.Name]++

		// check if excluded
		if !excluded[custom.Name] {

			// append to collected alerts
			labels = append(labels, custom)

			startTsString := strconv.FormatInt(custom.Start.Unix(), 10)

			// add to label map
			if _, ok := labelMap[startTsString]; ok {
				// an alert for this timestamp already exists
				// if configured the execution will stop
				// for now the first seen alert for a timestamp will be kept
				duplicates = append(duplicates, custom)
			} else {
				labelMap[startTsString] = custom
			}
		}
	}

	return
}

// CustomLabels uses info from a csv file to label the data
func CustomLabels(pathMappingInfo, outputPath string, useDescription bool, separator, selection string) error {

	var (
		start            = time.Now()
		labelMap, labels = ParseAttackInfos(pathMappingInfo)
	)
	if len(labels) == 0 {
		fmt.Println("no labels found.")
		os.Exit(0)
	}

	fmt.Println("got", len(labels), "labels")

	rows := [][]string{}
	for i, c := range labels {
		rows = append(rows, []string{strconv.Itoa(i + 1), c.Name})
	}

	// print alert summary
	tui.Table(os.Stdout, []string{"Num", "AttackName"}, rows)
	fmt.Println()

	// apply labels to data
	// set outDir to current dir or flagOut
	var outDir string
	if outputPath != "" {
		outDir = outputPath
	} else {
		outDir = "."
	}

	// label all layer data in outDir
	// first read directory
	files, err := ioutil.ReadDir(outDir)
	if err != nil {
		return err
	}

	var (
		wg  sync.WaitGroup
		pbs []*pb.ProgressBar
	)

	// iterate over all files in dir
	for _, f := range files {

		// check if its an audit record file
		if strings.HasSuffix(f.Name(), ".ncap.gz") || strings.HasSuffix(f.Name(), ".ncap") {
			wg.Add(1)

			var (
				// get record name
				filename = f.Name()
				typ      = strings.TrimSuffix(strings.TrimSuffix(filename, ".ncap.gz"), ".ncap")
			)

			//fmt.Println("type", typ)
			pbs = append(pbs, CustomMap(&wg, filename, typ, labelMap, labels, outputPath, separator, selection))
		}
	}

	var pool *pb.Pool
	if UseProgressBars {

		// wait for goroutines to start and initialize
		// otherwise progress bars will bug
		time.Sleep(3 * time.Second)

		// start pool
		pool, err = pb.StartPool(pbs...)
		if err != nil {
			return err
		}
		utils.ClearScreen()
	}

	wg.Wait()

	if UseProgressBars {
		// close pool
		if err := pool.Stop(); err != nil {
			fmt.Println("failed to stop progress bar pool:", err)
		}
	}

	fmt.Println("\ndone in", time.Since(start))
	return nil
}

// CustomMap uses info from a csv file to label the data
//func CustomMap(wg *sync.WaitGroup, file string, typ string, labelMap map[string]*SuricataAlert, labels []*SuricataAlert, outDir, separator, selection string) *pb.ProgressBar {
func CustomMap(wg *sync.WaitGroup, file string, typ string, labelMap map[string]*AttackInfo, labels []*AttackInfo, outDir, separator, selection string) *pb.ProgressBar {

	var (
		fname       = filepath.Join(outDir, file)
		total       = netcap.Count(fname)
		labelsTotal = 0
		outFileName = filepath.Join(outDir, typ+"_labeled.csv.gz")
		progress    = pb.New(int(total)).Prefix(utils.Pad(utils.TrimFileExtension(file), 25))
	)

	go func() {

		// open layer data file
		r, err := netcap.Open(fname, netcap.DefaultBufferSize)
		if err != nil {
			panic(err)
		}

		// read netcap header
		header := r.ReadHeader()

		// create outfile handle
		f, err := os.Create(outFileName)
		if err != nil {
			panic(err)
		}

		gzipWriter := gzip.NewWriter(f)

		var (
			record = netcap.InitRecord(header.Type)
			ok     bool
			p      types.AuditRecord
		)

		// check if we can decode it as CSV
		if p, ok = record.(types.AuditRecord); !ok {
			panic("type does not implement types.AuditRecord interface:" + typ)
		}

		// run selection
		types.Select(record, selection)

		// write header
		_, err = gzipWriter.Write([]byte(strings.Join(p.CSVHeader(), separator) + separator + "result" + "\n"))
		if err != nil {
			panic(err)
		}

		for {
			err := r.Next(record)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			} else if err != nil {
				panic(err)
			}

			if UseProgressBars {
				progress.Increment()
			}

			var label string

			// check if flow has a source or destination adress matching an alert
			// if not label it as normal
			for _, l := range labels {

				var numMatches int

				// check if any of the addresses from the labeling info
				// is either source or destination of the current audit record
				for _, addr := range l.IPs {
					if p.Src() == addr || p.Dst() == addr {
						numMatches++
					}
				}
				if numMatches != 2 {
					// label as normal
					gzipWriter.Write([]byte(strings.Join(p.CSVRecord(), separator) + separator + "normal\n"))
					continue
				}

				// verify time interval of audit record is within the attack period
				auditRecordTime := utils.StringToTime(p.Time()).UTC().Add(8 * time.Hour)

				// if the audit record has a timestamp in the attack period
				if (l.Start.Before(auditRecordTime) && l.End.After(auditRecordTime)) ||

					// or matches exactly the one on the audit record
					l.Start.Equal(auditRecordTime) || l.End.Equal(auditRecordTime) {

					if Debug {
						fmt.Println("-----------------------", typ, l.Name, l.Category)
						fmt.Println("flow:", p.Src(), "->", p.Dst(), "addr:", "attack ips:", l.IPs)
						fmt.Println("start", l.Start)
						fmt.Println("end", l.End)
						fmt.Println("auditRecordTime", auditRecordTime)
						fmt.Println("(l.Start.Before(auditRecordTime) && l.End.After(auditRecordTime))", l.Start.Before(auditRecordTime) && l.End.After(auditRecordTime))
						fmt.Println("l.Start.Equal(auditRecordTime)", l.Start.Equal(auditRecordTime))
						fmt.Println("l.End.Equal(auditRecordTime))", l.End.Equal(auditRecordTime))
					}

					// only if it is not already part of the label
					if !strings.Contains(label, l.Category) {
						if label == "" {
							label = l.Category
						} else {
							label += " | " + l.Category
						}
					}
				}
			}

			if len(label) != 0 {
				if strings.HasPrefix(label, " |") {
					log.Fatal("invalid label: ", label)
				}

				// add label
				gzipWriter.Write([]byte(strings.Join(p.CSVRecord(), separator) + separator + label + "\n"))
				labelsTotal++
			} else {
				// label as normal
				gzipWriter.Write([]byte(strings.Join(p.CSVRecord(), separator) + separator + "normal\n"))
			}
		}
		err = gzipWriter.Flush()
		if err != nil {
			log.Fatal(err)
		}
		err = gzipWriter.Close()
		if err != nil {
			log.Fatal(err)
		}
		finish(wg, r, f, labelsTotal, outFileName, progress)
	}()

	return progress
}
