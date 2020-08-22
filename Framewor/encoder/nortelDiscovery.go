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

var nortelDiscoveryEncoder = CreateLayerEncoder(types.Type_NC_NortelDiscovery, layers.LayerTypeNortelDiscovery, func(layer gopacket.Layer, timestamp string) proto.Message {
	if nortel, ok := layer.(*layers.NortelDiscovery); ok {
		return &types.NortelDiscovery{
			Timestamp: timestamp,
			IPAddress: string(nortel.IPAddress),
			SegmentID: []byte(nortel.SegmentID),
			Chassis:   int32(nortel.Chassis),
			Backplane: int32(nortel.Backplane),
			State:     int32(nortel.State),
			NumLinks:  int32(nortel.NumLinks),
		}
	}
	return nil
})
