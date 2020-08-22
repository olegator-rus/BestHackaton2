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

var fieldsBFD = []string{
	"Timestamp",
	"Version",
	"Diagnostic",
	"State",
	"Poll",
	"Final",
	"ControlPlaneIndependent",
	"AuthPresent",
	"Demand",
	"Multipoint",
	"DetectMultiplier",
	"MyDiscriminator",
	"YourDiscriminator",
	"DesiredMinTxInterval",
	"RequiredMinRxInterval",
	"RequiredMinEchoRxInterval",
	"AuthHeader",
	"SrcIP",
	"DstIP",
	"SrcPort",
	"DstPort",
}

// CSVHeader returns the CSV header for the audit record.
func (b *BFD) CSVHeader() []string {
	return filter(fieldsBFD)
}

// CSVRecord returns the CSV record for the audit record.
func (b *BFD) CSVRecord() []string {
	// prevent accessing nil pointer
	if b.Context == nil {
		b.Context = &PacketContext{}
	}

	return filter([]string{
		formatTimestamp(b.Timestamp),
		formatInt32(b.Version),                        // int32
		formatInt32(b.Diagnostic),                     // int32
		formatInt32(b.State),                          // int32
		strconv.FormatBool(b.Poll),                    // bool
		strconv.FormatBool(b.Final),                   // bool
		strconv.FormatBool(b.ControlPlaneIndependent), // bool
		strconv.FormatBool(b.AuthPresent),             // bool
		strconv.FormatBool(b.Demand),                  // bool
		strconv.FormatBool(b.Multipoint),              // bool
		formatInt32(b.DetectMultiplier),               // int32
		formatInt32(b.MyDiscriminator),                // int32
		formatInt32(b.YourDiscriminator),              // int32
		formatInt32(b.DesiredMinTxInterval),           // int32
		formatInt32(b.RequiredMinRxInterval),          // int32
		formatInt32(b.RequiredMinEchoRxInterval),      // int32
		b.AuthHeader.getString(),                      // *BFDAuthHeader
		b.Context.SrcIP,
		b.Context.DstIP,
		b.Context.SrcPort,
		b.Context.DstPort,
	})
}

// Time returns the timestamp associated with the audit record.
func (b *BFD) Time() string {
	return b.Timestamp
}

func (bah BFDAuthHeader) getString() string {
	var b strings.Builder

	b.WriteString(StructureBegin)
	b.WriteString(formatInt32(bah.AuthType))
	b.WriteString(FieldSeparator)
	b.WriteString(formatInt32(bah.KeyID))
	b.WriteString(FieldSeparator)
	b.WriteString(formatInt32(bah.SequenceNumber))
	b.WriteString(FieldSeparator)
	b.WriteString(hex.EncodeToString(bah.Data))
	b.WriteString(StructureEnd)

	return b.String()
}

// JSON returns the JSON representation of the audit record.
func (b *BFD) JSON() (string, error) {
	b.Timestamp = utils.TimeToUnixMilli(b.Timestamp)
	return jsonMarshaler.MarshalToString(b)
}

var bfdMetric = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: strings.ToLower(Type_NC_BFD.String()),
		Help: Type_NC_BFD.String() + " audit records",
	},
	fieldsBFD[1:],
)

// Inc increments the metrics for the audit record.
func (b *BFD) Inc() {
	bfdMetric.WithLabelValues(b.CSVRecord()[1:]...).Inc()
}

// SetPacketContext sets the associated packet context for the audit record.
func (b *BFD) SetPacketContext(ctx *PacketContext) {
	b.Context = ctx
}

// Src returns the source address of the audit record.
func (b *BFD) Src() string {
	if b.Context != nil {
		return b.Context.SrcIP
	}

	return ""
}

// Dst returns the destination address of the audit record.
func (b *BFD) Dst() string {
	if b.Context != nil {
		return b.Context.DstIP
	}

	return ""
}
