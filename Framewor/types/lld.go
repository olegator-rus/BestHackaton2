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
	"encoding/hex"
	"strings"

	"github.com/dreadl0ck/netcap/utils"
	"github.com/prometheus/client_golang/prometheus"
)

var fieldsLLD = []string{
	"Timestamp",
	"ChassisID", // *LLDPChassisID
	"PortID",    // *LLDPPortID
	"TTL",       // int32
	"Values",    // []*LinkLayerDiscoveryValue
}

// CSVHeader returns the CSV header for the audit record.
func (l *LinkLayerDiscovery) CSVHeader() []string {
	return filter(fieldsLLD)
}

// CSVRecord returns the CSV record for the audit record.
func (l *LinkLayerDiscovery) CSVRecord() []string {
	values := make([]string, len(l.Values))
	for i, v := range l.Values {
		values[i] = v.toString()
	}
	return filter([]string{
		formatTimestamp(l.Timestamp),
		l.ChassisID.toString(), // *LLDPChassisID
		l.PortID.toString(),    // *LLDPPortID
		formatInt32(l.TTL),     // int32
		join(values...),        // []*LinkLayerDiscoveryValue
	})
}

// Time returns the timestamp associated with the audit record.
func (l *LinkLayerDiscovery) Time() string {
	return l.Timestamp
}

func (l LLDPChassisID) toString() string {
	return join(formatInt32(l.Subtype), hex.EncodeToString(l.ID))
}

func (l LLDPPortID) toString() string {
	return join(formatInt32(l.Subtype), hex.EncodeToString(l.ID))
}

// JSON returns the JSON representation of the audit record.
func (l *LinkLayerDiscovery) JSON() (string, error) {
	l.Timestamp = utils.TimeToUnixMilli(l.Timestamp)
	return jsonMarshaler.MarshalToString(l)
}

var lldMetric = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: strings.ToLower(Type_NC_LinkLayerDiscovery.String()),
		Help: Type_NC_LinkLayerDiscovery.String() + " audit records",
	},
	fieldsLLD[1:],
)

// Inc increments the metrics for the audit record.
func (l *LinkLayerDiscovery) Inc() {
	lldMetric.WithLabelValues(l.CSVRecord()[1:]...).Inc()
}

// SetPacketContext sets the associated packet context for the audit record.
func (l *LinkLayerDiscovery) SetPacketContext(*PacketContext) {}

// Src TODO.
// Src returns the source address of the audit record.
func (l *LinkLayerDiscovery) Src() string {
	return ""
}

// Dst returns the destination address of the audit record.
func (l *LinkLayerDiscovery) Dst() string {
	return ""
}
