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

package decoder

import (
	"github.com/dreadl0ck/gopacket"
	"github.com/dreadl0ck/gopacket/layers"
	"github.com/gogo/protobuf/proto"

	"github.com/dreadl0ck/netcap/types"
)

var icmpv6NeighborAdvertisementDecoder = newGoPacketDecoder(
	types.Type_NC_ICMPv6NeighborAdvertisement,
	layers.LayerTypeICMPv6NeighborAdvertisement,
	"The Internet Control Message Protocol (ICMP) is a supporting protocol in the Internet protocol suite",
	func(layer gopacket.Layer, timestamp string) proto.Message {
		if icmp6na, ok := layer.(*layers.ICMPv6NeighborAdvertisement); ok {
			var opts []*types.ICMPv6Option
			for _, o := range icmp6na.Options {
				opts = append(opts, &types.ICMPv6Option{
					Data: o.Data,
					Type: int32(o.Type),
				})
			}

			return &types.ICMPv6NeighborAdvertisement{
				Timestamp:     timestamp,
				Flags:         int32(icmp6na.Flags),
				TargetAddress: icmp6na.TargetAddress.String(),
				Options:       opts,
			}
		}

		return nil
	},
)
