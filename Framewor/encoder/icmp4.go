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

package encoder

import (
	"github.com/dreadl0ck/netcap/types"
	"github.com/golang/protobuf/proto"
	"github.com/dreadl0ck/gopacket"
	"github.com/dreadl0ck/gopacket/layers"
)

var icmpv4Encoder = CreateLayerEncoder(types.Type_NC_ICMPv4, layers.LayerTypeICMPv4, func(layer gopacket.Layer, timestamp string) proto.Message {
	if icmp4, ok := layer.(*layers.ICMPv4); ok {
		return &types.ICMPv4{
			Timestamp: timestamp,
			TypeCode:  int32(icmp4.TypeCode),
			Checksum:  int32(icmp4.Checksum),
			Id:        int32(icmp4.Id),
			Seq:       int32(icmp4.Seq),
		}
	}
	return nil
})
