package bits_test

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/segmentio/parquet-go/internal/bits"
)

func TestWriter(t *testing.T) {
	t.Run("WriteInt8x8", testWriterWriteInt8x8)
	t.Run("WriteInt16x8", testWriterWriteInt16x8)
	t.Run("WriteInt32x8", testWriterWriteInt32x8)
	t.Run("WriteInt64x8", testWriterWriteInt64x8)
}

var testWriterSizes = sizes{
	1, 2, 3, 4, 5, 6, 7, 8, 9,
	10, 11, 12, 13, 14, 15, 16, 17, 18, 19,
	20, 21, 22, 23, 24, 25, 26, 27, 28, 29,
	30, 31, 32, 33, 34, 35, 36, 37, 38, 39,
	99, 100, 101,
	127, 128, 129,
	255, 256, 257,
}

func testWriterWriteInt8x8(t *testing.T) {
	for bitWidth := uint(1); bitWidth <= 8; bitWidth++ {
		t.Run(fmt.Sprintf("bitWidth=%d", bitWidth), func(t *testing.T) {
			b := new(bytes.Buffer)
			r := new(bits.Reader)
			w := new(bits.Writer)

			err := testWriterSizes.quickCheck(func(values [][8]int8) error {
				r.Reset(b)
				w.Reset(b)

				if err := w.WriteInt8x8(values, bitWidth); err != nil {
					return err
				}

				for i := range values {
					for _, v := range values[i] {
						if err := testReadBits(r, uint64(v), bitWidth); err != nil {
							return err
						}
					}
				}

				return nil
			})
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func testWriterWriteInt16x8(t *testing.T) {
	for bitWidth := uint(1); bitWidth <= 16; bitWidth++ {
		t.Run(fmt.Sprintf("bitWidth=%d", bitWidth), func(t *testing.T) {
			b := new(bytes.Buffer)
			r := new(bits.Reader)
			w := new(bits.Writer)

			err := testWriterSizes.quickCheck(func(values [][8]int16) error {
				r.Reset(b)
				w.Reset(b)

				if err := w.WriteInt16x8(values, bitWidth); err != nil {
					return err
				}

				for i := range values {
					for _, v := range values[i] {
						if err := testReadBits(r, uint64(v), bitWidth); err != nil {
							return err
						}
					}
				}

				return nil
			})
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func testWriterWriteInt32x8(t *testing.T) {
	for bitWidth := uint(1); bitWidth <= 32; bitWidth++ {
		t.Run(fmt.Sprintf("bitWidth=%d", bitWidth), func(t *testing.T) {
			b := new(bytes.Buffer)
			r := new(bits.Reader)
			w := new(bits.Writer)

			err := testWriterSizes.quickCheck(func(values [][8]int32) error {
				r.Reset(b)
				w.Reset(b)

				if err := w.WriteInt32x8(values, bitWidth); err != nil {
					return err
				}

				for i := range values {
					for _, v := range values[i] {
						if err := testReadBits(r, uint64(v), bitWidth); err != nil {
							return err
						}
					}
				}

				return nil
			})
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func testWriterWriteInt64x8(t *testing.T) {
	for bitWidth := uint(1); bitWidth <= 64; bitWidth++ {
		t.Run(fmt.Sprintf("bitWidth=%d", bitWidth), func(t *testing.T) {
			b := new(bytes.Buffer)
			r := new(bits.Reader)
			w := new(bits.Writer)

			err := testWriterSizes.quickCheck(func(values [][8]int64) error {
				r.Reset(b)
				w.Reset(b)

				if err := w.WriteInt64x8(values, bitWidth); err != nil {
					return err
				}

				for i := range values {
					for _, v := range values[i] {
						if err := testReadBits(r, uint64(v), bitWidth); err != nil {
							return err
						}
					}
				}

				return nil
			})
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func testReadBits(r *bits.Reader, value uint64, bitWidth uint) error {
	mask := uint64(1<<bitWidth) - 1
	value &= mask
	v, n, err := r.ReadBits(bitWidth)
	if err != nil {
		return fmt.Errorf("expected %d bits value but got error: %v", bitWidth, err)
	}
	if n != bitWidth {
		return fmt.Errorf("expected %d bits value but got %d bits", bitWidth, n)
	}
	if v != value {
		return fmt.Errorf("expected value 0x%016x but got 0x%016x", value, v)
	}
	return nil
}

func BenchmarkWriter(b *testing.B) {
	b.Run("WriteInt8x8", benchmarkWriterWriteInt8x8)
	b.Run("WriteInt16x8", benchmarkWriterWriteInt16x8)
	b.Run("WriteInt32x8", benchmarkWriterWriteInt32x8)
	b.Run("WriteInt64x8", benchmarkWriterWriteInt64x8)
}

func benchmarkWriterWriteInt8x8(b *testing.B) {
	w := new(bits.Writer)
	w.Reset(io.Discard)

	forEachBenchmarkBufferSize(b, func(b *testing.B, bufferSize int) {
		values := make([][8]int8, bufferSize/8)

		for i := 0; i < b.N; i++ {
			w.WriteInt8x8(values, 3)
		}
	})
}

func benchmarkWriterWriteInt16x8(b *testing.B) {
	w := new(bits.Writer)
	w.Reset(io.Discard)

	forEachBenchmarkBufferSize(b, func(b *testing.B, bufferSize int) {
		values := make([][8]int16, bufferSize/16)

		for i := 0; i < b.N; i++ {
			w.WriteInt16x8(values, 3)
		}
	})
}

func benchmarkWriterWriteInt32x8(b *testing.B) {
	w := new(bits.Writer)
	w.Reset(io.Discard)

	forEachBenchmarkBufferSize(b, func(b *testing.B, bufferSize int) {
		values := make([][8]int32, bufferSize/32)

		for i := 0; i < b.N; i++ {
			w.WriteInt32x8(values, 3)
		}
	})
}

func benchmarkWriterWriteInt64x8(b *testing.B) {
	w := new(bits.Writer)
	w.Reset(io.Discard)

	forEachBenchmarkBufferSize(b, func(b *testing.B, bufferSize int) {
		values := make([][8]int64, bufferSize/64)

		for i := 0; i < b.N; i++ {
			w.WriteInt64x8(values, 3)
		}
	})
}
