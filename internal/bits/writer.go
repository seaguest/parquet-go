package bits

import (
	"fmt"
	"io"
	"unsafe"
)

type Writer struct {
	writer io.Writer
	buffer [512]byte
}

func (w *Writer) Reset(ww io.Writer) {
	w.writer = ww
}

func (w *Writer) WriteInt8x8(data [][8]int8, bitWidth uint) error {
	if bitWidth > 8 {
		return fmt.Errorf("cannot write 8 bits values to %d bits", bitWidth)
	}
	buf := &w.buffer

	for i := 0; i < len(data); {
		n := writeInt8x8(buf, data[i:len(data):len(data)], bitWidth)
		k := uint(n) * bitWidth

		if _, err := w.writer.Write(buf[:k]); err != nil {
			return err
		}

		i += n
	}

	return nil
}

func (w *Writer) WriteInt16x8(data [][8]int16, bitWidth uint) error {
	if bitWidth > 16 {
		return fmt.Errorf("cannot write 16 bits values to %d bits", bitWidth)
	}
	buf := (*[64]byte)(w.buffer[:])

	for i := range data {
		*buf = [64]byte{}
		writeInt16x8(buf, &data[i], bitWidth)

		if _, err := w.writer.Write(buf[:bitWidth]); err != nil {
			return err
		}
	}

	return nil
}

func (w *Writer) WriteInt32x8(data [][8]int32, bitWidth uint) error {
	if bitWidth > 32 {
		return fmt.Errorf("cannot write 32 bits values to %d bits", bitWidth)
	}
	buf := (*[64]byte)(w.buffer[:])

	for i := range data {
		*buf = [64]byte{}
		writeInt32x8(buf, &data[i], bitWidth)

		if _, err := w.writer.Write(buf[:bitWidth]); err != nil {
			return err
		}
	}

	return nil
}

func (w *Writer) WriteInt64x8(data [][8]int64, bitWidth uint) error {
	if bitWidth > 64 {
		return fmt.Errorf("cannot write 64 bits values to %d bits", bitWidth)
	}
	buf := (*[64]byte)(w.buffer[:])

	for i := range data {
		*buf = [64]byte{}
		writeInt64x8(buf, &data[i], bitWidth)

		if _, err := w.writer.Write(buf[:bitWidth]); err != nil {
			return err
		}
	}

	return nil
}

func writeInt8x8(dst *[512]byte, src [][8]int8, bitWidth uint) int {
	m := uint64(1<<bitWidth) - 1
	n := 512 / bitWidth

	if uint(len(src)) > n {
		src = src[:n:n]
	}

	b := dst[:bitWidth*uint(len(src))]
	for i := range b {
		b[i] = 0
	}

	for i := range src {
		p := (*uint64)(unsafe.Pointer(&b[uint(i)*bitWidth]))
		*p |= (uint64(src[i][0])&m)<<(0*bitWidth) |
			(uint64(src[i][1])&m)<<(1*bitWidth) |
			(uint64(src[i][2])&m)<<(2*bitWidth) |
			(uint64(src[i][3])&m)<<(3*bitWidth) |
			(uint64(src[i][4])&m)<<(4*bitWidth) |
			(uint64(src[i][5])&m)<<(5*bitWidth) |
			(uint64(src[i][6])&m)<<(6*bitWidth) |
			(uint64(src[i][7])&m)<<(7*bitWidth)
	}

	return len(src)
}

func writeInt16x8(dst *[64]byte, src *[8]int16, bitWidth uint) {
	mask := uint64(1<<bitWidth) - 1
	bits := [8]uint64{
		0: uint64(src[0]) & mask,
		1: uint64(src[1]) & mask,
		2: uint64(src[2]) & mask,
		3: uint64(src[3]) & mask,
		4: uint64(src[4]) & mask,
		5: uint64(src[5]) & mask,
		6: uint64(src[6]) & mask,
		7: uint64(src[7]) & mask,
	}
	writeUint64x8(dst, &bits, bitWidth)
}

func writeInt32x8(dst *[64]byte, src *[8]int32, bitWidth uint) {
	mask := uint64(1<<bitWidth) - 1
	bits := [8]uint64{
		0: uint64(src[0]) & mask,
		1: uint64(src[1]) & mask,
		2: uint64(src[2]) & mask,
		3: uint64(src[3]) & mask,
		4: uint64(src[4]) & mask,
		5: uint64(src[5]) & mask,
		6: uint64(src[6]) & mask,
		7: uint64(src[7]) & mask,
	}
	writeUint64x8(dst, &bits, bitWidth)
}

func writeInt64x8(dst *[64]byte, src *[8]int64, bitWidth uint) {
	mask := uint64(1<<bitWidth) - 1
	bits := [8]uint64{
		0: uint64(src[0]) & mask,
		1: uint64(src[1]) & mask,
		2: uint64(src[2]) & mask,
		3: uint64(src[3]) & mask,
		4: uint64(src[4]) & mask,
		5: uint64(src[5]) & mask,
		6: uint64(src[6]) & mask,
		7: uint64(src[7]) & mask,
	}
	writeUint64x8(dst, &bits, bitWidth)
}

func writeUint64x8(dst *[64]byte, src *[8]uint64, bitWidth uint) {
	b := (*[8]uint64)(unsafe.Pointer(dst))
	c := uint64(0)
	i := uint(0)
	j := uint(0)
	k := uint(0)

	i, j = k/64, k%64
	b[i] |= src[0] << j

	k += bitWidth
	i, j = k/64, k%64
	b[i] |= src[1] << j

	c = src[1] >> (64 - j)
	k += bitWidth
	i, j = k/64, k%64
	b[i] |= (src[2] << j) | c

	c = src[2] >> (64 - j)
	k += bitWidth
	i, j = k/64, k%64
	b[i] |= (src[3] << j) | c

	c = src[3] >> (64 - j)
	k += bitWidth
	i, j = k/64, k%64
	b[i] |= (src[4] << j) | c

	c = src[4] >> (64 - j)
	k += bitWidth
	i, j = k/64, k%64
	b[i] |= (src[5] << j) | c

	c = src[5] >> (64 - j)
	k += bitWidth
	i, j = k/64, k%64
	b[i] |= (src[6] << j) | c

	c = src[6] >> (64 - j)
	k += bitWidth
	i, j = k/64, k%64
	b[i] |= (src[7] << j) | c

	if c = src[7] >> (64 - j); c != 0 {
		b[(k+bitWidth)/64] |= c
	}
}
