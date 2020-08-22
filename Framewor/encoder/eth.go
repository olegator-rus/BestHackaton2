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

var ethernetEncoder = CreateLayerEncoder(types.Type_NC_Ethernet, layers.LayerTypeEthernet, func(layer gopacket.Layer, timestamp string) proto.Message {
	if eth, ok := layer.(*layers.Ethernet); ok {
		return &types.Ethernet{
			Timestamp:      timestamp,
			SrcMAC:         eth.SrcMAC.String(),
			DstMAC:         eth.DstMAC.String(),
			EthernetType:   int32(eth.EthernetType),
			PayloadEntropy: Entropy(eth.Payload),
			PayloadSize:    int32(len(eth.Payload)),
		}
	}
	return nil
})
