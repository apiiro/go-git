package idxfile

import (
	"bytes"
	"math"
	"sort"

	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/utils/binary"
)

// objects implements sort.Interface and uses hash as sorting key.
type objects []Entry

// Writer implements a packfile Observer interface and is used to generate
// indexes.
type Writer struct {
	count    uint32
	checksum plumbing.Hash
	objects  objects
	offset64 uint32
	idx      *MemoryIndex
}

// Create index returns a filled MemoryIndex with the information filled by
// the observer callbacks.
func (w *Writer) CreateIndex() (*MemoryIndex, error) {
	idx := new(MemoryIndex)
	w.idx = idx

	sort.Sort(w.objects)

	// unmap all fans by default
	for i := range idx.FanoutMapping {
		idx.FanoutMapping[i] = noMapping
	}

	buf := new(bytes.Buffer)

	last := -1
	bucket := -1
	for i, o := range w.objects {
		fan := o.Hash[0]

		// fill the gaps between fans
		for j := last + 1; j < int(fan); j++ {
			idx.Fanout[j] = uint32(i)
		}

		// update the number of objects for this position
		idx.Fanout[fan] = uint32(i + 1)

		// we move from one bucket to another, update counters and allocate
		// memory
		if last != int(fan) {
			bucket++
			idx.FanoutMapping[fan] = bucket
			last = int(fan)

			idx.Names = append(idx.Names, make([]byte, 0))
			idx.Offset32 = append(idx.Offset32, make([]byte, 0))
			idx.Crc32 = append(idx.Crc32, make([]byte, 0))
		}

		idx.Names[bucket] = append(idx.Names[bucket], o.Hash[:]...)

		offset := o.Offset
		if offset > math.MaxInt32 {
			offset = w.addOffset64(offset)
		}

		buf.Truncate(0)
		binary.WriteUint32(buf, uint32(offset))
		idx.Offset32[bucket] = append(idx.Offset32[bucket], buf.Bytes()...)

		buf.Truncate(0)
		binary.WriteUint32(buf, uint32(o.CRC32))
		idx.Crc32[bucket] = append(idx.Crc32[bucket], buf.Bytes()...)
	}

	for j := last + 1; j < 256; j++ {
		idx.Fanout[j] = uint32(len(w.objects))
	}

	idx.Version = VersionSupported
	idx.PackfileChecksum = w.checksum

	return idx, nil
}

func (w *Writer) addOffset64(pos uint64) uint64 {
	buf := new(bytes.Buffer)
	binary.WriteUint64(buf, pos)
	w.idx.Offset64 = append(w.idx.Offset64, buf.Bytes()...)

	index := uint64(w.offset64 | (1 << 31))
	w.offset64++

	return index
}

// Add appends new object data.
func (w *Writer) Add(h plumbing.Hash, pos uint64, crc uint32) {
	w.objects = append(w.objects, Entry{h, crc, pos})
}

// OnHeader implements packfile.Observer interface.
func (w *Writer) OnHeader(count uint32) error {
	w.count = count
	w.objects = make(objects, 0, count)
	return nil
}

// OnInflatedObjectHeader implements packfile.Observer interface.
func (w *Writer) OnInflatedObjectHeader(t plumbing.ObjectType, objSize int64, pos int64) error {
	return nil
}

// OnInflatedObjectContent implements packfile.Observer interface.
func (w *Writer) OnInflatedObjectContent(h plumbing.Hash, pos int64, crc uint32) error {
	w.Add(h, uint64(pos), crc)
	return nil
}

// OnFooter implements packfile.Observer interface.
func (w *Writer) OnFooter(h plumbing.Hash) error {
	w.checksum = h
	return nil
}

func (o objects) Len() int {
	return len(o)
}

func (o objects) Less(i int, j int) bool {
	cmp := bytes.Compare(o[i].Hash[:], o[j].Hash[:])
	return cmp < 0
}

func (o objects) Swap(i int, j int) {
	o[i], o[j] = o[j], o[i]
}
