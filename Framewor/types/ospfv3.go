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

var fieldsOSPFv3 = []string{
	"Timestamp",
	"Version",      // int32
	"Type",         // int32
	"PacketLength", // int32
	"RouterID",     // uint32
	"AreaID",       // uint32
	"Checksum",     // int32
	"Instance",     // int32
	"Reserved",     // int32
	"Hello",        // *HelloPkg
	"DbDesc",       // *DbDescPkg
	"LSR",          // []*LSReq
	"LSU",          // *LSUpdate
	"LSAs",         // []*LSAheader
	"SrcIP",
	"DstIP",
}

func (a OSPFv3) CSVHeader() []string {
	return filter(fieldsOSPFv3)
}

func (a OSPFv3) CSVRecord() []string {
	var (
		lsas   []string
		lsreqs []string
	)
	for _, l := range a.LSAs {
		lsas = append(lsas, l.ToString())
	}
	for _, l := range a.LSR {
		lsreqs = append(lsreqs, l.ToString())
	}
	// prevent accessing nil pointer
	if a.Context == nil {
		a.Context = &PacketContext{}
	}
	return filter([]string{
		formatTimestamp(a.Timestamp),
		formatInt32(a.Version),      // int32
		formatInt32(a.Type),         // int32
		formatInt32(a.PacketLength), // int32
		formatUint32(a.RouterID),    // uint32
		formatUint32(a.AreaID),      // uint32
		formatInt32(a.Checksum),     // int32
		formatInt32(a.Instance),     // int32
		formatInt32(a.Reserved),     // int32
		toString(a.Hello),           // *HelloPkg
		toString(a.DbDesc),          // *DbDescPkg
		join(lsreqs...),             // []*LSReq
		toString(a.LSU),             // *LSUpdate
		join(lsas...),               // []*LSAheader
		a.Context.SrcIP,
		a.Context.DstIP,
	})
}

func (a OSPFv3) Time() string {
	return a.Timestamp
}

func (l HelloPkg) ToString() string {

	var b strings.Builder

	b.WriteString(Begin)
	b.WriteString(formatUint32(l.InterfaceID)) // uint32
	b.WriteString(Separator)
	b.WriteString(formatInt32(l.RtrPriority)) // int32
	b.WriteString(Separator)
	b.WriteString(formatUint32(l.Options)) // uint32
	b.WriteString(Separator)
	b.WriteString(formatInt32(l.HelloInterval)) // int32
	b.WriteString(Separator)
	b.WriteString(formatUint32(l.RouterDeadInterval)) // uint32
	b.WriteString(Separator)
	b.WriteString(formatUint32(l.DesignatedRouterID)) // uint32
	b.WriteString(Separator)
	b.WriteString(formatUint32(l.BackupDesignatedRouterID)) // uint32
	b.WriteString(Separator)
	b.WriteString(joinUints(l.NeighborID)) // []uint32
	b.WriteString(End)

	return b.String()
}

func (u HelloPkg) JSON() (string, error) {
	return jsonMarshaler.MarshalToString(&u)
}

var ospf3Metric = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: strings.ToLower(Type_NC_OSPFv3.String()),
		Help: Type_NC_OSPFv3.String() + " audit records",
	},
	fieldsOSPFv3[1:],
)

func init() {
	prometheus.MustRegister(ospf3Metric)
}

func (a OSPFv3) Inc() {
	ospf3Metric.WithLabelValues(a.CSVRecord()[1:]...).Inc()
}

func (a OSPFv3) JSON() (string, error) {
	return jsonMarshaler.MarshalToString(&a)
}

func (a *OSPFv3) SetPacketContext(ctx *PacketContext) {
	a.Context = ctx
}

func (a OSPFv3) Src() string {
	if a.Context != nil {
		return a.Context.SrcIP
	}
	return ""
}

func (a OSPFv3) Dst() string {
	if a.Context != nil {
		return a.Context.DstIP
	}
	return ""
}
