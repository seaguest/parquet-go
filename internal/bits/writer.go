package bits

import (
	"fmt"
	"io"
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
	buf := &w.buffer

	for i := 0; i < len(data); {
		n := writeInt16x8(buf, data[i:len(data):len(data)], bitWidth)
		k := uint(n) * bitWidth

		if _, err := w.writer.Write(buf[:k]); err != nil {
			return err
		}

		i += n
	}

	return nil
}

func (w *Writer) WriteInt32x8(data [][8]int32, bitWidth uint) error {
	if bitWidth > 32 {
		return fmt.Errorf("cannot write 32 bits values to %d bits", bitWidth)
	}
	buf := &w.buffer

	for i := 0; i < len(data); {
		n := writeInt32x8(buf, data[i:len(data):len(data)], bitWidth)
		k := uint(n) * bitWidth

		if _, err := w.writer.Write(buf[:k]); err != nil {
			return err
		}

		i += n
	}

	return nil
}

func (w *Writer) WriteInt64x8(data [][8]int64, bitWidth uint) error {
	if bitWidth > 64 {
		return fmt.Errorf("cannot write 64 bits values to %d bits", bitWidth)
	}
	buf := &w.buffer

	for i := 0; i < len(data); {
		n := writeInt64x8(buf, data[i:len(data):len(data)], bitWidth)
		k := uint(n) * bitWidth

		if _, err := w.writer.Write(buf[:k]); err != nil {
			return err
		}

		i += n
	}

	return nil
}
