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

package types

import (
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

var fieldsNetworkFlow = []string{
	"TimestampFirst",
	"TimestampLast",
	"Proto",
	"SrcIP",
	"DstIP",
	"TotalSize",
	"NumPackets",
	"UID",
	"Duration",
}

func (f NetworkFlow) CSVHeader() []string {
	return filter(fieldsNetworkFlow)
}

func (f NetworkFlow) CSVRecord() []string {
	return filter([]string{
		formatTimestamp(f.TimestampFirst),
		formatTimestamp(f.TimestampLast),
		f.Proto,
		f.SrcIP,
		f.DstIP,
		formatInt64(f.TotalSize),
		formatInt64(f.NumPackets),
		strconv.FormatUint(f.UID, 10),
		formatInt64(f.Duration),
	})
}

func (f NetworkFlow) Time() string {
	return f.TimestampFirst
}

func (u NetworkFlow) JSON() (string, error) {
	return jsonMarshaler.MarshalToString(&u)
}

var networkFlowMetric = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: strings.ToLower(Type_NC_NetworkFlow.String()),
		Help: Type_NC_NetworkFlow.String() + " audit records",
	},
	fieldsNetworkFlow[1:],
)

func init() {
	prometheus.MustRegister(networkFlowMetric)
}

func (a NetworkFlow) Inc() {
	networkFlowMetric.WithLabelValues(a.CSVRecord()[1:]...).Inc()
}

func (a *NetworkFlow) SetPacketContext(ctx *PacketContext) {}

func (a NetworkFlow) Src() string {
	return a.SrcIP
}

func (a NetworkFlow) Dst() string {
	return a.DstIP
}
