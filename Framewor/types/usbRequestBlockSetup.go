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
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

var fieldsUSBRequestBlockSetup = []string{
	"Timestamp",
	"RequestType", // int32
	"Request",     // int32
	"Value",       // int32
	"Index",       // int32
	"Length",      // int32
}

func (a USBRequestBlockSetup) CSVHeader() []string {
	return filter(fieldsUSBRequestBlockSetup)
}

func (a USBRequestBlockSetup) CSVRecord() []string {
	return filter([]string{
		formatTimestamp(a.Timestamp),
		formatInt32(a.RequestType), // int32
		formatInt32(a.Request),     // int32
		formatInt32(a.Value),       // int32
		formatInt32(a.Index),       // int32
		formatInt32(a.Length),      // int32
	})
}

func (a USBRequestBlockSetup) Time() string {
	return a.Timestamp
}

func (a USBRequestBlockSetup) JSON() (string, error) {
	return jsonMarshaler.MarshalToString(&a)
}

var usbRequestBlockSetupMetric = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: strings.ToLower(Type_NC_USBRequestBlockSetup.String()),
		Help: Type_NC_USBRequestBlockSetup.String() + " audit records",
	},
	fieldsUSBRequestBlockSetup[1:],
)

func init() {
	prometheus.MustRegister(usbRequestBlockSetupMetric)
}

func (a USBRequestBlockSetup) Inc() {
	usbRequestBlockSetupMetric.WithLabelValues(a.CSVRecord()[1:]...).Inc()
}

func (a *USBRequestBlockSetup) SetPacketContext(ctx *PacketContext) {}

// TODO return source DeviceAddress?
func (a USBRequestBlockSetup) Src() string {
	return ""
}

// TODO return destination DeviceAddress?
func (a USBRequestBlockSetup) Dst() string {
	return ""
}
