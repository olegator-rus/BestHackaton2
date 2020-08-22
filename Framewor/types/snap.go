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

var fieldsSNAP = []string{
	"Timestamp",
	"OrganizationalCode",
	"Type",
}

// CSVHeader returns the CSV header for the audit record.
func (s *SNAP) CSVHeader() []string {
	return filter(fieldsSNAP)
}

// CSVRecord returns the CSV record for the audit record.
func (s *SNAP) CSVRecord() []string {
	return filter([]string{
		formatTimestamp(s.Timestamp),
		hex.EncodeToString(s.OrganizationalCode),
		formatInt32(s.Type),
	})
}

// Time returns the timestamp associated with the audit record.
func (s *SNAP) Time() string {
	return s.Timestamp
}

// JSON returns the JSON representation of the audit record.
func (u *SNAP) JSON() (string, error) {
	u.Timestamp = utils.TimeToUnixMilli(u.Timestamp)
	return jsonMarshaler.MarshalToString(u)
}

var snapMetric = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: strings.ToLower(Type_NC_SNAP.String()),
		Help: Type_NC_SNAP.String() + " audit records",
	},
	fieldsSNAP[1:],
)

// Inc increments the metrics for the audit record.
func (s *SNAP) Inc() {
	snapMetric.WithLabelValues(s.CSVRecord()[1:]...).Inc()
}

// SetPacketContext sets the associated packet context for the audit record.
func (s *SNAP) SetPacketContext(*PacketContext) {}

// Src TODO.
// Src returns the source address of the audit record.
func (s *SNAP) Src() string {
	return ""
}

// Dst returns the destination address of the audit record.
func (s *SNAP) Dst() string {
	return ""
}
