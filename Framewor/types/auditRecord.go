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

// Contains the type definitions for the supported network protocols
package types

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gogo/protobuf/jsonpb"
	proto "github.com/golang/protobuf/proto"
	"github.com/mgutz/ansi"
)

var (
	selection []int

	// UTC allows to print timestamp in the utc format
	UTC bool

	jsonMarshaler = &jsonpb.Marshaler{}
)

// AuditRecord is the interface for basic operations with NETCAP audit records
// this includes dumping as CSV or JSON or prometheus metrics
// and provides access to the timestamp of the audit record
type AuditRecord interface {

	// returns CSV values
	CSVRecord() []string

	// returns CSV header fields
	CSVHeader() []string

	// used to retrieve the timestamp of the audit record for labeling
	Time() string

	// Src returns the source of an audit record
	// for Layer 2 records this shall be the MAC address
	// for Layer 3+ records this shall be the IP address
	Src() string

	// Dst returns the source of an audit record
	// for Layer 2 records this shall be the MAC address
	// for Layer 3+ records this shall be the IP address
	Dst() string

	// increments the metric for the audit record
	Inc()

	// returns the audit record as JSON
	JSON() (string, error)

	// can be implemented to set additional information for each audit record
	// important:
	//  - MUST be implemented on a pointer of an instance
	//  - the passed in packet context MUST be set on the Context field of the current audit record
	SetPacketContext(ctx *PacketContext)
}

// selectFields returns an array with the indices of the desired fields for selection
func selectFields(all []string, selection string) (s []int) {

	var (
		fields = strings.Split(selection, ",")
		ok     bool
	)

	s = make([]int, len(fields))
	for i, val := range fields {
		for index, name := range all {
			if name == val {
				s[i] = index
				ok = true
				break
			}
		}
		if !ok {
			fmt.Println("invalid field: ", ansi.Red+val+ansi.Reset)
			fmt.Println("available fields: ", ansi.Yellow+strings.Join(all, ",")+ansi.Reset)
			os.Exit(1)
		}
		ok = false
	}
	return s
}

// Select takes a proto.Message and sets the selection on the package level
func Select(msg proto.Message, vals string) {
	if vals != "" && vals != " " {
		if p, ok := msg.(AuditRecord); ok {
			selection = selectFields(p.CSVHeader(), vals)
		} else {
			fmt.Printf("type: %#v\n", msg)
			log.Fatal("type does not implement the types.AuditRecord interface")
		}
	}
}

// filter applies a selection if configured
func filter(in []string) []string {
	if len(selection) == 0 {
		return in
	}
	r := make([]string, len(selection))
	for i, v := range selection {
		r[i] = in[v]
	}
	return r
}
