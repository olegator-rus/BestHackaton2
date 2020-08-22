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
	"strconv"
	"strings"

	"github.com/dreadl0ck/netcap/utils"
	"github.com/prometheus/client_golang/prometheus"
)

var fieldsCIP = []string{
	"Timestamp",
	"Response",         // bool
	"ServiceID",        // int32
	"ClassID",          // uint32
	"InstanceID",       // uint32
	"Status",           // int32
	"AdditionalStatus", // []uint32
	"Data",             // []byte
	"SrcIP",
	"DstIP",
	"SrcPort",
	"DstPort",
}

// CSVHeader returns the CSV header for the audit record.
func (c *CIP) CSVHeader() []string {
	return filter(fieldsCIP)
}

// CSVRecord returns the CSV record for the audit record.
func (c *CIP) CSVRecord() []string {
	additional := make([]string, len(c.AdditionalStatus))

	if c.Response {
		for _, v := range c.AdditionalStatus {
			additional = append(additional, formatUint32(v))
		}
	}

	// prevent accessing nil pointer
	if c.Context == nil {
		c.Context = &PacketContext{}
	}

	return filter([]string{
		formatTimestamp(c.Timestamp),
		strconv.FormatBool(c.Response), // bool
		formatInt32(c.ServiceID),       // int32
		formatUint32(c.ClassID),        // uint32
		formatUint32(c.InstanceID),     // uint32
		formatInt32(c.Status),          // int32
		strings.Join(additional, ""),   // []uint32
		hex.EncodeToString(c.Data),     // []byte
		c.Context.SrcIP,
		c.Context.DstIP,
		c.Context.SrcPort,
		c.Context.DstPort,
	})
}

// Time returns the timestamp associated with the audit record.
func (c *CIP) Time() string {
	return c.Timestamp
}

// JSON returns the JSON representation of the audit record.
func (c *CIP) JSON() (string, error) {
	c.Timestamp = utils.TimeToUnixMilli(c.Timestamp)
	return jsonMarshaler.MarshalToString(c)
}

var cipMetric = prometheus.NewCounterVec( //nolint:gochecknoglobals
	prometheus.CounterOpts{
		Name: strings.ToLower(Type_NC_CIP.String()),
		Help: Type_NC_CIP.String() + " audit records",
	},
	fieldsCIP[1:],
)

// Inc increments the metrics for the audit record.
func (c *CIP) Inc() {
	cipMetric.WithLabelValues(c.CSVRecord()[1:]...).Inc()
}

// SetPacketContext sets the associated packet context for the audit record.
func (c *CIP) SetPacketContext(ctx *PacketContext) {
	c.Context = ctx
}

// Src returns the source address of the audit record.
func (c *CIP) Src() string {
	if c.Context != nil {
		return c.Context.SrcIP
	}

	return ""
}

// Dst returns the destination address of the audit record.
func (c *CIP) Dst() string {
	if c.Context != nil {
		return c.Context.DstIP
	}

	return ""
}
