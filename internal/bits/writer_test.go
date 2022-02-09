package bits_test

import (
	"bytes"
	"testing"

	"github.com/segmentio/parquet-go/internal/bits"
)

func TestWriter(t *testing.T) {
	b := new(bytes.Buffer)
	w := new(bits.Writer)
	w.Reset(b)

	for i := 0; i < 123; i++ {
		w.WriteBit(i & 1)
	}

	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}

	data := b.Bytes()
	want := []byte{
		0b10101010, 0b10101010, 0b10101010, 0b10101010,
		0b10101010, 0b10101010, 0b10101010, 0b10101010,

		0b10101010, 0b10101010, 0b10101010, 0b10101010,
		0b10101010, 0b10101010, 0b10101010, 0b00000010,
	}

	if !bytes.Equal(data, want) {
		t.Errorf("data = %08b", data)
		t.Errorf("want = %08b", want)
	}
}
