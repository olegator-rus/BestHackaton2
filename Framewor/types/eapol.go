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

var fieldsEAPOL = []string{
	"Timestamp",
	"Version", //  int32
	"Type",    //  int32
	"Length",  //  int32
}

func (a EAPOL) CSVHeader() []string {
	return filter(fieldsEAPOL)
}

func (a EAPOL) CSVRecord() []string {
	return filter([]string{
		formatTimestamp(a.Timestamp),
		formatInt32(a.Version), //  int32
		formatInt32(a.Type),    //  int32
		formatInt32(a.Length),  //  int32
	})
}

func (a EAPOL) Time() string {
	return a.Timestamp
}

func (a EAPOL) JSON() (string, error) {
	return jsonMarshaler.MarshalToString(&a)
}

var eapPolMetric = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: strings.ToLower(Type_NC_EAPOL.String()),
		Help: Type_NC_EAPOL.String() + " audit records",
	},
	fieldsEAPOL[1:],
)

func init() {
	prometheus.MustRegister(eapPolMetric)
}

func (a EAPOL) Inc() {
	eapPolMetric.WithLabelValues(a.CSVRecord()[1:]...).Inc()
}

func (a *EAPOL) SetPacketContext(ctx *PacketContext) {}

// TODO: return Mac addr
func (a EAPOL) Src() string {
	return ""
}

func (a EAPOL) Dst() string {
	return ""
}
