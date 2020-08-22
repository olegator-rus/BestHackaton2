// +build windows

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

package dpi

// this file function stubs for windows that do nothing, but allow us to compile
// getting the C bindings to cross compile for windows is a pain
// so currently no DPI support for windows

import (
	"github.com/dreadl0ck/gopacket"

	"github.com/dreadl0ck/netcap/types"
)

func Init() {}

func Destroy() {}

func GetProtocols(packet gopacket.Packet) map[string]struct{} {
	uniqueResults := make(map[string]struct{})

	return uniqueResults
}

func NewProto(i *struct{}) *types.Protocol {
	return &types.Protocol{}
}
