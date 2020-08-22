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

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dreadl0ck/netcap"
	"github.com/dreadl0ck/netcap/utils"
)

func main() {

	// parse commandline flags
	flag.Usage = printUsage
	flag.Parse()

	// print version and exit
	if *flagVersion {
		fmt.Println(netcap.Version)
		os.Exit(0)
	}

	// util to convert netcap timestamp to UTC time
	if *flagToUTC != "" {
		fmt.Println(utils.TimeToUTC(*flagToUTC))
		os.Exit(1)
	}

	// util to check if fields count matches for all generated rows
	if *flagCheckFields {
		checkFields()
		return
	}
}
