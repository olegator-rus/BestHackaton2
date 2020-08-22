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
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

var fieldsTCP = []string{
	"Timestamp",
	"SrcPort",
	"DstPort",
	"SeqNum",
	"AckNum",
	"DataOffset",
	"FIN",
	"SYN",
	"RST",
	"PSH",
	"ACK",
	"URG",
	"ECE",
	"CWR",
	"NS",
	"Window",
	"Checksum",
	"Urgent",
	"Padding",
	"Options",
	"PayloadEntropy",
	"PayloadSize",
	"Payload",
	"SrcIP",
	"DstIP",
}

func (t TCP) CSVHeader() []string {
	return filter(fieldsTCP)
}

func (t TCP) CSVRecord() []string {
	// prevent accessing nil pointer
	if t.Context == nil {
		t.Context = &PacketContext{}
	}
	return filter([]string{
		formatTimestamp(t.Timestamp),                      // string
		formatInt32(t.SrcPort),                            // int32
		formatInt32(t.DstPort),                            // int32
		strconv.FormatUint(uint64(t.SeqNum), 10),          // uint32
		strconv.FormatUint(uint64(t.AckNum), 10),          // uint32
		formatInt32(t.DataOffset),                         // int32
		strconv.FormatBool(t.FIN),                         // bool
		strconv.FormatBool(t.SYN),                         // bool
		strconv.FormatBool(t.RST),                         // bool
		strconv.FormatBool(t.PSH),                         // bool
		strconv.FormatBool(t.ACK),                         // bool
		strconv.FormatBool(t.URG),                         // bool
		strconv.FormatBool(t.ECE),                         // bool
		strconv.FormatBool(t.CWR),                         // bool
		strconv.FormatBool(t.NS),                          // bool
		formatInt32(t.Window),                             // int32
		formatInt32(t.Checksum),                           // int32
		formatInt32(t.Urgent),                             // int32
		string(t.Padding),                                 // []byte
		t.GetOptionString(),                               // []*TCPOption
		strconv.FormatFloat(t.PayloadEntropy, 'f', 8, 64), // float64
		formatInt32(t.PayloadSize),                        // int32
		hex.EncodeToString(t.Payload),
		t.Context.SrcIP,
		t.Context.DstIP,
	})
}

func (t TCP) GetOptionString() string {
	var b strings.Builder
	for _, o := range t.Options {
		b.WriteString(Begin)
		b.WriteString(strconv.Itoa(int(o.OptionType)))
		b.WriteString(Separator)
		b.WriteString(strconv.Itoa(int(o.OptionLength)))
		b.WriteString(Separator)
		b.WriteString(hex.EncodeToString(o.OptionData))
		b.WriteString(End)
	}
	return b.String()
}

func (f TCP) Time() string {
	return f.Timestamp
}

func (u TCP) JSON() (string, error) {
	return jsonMarshaler.MarshalToString(&u)
}

var (
	tcpMetric = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: strings.ToLower(Type_NC_TCP.String()),
			Help: Type_NC_TCP.String() + " audit records",
		},
		fieldsTCPMetrics,
	)
	tcpPayloadEntropy = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    strings.ToLower(Type_NC_TCP.String()) + "_entropy",
			Help:    Type_NC_TCP.String() + " payload entropy",
			Buckets: prometheus.LinearBuckets(20, 5, 5),
		},
		// []string{"SrcPort", "DstPort"},
		[]string{},
	)
	tcpPayloadSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    strings.ToLower(Type_NC_TCP.String()) + "_size",
			Help:    Type_NC_TCP.String() + " payload sizes",
			Buckets: prometheus.LinearBuckets(20, 5, 5),
		},
		// []string{"SrcPort", "DstPort"},
		[]string{},
	)
)

var fieldsTCPMetrics = []string{
	"SrcPort",
	"DstPort",
	// "SeqNum",
	// "AckNum",
	"DataOffset",
	"FIN",
	"SYN",
	"RST",
	"PSH",
	"ACK",
	"URG",
	"ECE",
	"CWR",
	"NS",
	// "Window",
	"Urgent",
	// "Padding",
	// "Options",
}

func (t TCP) metricValues() []string {
	return []string{
		formatInt32(t.SrcPort), // int32
		formatInt32(t.DstPort), // int32
		// strconv.FormatUint(uint64(t.SeqNum), 10), // uint32
		// strconv.FormatUint(uint64(t.AckNum), 10), // uint32
		formatInt32(t.DataOffset), // int32
		strconv.FormatBool(t.FIN), // bool
		strconv.FormatBool(t.SYN), // bool
		strconv.FormatBool(t.RST), // bool
		strconv.FormatBool(t.PSH), // bool
		strconv.FormatBool(t.ACK), // bool
		strconv.FormatBool(t.URG), // bool
		strconv.FormatBool(t.ECE), // bool
		strconv.FormatBool(t.CWR), // bool
		strconv.FormatBool(t.NS),  // bool
		// formatInt32(t.Window),     // int32
		formatInt32(t.Urgent), // int32
		// string(t.Padding),                        // []byte
		// t.GetOptionString(),                      // []*TCPOption
	}
}

func init() {
	prometheus.MustRegister(tcpMetric)
	prometheus.MustRegister(tcpPayloadEntropy)
	prometheus.MustRegister(tcpPayloadSize)
}

func (a TCP) Inc() {
	tcpMetric.WithLabelValues(a.metricValues()...).Inc()
	tcpPayloadEntropy.WithLabelValues().Observe(a.PayloadEntropy)
	tcpPayloadSize.WithLabelValues().Observe(float64(a.PayloadSize))
}

func (a *TCP) SetPacketContext(ctx *PacketContext) {

	// create new context and only add information that is
	// not yet present on the audit record type
	a.Context = &PacketContext{
		SrcIP: ctx.SrcIP,
		DstIP: ctx.DstIP,
	}
}

func (a TCP) Src() string {
	if a.Context != nil {
		return a.Context.SrcIP
	}
	return ""
}

func (a TCP) Dst() string {
	if a.Context != nil {
		return a.Context.DstIP
	}
	return ""
}
