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

var fieldsUSB = []string{
	"Timestamp",
	"ID",
	"EventType",
	"TransferType",
	"Direction",
	"EndpointNumber",
	"DeviceAddress",
	"BusID",
	"TimestampSec",
	"TimestampUsec",
	"Setup",
	"Data",
	"Status",
	"UrbLength",
	"UrbDataLength",
	"UrbInterval",
	"UrbStartFrame",
	"UrbCopyOfTransferFlags",
	"IsoNumDesc",
	"Payload",
}

// CSVHeader returns the CSV header for the audit record.
func (u *USB) CSVHeader() []string {
	return filter(fieldsUSB)
}

// CSVRecord returns the CSV record for the audit record.
func (u *USB) CSVRecord() []string {
	return filter([]string{
		formatTimestamp(u.Timestamp), // string
		formatUint64(u.ID),
		formatInt32(u.EventType),
		formatInt32(u.TransferType),
		formatInt32(u.Direction),
		formatInt32(u.EndpointNumber),
		formatInt32(u.DeviceAddress),
		formatInt32(u.BusID),
		formatInt64(u.TimestampSec),
		formatInt32(u.TimestampUsec),
		strconv.FormatBool(u.Setup),
		strconv.FormatBool(u.Data),
		formatInt32(u.Status),
		formatUint32(u.UrbLength),
		formatUint32(u.UrbDataLength),
		formatUint32(u.UrbInterval),
		formatUint32(u.UrbStartFrame),
		formatUint32(u.UrbCopyOfTransferFlags),
		formatUint32(u.IsoNumDesc),
		hex.EncodeToString(u.Payload),
	})
}

// Time returns the timestamp associated with the audit record.
func (u *USB) Time() string {
	return u.Timestamp
}

// JSON returns the JSON representation of the audit record.
func (u *USB) JSON() (string, error) {
	u.Timestamp = utils.TimeToUnixMilli(u.Timestamp)
	return jsonMarshaler.MarshalToString(u)
}

var usbMetric = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: strings.ToLower(Type_NC_USB.String()),
		Help: Type_NC_USB.String() + " audit records",
	},
	fieldsUSB[1:],
)

// Inc increments the metrics for the audit record.
func (u *USB) Inc() {
	usbMetric.WithLabelValues(u.CSVRecord()[1:]...).Inc()
}

// SetPacketContext sets the associated packet context for the audit record.
func (u *USB) SetPacketContext(*PacketContext) {}

// Src TODO return source DeviceAddress?
// Src returns the source address of the audit record.
func (u *USB) Src() string {
	return ""
}

// Dst TODO return destination DeviceAddress?
// Dst returns the destination address of the audit record.
func (u *USB) Dst() string {
	return ""
}
