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

import "flag"

var (
	flagDebug     = flag.Bool("debug", false, "toggle debug mode")
	flagInput     = flag.String("r", "", "(required) read specified file, can either be a pcap or netcap audit record file")
	flagSeparator = flag.String("sep", ",", "set separator string for csv output")
	flagOutDir    = flag.String("out", "", "specify output directory, will be created if it does not exist")

	flagDescription           = flag.Bool("description", false, "use attack description instead of classification for labels")
	flagProgressBars          = flag.Bool("progress", false, "use progress bars")
	flagStopOnDuplicateLabels = flag.Bool("strict", false, "fail when there is more than one alert for the same timestamp")
	flagExcludeLabels         = flag.String("exclude", "", "specify a comma separated list of suricata classifications that shall be excluded from the generated labeled csv")
	flagCollectLabels         = flag.Bool("collect", false, "append classifications from alert with duplicate timestamps to the generated label")
	flagDisableLayerMapping   = flag.Bool("disable-layers", false, "do not map layer types by timestamp")
	flagSuricataConfigPath    = flag.String("suricata-config", "/usr/local/etc/suricata/suricata.yaml", "set the path to the suricata config file")

	flagVersion = flag.Bool("version", false, "print netcap package version and exit")
	flagCustom  = flag.String("custom", "", "use custom mappings at path")

	// this wont work currently, because the Select() func will stop if there are fields that are not present on an audit record
	// as labeling iterates over all available records, there will always be a record that does not have all selected fields
	// TODO: create a func that ignores fields that do not exist on the target audit record, maybe Select() and SelectStrict()
	// flagSelect    = flag.String("select", "", "select specific fields of an audit records when generating csv or tables")
)
