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

var dot1QEncoder = CreateLayerEncoder(types.Type_NC_Dot1Q, layers.LayerTypeDot1Q, func(layer gopacket.Layer, timestamp string) proto.Message {
	if dot1q, ok := layer.(*layers.Dot1Q); ok {
		return &types.Dot1Q{
			Timestamp:      timestamp,
			Priority:       int32(dot1q.Priority),
			DropEligible:   dot1q.DropEligible,
			VLANIdentifier: int32(dot1q.VLANIdentifier),
			Type:           int32(dot1q.Type),
		}
	}
	return nil
})
