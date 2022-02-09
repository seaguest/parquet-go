package bits

import (
	"encoding/binary"
	"io"
)

type Reader struct {
	reader io.Reader
	length uint
	cache  uint64
	buffer [8]byte
}

func (r *Reader) Reset(rr io.Reader) {
	r.reader = rr
	r.length = 0
	r.cache = 0
}

func (r *Reader) ReadBit() (int, error) {
	bits, _, err := r.ReadBits(1)
	return int(bits), err
}

func (r *Reader) ReadBits(bitWidth uint) (uint64, uint, error) {
	bits, nbits := uint64(0), uint(0)

	for bitWidth > 0 {
		if r.length == 0 {
			n, err := r.reader.Read(r.buffer[:ByteCount(bitWidth)])
			if err != nil && n == 0 {
				if err == io.EOF && nbits != 0 {
					err = io.ErrUnexpectedEOF
				}
				return bits, nbits, err
			}
			b := [8]byte{}
			copy(b[:], r.buffer[:n])
			r.length = 8 * uint(n)
			r.cache = binary.LittleEndian.Uint64(b[:])
		}

		n := bitWidth
		if n > r.length {
			n = r.length
		}

		bits |= (r.cache & ((1 << n) - 1)) << nbits
		nbits += n
		bitWidth -= n
		r.length -= n
		r.cache >>= n
	}

	return bits, nbits, nil
}
