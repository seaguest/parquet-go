package bits

import (
	"encoding/binary"
	"io"
)

type Writer struct {
	writer io.Writer
	length uint
	cache  uint64
	buffer []byte
}

func (w *Writer) Buffered() int {
	return len(w.buffer)
}

func (w *Writer) Reset(ww io.Writer) {
	w.writer = ww
	w.length = 0
	w.buffer = w.buffer[:0]
}

func (w *Writer) Flush() error {
	w.flush()
	_, err := w.writer.Write(w.buffer)
	w.buffer = w.buffer[:0]
	return err
}

func (w *Writer) flush() {
	b := [8]byte{}
	binary.LittleEndian.PutUint64(b[:], w.cache)
	w.buffer = append(w.buffer, b[:ByteCount(w.length)]...)
	w.length = 0
	w.cache = 0
}

func (w *Writer) WriteBit(bit int) {
	w.WriteBits(uint64(bit), 1)
}

func (w *Writer) WriteBits(bits uint64, bitWidth uint) {
	for {
		w.cache |= (bits & ((1 << bitWidth) - 1)) << w.length
		n := 64 - w.length
		if n >= bitWidth {
			w.length += bitWidth
			break
		}
		w.length += n
		bits >>= n
		bitWidth -= n
		w.flush()
	}
}

func (w *Writer) WriteInt8x8(data [][8]int8, bitWidth uint) error {
	for i := range data {
		w.WriteBits(uint64(data[i][0]), bitWidth)
		w.WriteBits(uint64(data[i][1]), bitWidth)
		w.WriteBits(uint64(data[i][2]), bitWidth)
		w.WriteBits(uint64(data[i][3]), bitWidth)
		w.WriteBits(uint64(data[i][4]), bitWidth)
		w.WriteBits(uint64(data[i][5]), bitWidth)
		w.WriteBits(uint64(data[i][6]), bitWidth)
		w.WriteBits(uint64(data[i][7]), bitWidth)
	}
	return w.Flush()
}

func (w *Writer) WriteInt16x8(data [][8]int16, bitWidth uint) error {
	for i := range data {
		w.WriteBits(uint64(data[i][0]), bitWidth)
		w.WriteBits(uint64(data[i][1]), bitWidth)
		w.WriteBits(uint64(data[i][2]), bitWidth)
		w.WriteBits(uint64(data[i][3]), bitWidth)
		w.WriteBits(uint64(data[i][4]), bitWidth)
		w.WriteBits(uint64(data[i][5]), bitWidth)
		w.WriteBits(uint64(data[i][6]), bitWidth)
		w.WriteBits(uint64(data[i][7]), bitWidth)
	}
	return w.Flush()
}

func (w *Writer) WriteInt32x8(data [][8]int32, bitWidth uint) error {
	for i := range data {
		w.WriteBits(uint64(data[i][0]), bitWidth)
		w.WriteBits(uint64(data[i][1]), bitWidth)
		w.WriteBits(uint64(data[i][2]), bitWidth)
		w.WriteBits(uint64(data[i][3]), bitWidth)
		w.WriteBits(uint64(data[i][4]), bitWidth)
		w.WriteBits(uint64(data[i][5]), bitWidth)
		w.WriteBits(uint64(data[i][6]), bitWidth)
		w.WriteBits(uint64(data[i][7]), bitWidth)
	}
	return w.Flush()
}

func (w *Writer) WriteInt64x8(data [][8]int64, bitWidth uint) error {
	for i := range data {
		w.WriteBits(uint64(data[i][0]), bitWidth)
		w.WriteBits(uint64(data[i][1]), bitWidth)
		w.WriteBits(uint64(data[i][2]), bitWidth)
		w.WriteBits(uint64(data[i][3]), bitWidth)
		w.WriteBits(uint64(data[i][4]), bitWidth)
		w.WriteBits(uint64(data[i][5]), bitWidth)
		w.WriteBits(uint64(data[i][6]), bitWidth)
		w.WriteBits(uint64(data[i][7]), bitWidth)
	}
	return w.Flush()
}
