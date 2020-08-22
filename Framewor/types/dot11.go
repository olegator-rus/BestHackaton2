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

var fieldsDot11 = []string{
	"Timestamp",
	"Type",           // int32
	"Proto",          // int32
	"Flags",          // int32
	"DurationID",     // int32
	"Address1",       // string
	"Address2",       // string
	"Address3",       // string
	"Address4",       // string
	"SequenceNumber", // int32
	"FragmentNumber", // int32
	"Checksum",       // uint32
	"QOS",            // *Dot11QOS
	"HTControl",      // *Dot11HTControl
}

func (d Dot11) CSVHeader() []string {
	return filter(fieldsDot11)
}

func (d Dot11) CSVRecord() []string {
	return filter([]string{
		formatTimestamp(d.Timestamp),
		formatInt32(d.Type),           // int32
		formatInt32(d.Proto),          // int32
		formatInt32(d.Flags),          // int32
		formatInt32(d.DurationID),     // int32
		d.Address1,                    // string
		d.Address2,                    // string
		d.Address3,                    // string
		d.Address4,                    // string
		formatInt32(d.SequenceNumber), // int32
		formatInt32(d.FragmentNumber), // int32
		formatUint32(d.Checksum),      // uint32
		d.QOS.ToString(),              // *Dot11QOS
		d.HTControl.ToString(),        // *Dot11HTControl
	})
}

func (d Dot11) Time() string {
	return d.Timestamp
}

func (d Dot11QOS) ToString() string {
	var b strings.Builder
	b.WriteString(Begin)
	b.WriteString(formatInt32(d.TID))
	b.WriteString(Separator)
	b.WriteString(strconv.FormatBool(d.EOSP))
	b.WriteString(Separator)
	b.WriteString(formatInt32(d.AckPolicy))
	b.WriteString(Separator)
	b.WriteString(formatInt32(d.TXOP))
	b.WriteString(End)
	return b.String()
}

func (d Dot11HTControl) ToString() string {
	var b strings.Builder
	b.WriteString(Begin)
	b.WriteString(strconv.FormatBool(d.ACConstraint))
	b.WriteString(Separator)
	b.WriteString(strconv.FormatBool(d.RDGMorePPDU))
	b.WriteString(Separator)
	b.WriteString(d.VHT.ToString())
	b.WriteString(Separator)
	b.WriteString(d.HT.ToString())
	b.WriteString(End)
	return b.String()
}

func (d *Dot11HTControlVHT) ToString() string {
	var b strings.Builder
	b.WriteString(Begin)
	b.WriteString(strconv.FormatBool(d.MRQ))
	b.WriteString(Separator)
	b.WriteString(strconv.FormatBool(d.UnsolicitedMFB))
	b.WriteString(Separator)
	b.WriteString(formatInt32(d.MSI))
	b.WriteString(Separator)
	b.WriteString(d.MFB.ToString())
	b.WriteString(Separator)
	b.WriteString(formatInt32(d.CompressedMSI))
	b.WriteString(Separator)
	b.WriteString(strconv.FormatBool(d.STBCIndication))
	b.WriteString(Separator)
	b.WriteString(formatInt32(d.MFSI))
	b.WriteString(Separator)
	b.WriteString(formatInt32(d.GID))
	b.WriteString(Separator)
	b.WriteString(formatInt32(d.CodingType))
	b.WriteString(Separator)
	b.WriteString(strconv.FormatBool(d.FbTXBeamformed))
	b.WriteString(End)
	return b.String()
}

func (d *Dot11HTControlMFB) ToString() string {
	var b strings.Builder
	b.WriteString(Begin)
	b.WriteString(formatInt32(d.NumSTS))
	b.WriteString(Separator)
	b.WriteString(formatInt32(d.VHTMCS))
	b.WriteString(Separator)
	b.WriteString(formatInt32(d.BW))
	b.WriteString(Separator)
	b.WriteString(formatInt32(d.SNR))
	b.WriteString(End)
	return b.String()
}

func (d *Dot11HTControlHT) ToString() string {
	var b strings.Builder
	b.WriteString(Begin)
	b.WriteString(d.LinkAdapationControl.ToString())
	b.WriteString(Separator)
	b.WriteString(formatInt32(d.CalibrationPosition))
	b.WriteString(Separator)
	b.WriteString(formatInt32(d.CalibrationSequence))
	b.WriteString(Separator)
	b.WriteString(formatInt32(d.CSISteering))
	b.WriteString(Separator)
	b.WriteString(strconv.FormatBool(d.NDPAnnouncement))
	b.WriteString(Separator)
	b.WriteString(strconv.FormatBool(d.DEI))
	b.WriteString(End)
	return b.String()
}

func (d *Dot11LinkAdapationControl) ToString() string {
	var b strings.Builder
	b.WriteString(Begin)
	b.WriteString(strconv.FormatBool(d.TRQ))
	b.WriteString(Separator)
	b.WriteString(strconv.FormatBool(d.MRQ))
	b.WriteString(Separator)
	b.WriteString(formatInt32(d.MSI))
	b.WriteString(Separator)
	b.WriteString(formatInt32(d.MFSI))
	b.WriteString(Separator)
	b.WriteString(formatInt32(d.MFB))
	b.WriteString(Separator)
	b.WriteString(d.ASEL.ToString())
	b.WriteString(End)
	return b.String()
}

func (d *Dot11ASEL) ToString() string {
	var b strings.Builder
	b.WriteString(Begin)
	b.WriteString(formatInt32(d.Command))
	b.WriteString(Separator)
	b.WriteString(formatInt32(d.Data))
	b.WriteString(End)
	return b.String()
}

func (a Dot11) JSON() (string, error) {
	return jsonMarshaler.MarshalToString(&a)
}

var dot11Metric = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: strings.ToLower(Type_NC_Dot11.String()),
		Help: Type_NC_Dot11.String() + " audit records",
	},
	fieldsDot11[1:],
)

func init() {
	prometheus.MustRegister(dot11Metric)
}

func (a Dot11) Inc() {
	dot11Metric.WithLabelValues(a.CSVRecord()[1:]...).Inc()
}

func (a *Dot11) SetPacketContext(ctx *PacketContext) {}

// TODO: return Mac addr
func (a Dot11) Src() string {
	return ""
}

func (a Dot11) Dst() string {
	return ""
}
