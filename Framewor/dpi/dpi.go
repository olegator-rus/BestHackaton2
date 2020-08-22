// +build !windows

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

// Package dpi implements an interface for application layer classification via bindings to nDPI and libprotoident
package dpi

import (
	"fmt"
	"log"

	godpi "github.com/dreadl0ck/go-dpi"
	"github.com/dreadl0ck/go-dpi/modules/classifiers"
	"github.com/dreadl0ck/go-dpi/modules/wrappers"
	. "github.com/dreadl0ck/go-dpi/types"
	"github.com/dreadl0ck/gopacket"

	"github.com/dreadl0ck/netcap/types"
)

var disableDPI = true

// Init initializes the deep packet inspection engines.
func Init() {
	disableDPI = false

	var (
		nDPI  = wrappers.NewNDPIWrapper()
		lPI   = wrappers.NewLPIWrapper()
		goDPI = classifiers.NewClassifierModule()
		wm    = wrappers.NewWrapperModule()
	)

	// init DPI
	wm.ConfigureModule(wrappers.WrapperModuleConfig{Wrappers: []wrappers.Wrapper{lPI, nDPI}})
	godpi.SetModules([]Module{wm, goDPI})

	if err := godpi.Initialize(); err != nil {
		log.Fatal("goDPI initialization returned error: ", err)
	}
}

// Destroy tears down godpi and frees the memory allocated for cgo
// returned errors are logged to stdout.
func Destroy() {
	for _, e := range godpi.Destroy() {
		if e != nil {
			fmt.Println(e)
		}
	}
}

// GetProtocols returns a map of all the identified protocol names to a result datastructure
// packets are identified with libprotoident, nDPI and a few custom heuristics from godpi.
func GetProtocols(packet gopacket.Packet) map[string]ClassificationResult {
	protocols := make(map[string]ClassificationResult)

	if disableDPI {
		return protocols
	}

	// start := time.Now()
	// fmt.Println("DPI", packet.NetworkLayer().NetworkFlow(), packet.TransportLayer().TransportFlow())

	flow, _ := godpi.GetPacketFlow(packet)
	results := godpi.ClassifyFlowAllModules(flow)

	// fmt.Println(packet.NetworkLayer().NetworkFlow(), packet.TransportLayer().TransportFlow(), "complete", time.Since(start))

	// when using all modules we might receive duplicate classifications
	// so they will be deduplicated by protocol name before counting them later
	for _, r := range results {
		protocols[string(r.Protocol)] = r
	}

	return protocols
}

// NewProto initializes a new protocol.
func NewProto(res *ClassificationResult) *types.Protocol {
	return &types.Protocol{
		Packets:  1,
		Category: string(res.Class),
	}
}

// GetProtocolsTimeout returns a map of all the identified protocol names to a result datastructure
// packets are identified with libprotoident, nDPI and a few custom heuristics from godpi
// this function spawn a goroutine to allow setting a timeout for each packet
//func GetProtocolsTimeout(packet gopacket.Packet) map[string]ClassificationResult {
//	protocols := make(map[string]ClassificationResult)
//
//	if disableDPI {
//		return protocols
//	}
//
//	results := make(chan []ClassificationResult, 1)
//	go func() {
//		flow, _ := godpi.GetPacketFlow(packet)
//		results <- godpi.ClassifyFlowAllModules(flow)
//	}()
//
//	// start := time.Now()
//
//	select {
//	case res := <-results:
//
//		// fmt.Println("got result after", time.Since(start))
//
//		// when using all modules we might receive duplicate classifications
//		// so they will be deduplicated by protocol name before counting them later
//		for _, r := range res {
//			protocols[string(r.Protocol)] = r
//		}
//	case <-time.After(3 * time.Second):
//		fmt.Println("get protocols timeout", packet.NetworkLayer().NetworkFlow(), packet.TransportLayer().TransportFlow())
//	}
//
//	return protocols
//}
