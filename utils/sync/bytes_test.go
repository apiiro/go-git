package sync

import (
	"bytes"
	"testing"
)

func TestGetAndPutBytesBuffer(t *testing.T) {
	buf := GetBytesBuffer()
	if buf == nil {
		t.Error("nil was not expected")
	}

	initialLen := buf.Len()
	buf.Grow(initialLen * 2)
	grownLen := buf.Len()

	PutBytesBuffer(buf)

	buf = GetBytesBuffer()
	if buf.Len() != grownLen {
		t.Error("bytes buffer was not reused")
	}

	buf2 := GetBytesBuffer()
	if buf2.Len() != initialLen {
		t.Errorf("new bytes buffer length: wanted %d got %d", initialLen, buf2.Len())
	}
}

func TestPoolableBytesBufferCapacity(t *testing.T) {
	if !poolable(bytes.NewBuffer(make([]byte, 0, maxPooledBytesBufferCap))) {
		t.Errorf("a buffer of capacity %d must be pooled", maxPooledBytesBufferCap)
	}
	if poolable(bytes.NewBuffer(make([]byte, 0, maxPooledBytesBufferCap+1))) {
		t.Errorf("a buffer of capacity %d must not be pooled", maxPooledBytesBufferCap+1)
	}
}

func TestGetAndPutByteSlice(t *testing.T) {
	slice := GetByteSlice()
	if slice == nil {
		t.Error("nil was not expected")
	}

	wanted := 16 * 1024
	got := len(*slice)
	if wanted != got {
		t.Errorf("byte slice length: wanted %d got %d", wanted, got)
	}

	newByteSlice := make([]byte, wanted*2)
	PutByteSlice(&newByteSlice)

	newSlice := GetByteSlice()
	if len(*newSlice) != len(newByteSlice) {
		t.Error("byte slice was not reused")
	}
}
