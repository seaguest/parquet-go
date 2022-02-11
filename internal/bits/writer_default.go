//go:build purego || !amd64

package bits

import "unsafe"

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

func writeInt16x8(dst *[512]byte, src [][8]int16, bitWidth uint) int {
	m := int16(1<<bitWidth) - 1
	n := 512 / bitWidth
	if uint(len(src)) > n {
		src = src[:n:n]
	}

	b := unsafe.Slice(*(**int16)(unsafe.Pointer(&dst)), len(dst)/2)
	for i := range b {
		b[i] = 0
	}
	_ = b[:(bitWidth*uint(len(src)))/2]

	k := uint(0)
	c := int16(0)
	for i := range src {
		for _, v := range src[i] {
			q, r := k/16, k%16
			bits := v & m
			b[q] |= (bits << r) | c
			c = bits >> (16 - r)
			k += bitWidth
		}
	}
	if c != 0 {
		b[k/16] |= c
	}
	return len(src)
}

func writeInt32x8(dst *[512]byte, src [][8]int32, bitWidth uint) int {
	m := int32(1<<bitWidth) - 1
	n := 512 / bitWidth
	if uint(len(src)) > n {
		src = src[:n:n]
	}

	b := unsafe.Slice(*(**int32)(unsafe.Pointer(&dst)), len(dst)/4)
	for i := range b {
		b[i] = 0
	}
	_ = b[:(bitWidth*uint(len(src)))/4]

	k := uint(0)
	c := int32(0)
	for i := range src {
		for _, v := range src[i] {
			q, r := k/32, k%32
			bits := v & m
			b[q] |= (bits << r) | c
			c = bits >> (32 - r)
			k += bitWidth
		}
	}
	if c != 0 {
		b[k/32] |= c
	}
	return len(src)
}

func writeInt64x8(dst *[512]byte, src [][8]int64, bitWidth uint) int {
	m := int64(1<<bitWidth) - 1
	n := 512 / bitWidth
	if uint(len(src)) > n {
		src = src[:n:n]
	}

	b := unsafe.Slice(*(**int64)(unsafe.Pointer(&dst)), len(dst)/8)
	for i := range b {
		b[i] = 0
	}
	_ = b[:(bitWidth*uint(len(src)))/8]

	k := uint(0)
	c := int64(0)
	for i := range src {
		for _, v := range src[i] {
			q, r := k/64, k%64
			bits := v & m
			b[q] |= (bits << r) | c
			c = bits >> (64 - r)
			k += bitWidth
		}
	}
	if c != 0 {
		b[k/64] |= c
	}
	return len(src)
}
