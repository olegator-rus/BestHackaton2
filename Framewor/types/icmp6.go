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

package types

import (
	"strings"

	"github.com/dreadl0ck/netcap/utils"
	"github.com/prometheus/client_golang/prometheus"
)

var fieldsICMPv6 = []string{
	"Timestamp",
	"TypeCode", // int32
	"Checksum", // int32
	"SrcIP",
	"DstIP",
}

// CSVHeader returns the CSV header for the audit record.
func (i *ICMPv6) CSVHeader() []string {
	return filter(fieldsICMPv6)
}

// CSVRecord returns the CSV record for the audit record.
func (i *ICMPv6) CSVRecord() []string {
	// prevent accessing nil pointer
	if i.Context == nil {
		i.Context = &PacketContext{}
	}
	return filter([]string{
		formatTimestamp(i.Timestamp),
		formatInt32(i.TypeCode),
		formatInt32(i.Checksum),
		i.Context.SrcIP,
		i.Context.DstIP,
	})
}

// Time returns the timestamp associated with the audit record.
func (i *ICMPv6) Time() string {
	return i.Timestamp
}

// JSON returns the JSON representation of the audit record.
func (i *ICMPv6) JSON() (string, error) {
	i.Timestamp = utils.TimeToUnixMilli(i.Timestamp)
	return jsonMarshaler.MarshalToString(i)
}

var icmp6Metric = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: strings.ToLower(Type_NC_ICMPv6.String()),
		Help: Type_NC_ICMPv6.String() + " audit records",
	},
	fieldsICMPv6[1:],
)

// Inc increments the metrics for the audit record.
func (i *ICMPv6) Inc() {
	icmp6Metric.WithLabelValues(i.CSVRecord()[1:]...).Inc()
}

// SetPacketContext sets the associated packet context for the audit record.
func (i *ICMPv6) SetPacketContext(ctx *PacketContext) {
	i.Context = ctx
}

// Src returns the source address of the audit record.
func (i *ICMPv6) Src() string {
	if i.Context != nil {
		return i.Context.SrcIP
	}
	return ""
}

// Dst returns the destination address of the audit record.
func (i *ICMPv6) Dst() string {
	if i.Context != nil {
		return i.Context.DstIP
	}
	return ""
}
