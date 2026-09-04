// SPDX-License-Identifier: Apache-2.0

package wdb

import (
	"container/heap"
	"fmt"
	"sort"
)

// Stream calls fn for every value change of every logged object, in
// ascending time order, which is the order a waveform file is written
// in. Changes decodes one object over the whole file; Stream decodes
// every object in one pass, so a conversion holds the objects' values
// and not their history: the values together are the file's handle
// space, 2.8 MB for //hdl/neorv32:sim, whose changes number 18875466.
//
// The value passed to fn is the decoder's own buffer and is valid
// until fn returns; a caller that keeps it copies it. obj indexes
// File.Objects.
//
// Within one time the changes come object by object in object order,
// and within one object in the order Changes gives them. Records of
// one time in different arenas have no order between them in the file,
// so neither has this one; see the note on Changes.
func (f *File) Stream(fn func(obj int, t uint64, data []byte) error) error {
	decs := make([]*decoder, len(f.Objects))
	// Objects are indexed by the arena sized blocks their values cover,
	// so a record finds the objects it reaches without a scan.
	byBlock := map[int][]int{}
	for i, o := range f.Objects {
		if !o.Logged {
			continue
		}
		d, err := f.newDecoder(o)
		if err != nil {
			return fmt.Errorf("object %d (%s): %w", i, f.ObjectPath(o), err)
		}
		decs[i] = d
		for b := int(d.start / arenaSpan); b <= int((d.end-1)/arenaSpan); b++ {
			byBlock[b] = append(byBlock[b], i)
		}
	}
	pending := make([][]rec, len(f.Objects))
	seen := make([]bool, len(f.Objects))
	var touched []int
	var group []rec
	var cbErr error
	it := newRecordIter(f)
	for it.next(&group) {
		for _, r := range group {
			lo := int(r.addr / arenaSpan)
			hi := int((r.addr + uint64(len(r.data)) - 1) / arenaSpan)
			for b := lo; b <= hi; b++ {
				for _, i := range byBlock[b] {
					if seen[i] {
						continue
					}
					x, ok := decs[i].take(r)
					if !ok {
						continue
					}
					seen[i] = true
					if len(pending[i]) == 0 {
						touched = append(touched, i)
					}
					pending[i] = append(pending[i], x)
				}
			}
			// seen keeps a record that spans several blocks from
			// reaching one object twice.
			for b := lo; b <= hi; b++ {
				for _, i := range byBlock[b] {
					seen[i] = false
				}
			}
		}
		sort.Ints(touched)
		for _, i := range touched {
			err := decs[i].group(pending[i], func(t uint64, data []byte) {
				if cbErr == nil {
					cbErr = fn(i, t, data)
				}
			})
			if err != nil {
				return fmt.Errorf("object %d (%s): %w", i, f.ObjectPath(f.Objects[i]), err)
			}
			if cbErr != nil {
				return cbErr
			}
			pending[i] = pending[i][:0]
		}
		touched = touched[:0]
	}
	for i, d := range decs {
		if d != nil && !d.started {
			return fmt.Errorf("object %d (%s) is logged but has no records", i, f.ObjectPath(f.Objects[i]))
		}
	}
	return nil
}

// recordIter walks every record of the file in ascending time. An
// arena's pages are in ascending time and so are the records inside a
// page, so the walk is a merge of the arenas, and the records of one
// time come back in arena order, which is the order Changes sees them.
type recordIter struct {
	f    *File
	at   []arenaCursor
	heap arenaHeap
}

// arenaCursor is one arena's place in the walk.
type arenaCursor struct {
	page, idx int
	time      uint64
}

// arenaHeap orders the arenas by the time of the record each is at,
// and arenas at one time by their index.
type arenaHeap struct {
	arenas []int
	at     []arenaCursor
}

func (h arenaHeap) Len() int { return len(h.arenas) }
func (h arenaHeap) Less(i, j int) bool {
	a, b := h.at[h.arenas[i]], h.at[h.arenas[j]]
	if a.time != b.time {
		return a.time < b.time
	}
	return h.arenas[i] < h.arenas[j]
}
func (h arenaHeap) Swap(i, j int) { h.arenas[i], h.arenas[j] = h.arenas[j], h.arenas[i] }
func (h *arenaHeap) Push(x any)   { h.arenas = append(h.arenas, x.(int)) }
func (h *arenaHeap) Pop() any {
	n := len(h.arenas) - 1
	x := h.arenas[n]
	h.arenas = h.arenas[:n]
	return x
}

func newRecordIter(f *File) *recordIter {
	it := &recordIter{f: f, at: make([]arenaCursor, len(f.Arenas))}
	it.heap.at = it.at
	for k := range f.Arenas {
		if it.advance(k) {
			it.heap.arenas = append(it.heap.arenas, k)
		}
	}
	heap.Init(&it.heap)
	return it
}

// advance moves arena k to its next record and reports whether it has
// one. A page with no records is skipped.
func (it *recordIter) advance(k int) bool {
	c := &it.at[k]
	for c.page < len(it.f.Pages[k]) {
		if c.idx < len(it.f.Pages[k][c.page].Records) {
			c.time = it.f.Pages[k][c.page].Records[c.idx].Time
			return true
		}
		c.page++
		c.idx = 0
	}
	return false
}

// next fills group with every record of the next time, in arena order,
// and reports whether there was one.
func (it *recordIter) next(group *[]rec) bool {
	*group = (*group)[:0]
	if it.heap.Len() == 0 {
		return false
	}
	t := it.at[it.heap.arenas[0]].time
	for it.heap.Len() > 0 {
		k := it.heap.arenas[0]
		if it.at[k].time != t {
			break
		}
		c := &it.at[k]
		live := true
		for live {
			r := it.f.Pages[k][c.page].Records[c.idx]
			if r.Time != t {
				break
			}
			*group = append(*group, rec{r.Time, uint64(k)*arenaSpan + uint64(r.Key), r.Data})
			c.idx++
			live = it.advance(k)
		}
		if live {
			heap.Fix(&it.heap, 0)
		} else {
			heap.Pop(&it.heap)
		}
	}
	return true
}
