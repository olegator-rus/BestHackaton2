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

var fieldsCredentials = []string{
	"Timestamp",
}

// CSVHeader returns the CSV header for the audit record.
func (c *Credentials) CSVHeader() []string {
	return filter(fieldsCredentials)
}

// CSVRecord returns the CSV record for the audit record.
func (c *Credentials) CSVRecord() []string {
	return filter([]string{
		formatTimestamp(c.Timestamp),
	})
}

// Time returns the timestamp associated with the audit record.
func (c *Credentials) Time() string {
	return c.Timestamp
}

// JSON returns the JSON representation of the audit record.
func (c *Credentials) JSON() (string, error) {
	c.Timestamp = utils.TimeToUnixMilli(c.Timestamp)
	return jsonMarshaler.MarshalToString(c)
}

var credentialsMetric = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: strings.ToLower(Type_NC_Credentials.String()),
		Help: Type_NC_Credentials.String() + " audit records",
	},
	fieldsCredentials[1:],
)

// Inc increments the metrics for the audit record.
func (c *Credentials) Inc() {
	credentialsMetric.WithLabelValues(c.CSVRecord()[1:]...).Inc()
}

// SetPacketContext sets the associated packet context for the audit record.
func (a *Credentials) SetPacketContext(*PacketContext) {}

// Src TODO: preserve source and destination mac adresses for Credentials and return them here.
// Src returns the source address of the audit record.
func (c *Credentials) Src() string {
	return ""
}

// Dst TODO: preserve source and destination mac adresses for Credentials and return them here.
// Dst returns the destination address of the audit record.
func (c *Credentials) Dst() string {
	return ""
}
