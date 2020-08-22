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

var ipv6HopByHopEncoder = CreateLayerEncoder(
	types.Type_NC_IPv6HopByHop,
	layers.LayerTypeIPv6HopByHop,
	func(layer gopacket.Layer, timestamp string) proto.Message {
		if ip6hop, ok := layer.(*layers.IPv6HopByHop); ok {
			var options []*types.IPv6HopByHopOption
			for _, o := range ip6hop.Options {
				var a *types.IPv6HopByHopOptionAlignment
				if len(o.OptionAlignment) == 2 {
					a = &types.IPv6HopByHopOptionAlignment{
						One: int32(o.OptionAlignment[0]),
						Two: int32(o.OptionAlignment[1]),
					}
				}
				options = append(options, &types.IPv6HopByHopOption{
					OptionType:      int32(o.OptionType),
					OptionLength:    int32(o.OptionLength),
					ActualLength:    int32(o.ActualLength),
					OptionData:      o.OptionData,
					OptionAlignment: a,
				})
			}
			return &types.IPv6HopByHop{
				Timestamp: timestamp,
				Options:   options,
			}
		}
		return nil
	})
