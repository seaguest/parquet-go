package bits_test

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"github.com/segmentio/parquet-go/internal/bits"
)

func TestNearestPowerOfTwo(t *testing.T) {
	for _, test := range []struct {
		input  uint32
		output uint32
	}{
		{input: 0, output: 0},
		{input: 1, output: 1},
		{input: 2, output: 2},
		{input: 3, output: 4},
		{input: 4, output: 4},
		{input: 5, output: 8},
		{input: 6, output: 8},
		{input: 7, output: 8},
		{input: 8, output: 8},
		{input: 30, output: 32},
	} {
		t.Run(fmt.Sprintf("NearestPowerOfTwo(%d)", test.input), func(t *testing.T) {
			if nextPow2 := bits.NearestPowerOfTwo32(test.input); nextPow2 != test.output {
				t.Errorf("wrong 32 bits value: want=%d got=%d", test.output, nextPow2)
			}
			if nextPow2 := bits.NearestPowerOfTwo64(uint64(test.input)); nextPow2 != uint64(test.output) {
				t.Errorf("wrong 64 bits value: want=%d got=%d", test.output, nextPow2)
			}
		})
	}
}

func TestBitCount(t *testing.T) {
	for _, test := range []struct {
		bytes int
		bits  uint
	}{
		{bytes: 0, bits: 0},
		{bytes: 1, bits: 8},
		{bytes: 2, bits: 16},
		{bytes: 3, bits: 24},
		{bytes: 4, bits: 32},
		{bytes: 5, bits: 40},
		{bytes: 6, bits: 48},
	} {
		t.Run(fmt.Sprintf("BitCount(%d)", test.bytes), func(t *testing.T) {
			if bits := bits.BitCount(test.bytes); bits != test.bits {
				t.Errorf("wrong bit count: want=%d got=%d", test.bits, bits)
			}
		})
	}
}

func TestByteCount(t *testing.T) {
	for _, test := range []struct {
		bits  uint
		bytes int
	}{
		{bits: 0, bytes: 0},
		{bits: 1, bytes: 1},
		{bits: 7, bytes: 1},
		{bits: 8, bytes: 1},
		{bits: 9, bytes: 2},
		{bits: 30, bytes: 4},
		{bits: 63, bytes: 8},
	} {
		t.Run(fmt.Sprintf("ByteCount(%d)", test.bits), func(t *testing.T) {
			if bytes := bits.ByteCount(test.bits); bytes != test.bytes {
				t.Errorf("wrong byte count: want=%d got=%d", test.bytes, bytes)
			}
		})
	}
}

func TestRound(t *testing.T) {
	for _, test := range []struct {
		bits  uint
		round uint
	}{
		{bits: 0, round: 0},
		{bits: 1, round: 8},
		{bits: 8, round: 8},
		{bits: 9, round: 16},
		{bits: 30, round: 32},
		{bits: 63, round: 64},
	} {
		t.Run(fmt.Sprintf("Round(%d)", test.bits), func(t *testing.T) {
			if round := bits.Round(test.bits); round != test.round {
				t.Errorf("wrong rounded bit count: want=%d got=%d", test.round, round)
			}
		})
	}
}

var benchmarkBufferSizes = [...]int{
	4 * 1024,
	256 * 1024,
	2048 * 1024,
}

func forEachBenchmarkBufferSize(b *testing.B, f func(*testing.B, int)) {
	for _, bufferSize := range benchmarkBufferSizes {
		b.Run(fmt.Sprintf("%dKiB", bufferSize/1024), func(b *testing.B) {
			b.SetBytes(int64(bufferSize))
			f(b, bufferSize)
		})
	}
}

// quickCheck is inspired by the standard quick.Check package, but enhances the
// API and tests arrays of larger sizes than the maximum of 50 hardcoded in
// testing/quick.
func quickCheck(f interface{}) error {
	return sizes{
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9,
		10, 11, 12, 13, 14, 15, 16, 17, 18, 19,
		20, 21, 22, 23, 24, 25, 26, 27, 28, 29,
		30, 31, 32, 33, 34, 35, 36, 37, 38, 39,
		99, 100, 101,
		127, 128, 129,
		255, 256, 257,
		1000, 1023, 1024, 1025,
		2000, 2095, 2048, 2049,
		4000, 4095, 4096, 4097,
	}.quickCheck(f)
}

type sizes []int

func (sizes sizes) quickCheck(f interface{}) error {
	makeArray := makeArrayFuncOf(reflect.TypeOf(f).In(0))
	v := reflect.ValueOf(f)
	r := rand.New(rand.NewSource(0))

	for _, n := range sizes {
		in := makeArray(r, n)
		rv := v.Call([]reflect.Value{in})
		switch ret := rv[0].Interface().(type) {
		case bool:
			if !ret {
				return fmt.Errorf("failed on input of size %d: %#v\n", n, in)
			}
		case error:
			if ret != nil {
				return fmt.Errorf("failed on input of size %d: %v\n", n, ret)
			}
		case nil:
			// OK!
		default:
			panic(fmt.Sprintf("quick check function returned value of unsupported type: %T", ret))
		}
	}

	return nil
}

func makeArrayFuncOf(t reflect.Type) func(*rand.Rand, int) reflect.Value {
	var makeArray func(*rand.Rand, int) reflect.Value

	switch t.Kind() {
	case reflect.Bool:
		makeArray = func(r *rand.Rand, _ int) reflect.Value {
			return reflect.ValueOf(r.Int()%3 != 0)
		}

	case reflect.Int8:
		makeArray = func(r *rand.Rand, _ int) reflect.Value {
			return reflect.ValueOf(int8(r.Uint32() & 0x7F))
		}

	case reflect.Int16:
		makeArray = func(r *rand.Rand, _ int) reflect.Value {
			return reflect.ValueOf(int16(r.Uint32() & 0x7FFF))
		}

	case reflect.Int32:
		makeArray = func(r *rand.Rand, _ int) reflect.Value {
			return reflect.ValueOf(r.Int31())
		}

	case reflect.Int64:
		makeArray = func(r *rand.Rand, _ int) reflect.Value {
			return reflect.ValueOf(r.Int63())
		}

	case reflect.Uint8:
		makeArray = func(r *rand.Rand, _ int) reflect.Value {
			return reflect.ValueOf(uint8(r.Uint32()))
		}

	case reflect.Uint16:
		makeArray = func(r *rand.Rand, _ int) reflect.Value {
			return reflect.ValueOf(uint16(r.Uint32()))
		}

	case reflect.Uint32:
		makeArray = func(r *rand.Rand, _ int) reflect.Value {
			return reflect.ValueOf(r.Uint32())
		}

	case reflect.Uint64:
		makeArray = func(r *rand.Rand, _ int) reflect.Value {
			return reflect.ValueOf(r.Uint64())
		}

	case reflect.Float32:
		makeArray = func(r *rand.Rand, _ int) reflect.Value {
			return reflect.ValueOf(r.Float32())
		}

	case reflect.Float64:
		makeArray = func(r *rand.Rand, _ int) reflect.Value {
			return reflect.ValueOf(r.Float64())
		}

	case reflect.Array:
		f := makeArrayFuncOf(t.Elem())
		n := t.Len()
		makeArray = func(r *rand.Rand, _ int) reflect.Value {
			v := reflect.New(t).Elem()
			for i := 0; i < n; i++ {
				v.Index(i).Set(f(r, 0))
			}
			return v
		}

	case reflect.Slice:
		f := makeArrayFuncOf(t.Elem())
		makeArray = func(r *rand.Rand, n int) reflect.Value {
			v := reflect.MakeSlice(t, n, n)
			for i := 0; i < n; i++ {
				v.Index(i).Set(f(r, 0))
			}
			return v
		}
	}

	if makeArray == nil {
		panic("cannot run quick check on function with input of type " + t.String())
	}

	return makeArray
}
