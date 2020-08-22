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

package util

import (
	"os"

	"github.com/namsral/flag"

	"github.com/dreadl0ck/netcap"
)

// Flags returns all flags.
func Flags() (flags []string) {
	fs.VisitAll(func(f *flag.Flag) {
		flags = append(flags, f.Name)
	})

	return
}

var (
	// util.
	fs                 = flag.NewFlagSetWithEnvPrefix(os.Args[0], "NC", flag.ExitOnError)
	flagGenerateConfig = fs.Bool("gen-config", false, "generate config")
	_                  = fs.String("config", "", "read configuration from file at path")
	flagCheckFields    = fs.Bool("check", false, "check number of occurrences of the separator, in fields of an audit record file")
	flagToUTC          = fs.String("ts2utc", "", "util to convert seconds.microseconds timestamp to UTC")
	flagInput          = fs.String("read", "", "read specified audit record file")
	flagSeparator      = fs.String("sep", ",", "set separator string for csv output")
	flagVersion        = fs.Bool("version", false, "print netcap package version and exit")
	flagMemBufferSize  = fs.Int("membuf-size", netcap.DefaultBufferSize, "set size for membuf")
	flagEnv            = fs.Bool("env", false, "print netcap environment variables and exit")
	flagInterfaces     = fs.Bool("interfaces", false, "print netcap environment variables and exit")
	flagIndex          = fs.String("index", "", "index data for full text search")
)
