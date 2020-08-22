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

// udpDataSlice implements sort.Interface to sort data fragments based on their timestamps.
type udpDataSlice []*udpData

// Len will return the length.
func (d udpDataSlice) Len() int {
	return len(d)
}

// Less will return true if the value at index i is smaller than the other one.
func (d udpDataSlice) Less(i, j int) bool {
	return d[i].ci.Timestamp.Before(d[j].ci.Timestamp)
}

// Swap will switch the values.
func (d udpDataSlice) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}
