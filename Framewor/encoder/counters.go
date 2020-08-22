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
	"sync"
)

////////////////////////
// Atomic Counter Map //
////////////////////////

// AtomicCounterMap maps strings to integers
type AtomicCounterMap struct {
	Items map[string]int64
	sync.Mutex
}

// NewAtomicCounterMap returns a new AtomicCounterMap
func NewAtomicCounterMap() *AtomicCounterMap {
	return &AtomicCounterMap{
		Items: map[string]int64{},
	}
}

// Inc increments a value
func (a *AtomicCounterMap) Inc(val string) {
	a.Lock()
	a.Items[val]++
	a.Unlock()
}
