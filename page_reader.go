package parquet

import (
	"fmt"
	"io"

	"github.com/segmentio/parquet/deprecated"
	"github.com/segmentio/parquet/encoding"
	"github.com/segmentio/parquet/internal/bits"
)

// PageReader is an interface implemented by types that support reading values
// from pages of parquet files.
//
// The values read from the page do not have repetition or definition levels
// set, use a DataPageReader to decode values with levels.
type PageReader interface {
	// Returns the type of values read from the underlying page.
	Type() Type

	// Resets the decoder used to read values from the parquet page. This method
	// is useful to allow reusing readers. Calling this method drops all values
	// previously buffered by the reader.
	Reset(encoding.Decoder)

	ValueReader
}

// DataPageReader reads values from a data page.
//
// DataPageReader implements the ValueReader interface; when they exist,
// the reader decodes repetition and definition levels in order to assign
// levels to values returned to the application, which includes producing
// null values when needed.
type DataPageReader struct {
	page               PageReader
	remain             int
	maxRepetitionLevel int8
	maxDefinitionLevel int8
	columnIndex        int8
	repetitions        levelReader
	definitions        levelReader
}

func NewDataPageReader(repetitions, definitions encoding.Decoder, numValues int, page PageReader, maxRepetitionLevel, maxDefinitionLevel, columnIndex int8, bufferSize int) *DataPageReader {
	repetitionBufferSize := 0
	definitionBufferSize := 0

	switch {
	case maxRepetitionLevel > 0 && maxDefinitionLevel > 0:
		repetitionBufferSize = bufferSize / 2
		definitionBufferSize = bufferSize / 2

	case maxRepetitionLevel > 0:
		repetitionBufferSize = bufferSize

	case maxDefinitionLevel > 0:
		definitionBufferSize = bufferSize
	}

	repetitions.SetBitWidth(bits.Len8(maxRepetitionLevel))
	definitions.SetBitWidth(bits.Len8(maxDefinitionLevel))
	return &DataPageReader{
		page:               page,
		remain:             numValues,
		maxRepetitionLevel: maxRepetitionLevel,
		maxDefinitionLevel: maxDefinitionLevel,
		columnIndex:        ^columnIndex,
		repetitions:        makeLevelReader(repetitions, repetitionBufferSize),
		definitions:        makeLevelReader(definitions, definitionBufferSize),
	}
}

func (r *DataPageReader) Reset(repetitions, definitions encoding.Decoder, numValues int, page PageReader) {
	repetitions.SetBitWidth(bits.Len8(r.maxRepetitionLevel))
	definitions.SetBitWidth(bits.Len8(r.maxDefinitionLevel))
	r.page = page
	r.remain = numValues
	r.repetitions.reset(repetitions)
	r.definitions.reset(definitions)
}

func (r *DataPageReader) ReadValues(values []Value) (int, error) {
	read := 0

	for r.remain > 0 && len(values) > 0 {
		var err error
		var repetitionLevels []int8
		var definitionLevels []int8
		var numNulls int
		var numValues = r.remain

		if len(values) < numValues {
			numValues = len(values)
		}

		if r.maxRepetitionLevel > 0 {
			repetitionLevels, err = r.repetitions.peekLevels()
			if err != nil {
				return read, fmt.Errorf("reading parquet repetition level from data page: %w", err)
			}
			if len(repetitionLevels) < numValues {
				numValues = len(repetitionLevels)
			}
		}

		if r.maxDefinitionLevel > 0 {
			definitionLevels, err = r.definitions.peekLevels()
			if err != nil {
				return read, fmt.Errorf("reading parquet definition level from data page: %w", err)
			}
			if len(definitionLevels) < numValues {
				numValues = len(definitionLevels)
			}
		}

		if len(repetitionLevels) > 0 {
			repetitionLevels = repetitionLevels[:numValues]
		}
		if len(definitionLevels) > 0 {
			definitionLevels = definitionLevels[:numValues]
		}
		for _, d := range definitionLevels {
			if d != r.maxDefinitionLevel {
				numNulls++
			}
		}

		n, err := r.page.ReadValues(values[:numValues-numNulls])
		if err != nil {
			if err == io.EOF {
				// EOF should not happen at this stage since we successfully
				// decoded levels.
				err = io.ErrUnexpectedEOF
			}
			return read, fmt.Errorf("reading parquet values from data page: %w", err)
		}

		for i, j := n-1, len(definitionLevels)-1; j >= 0; j-- {
			if definitionLevels[j] != r.maxDefinitionLevel {
				values[j] = Value{}
			} else {
				values[j] = values[i]
				i--
			}
		}

		for i, lvl := range repetitionLevels {
			values[i].repetitionLevel = lvl
		}

		for i, lvl := range definitionLevels {
			values[i].definitionLevel = lvl
		}

		for i := range values[:numValues] {
			values[i].columnIndex = r.columnIndex
		}

		values = values[numValues:]
		r.repetitions.discardLevels(len(repetitionLevels))
		r.definitions.discardLevels(len(definitionLevels))
		r.remain -= numValues
		read += numValues
	}

	if r.remain == 0 && read == 0 {
		return 0, io.EOF
	}

	return read, nil
}

type levelReader struct {
	decoder encoding.Decoder
	levels  []int8
	offset  uint
	count   uint
}

func makeLevelReader(decoder encoding.Decoder, bufferSize int) levelReader {
	return levelReader{
		decoder: decoder,
		levels:  make([]int8, 0, bufferSize),
	}
}

func (r *levelReader) readLevel() (int8, error) {
	for {
		if r.offset < uint(len(r.levels)) {
			lvl := r.levels[r.offset]
			r.offset++
			return lvl, nil
		}
		if err := r.decodeLevels(); err != nil {
			return -1, err
		}
	}
}

func (r *levelReader) peekLevels() ([]int8, error) {
	if r.offset == uint(len(r.levels)) {
		if err := r.decodeLevels(); err != nil {
			return nil, err
		}
	}
	return r.levels[r.offset:], nil
}

func (r *levelReader) discardLevels(n int) {
	remain := uint(len(r.levels)) - r.offset
	switch {
	case uint(n) > remain:
		panic("BUG: cannot discard more levels than buffered")
	case uint(n) == remain:
		r.levels = r.levels[:0]
		r.offset = 0
	default:
		r.offset += uint(n)
	}
}

func (r *levelReader) decodeLevels() error {
	n, err := r.decoder.DecodeInt8(r.levels[:cap(r.levels)])
	if n == 0 {
		return err
	}
	r.levels = r.levels[:n]
	r.offset = 0
	r.count += uint(n)
	return nil
}

func (r *levelReader) reset(decoder encoding.Decoder) {
	r.decoder = decoder
	r.levels = r.levels[:0]
	r.offset = 0
	r.count = 0
}

type booleanPageReader struct {
	typ     Type
	decoder encoding.Decoder
	values  []bool
	offset  uint
}

func newBooleanPageReader(typ Type, decoder encoding.Decoder, bufferSize int) *booleanPageReader {
	return &booleanPageReader{
		typ:     typ,
		decoder: decoder,
		values:  make([]bool, 0, atLeastOne(bufferSize)),
	}
}

func (r *booleanPageReader) ReadValues(values []Value) (int, error) {
	i := 0
	for {
		for r.offset < uint(len(r.values)) && i < len(values) {
			values[i] = makeValueBoolean(r.values[r.offset])
			r.offset++
			i++
		}

		if i == len(values) {
			return i, nil
		}

		n, err := r.decoder.DecodeBoolean(r.values[:cap(r.values)])
		if n == 0 {
			return i, err
		}

		r.values = r.values[:n]
		r.offset = 0
	}
}

func (r *booleanPageReader) Reset(decoder encoding.Decoder) {
	r.decoder = decoder
	r.values = r.values[:0]
	r.offset = 0
}

func (r *booleanPageReader) Type() Type { return r.typ }

type int32PageReader struct {
	typ     Type
	decoder encoding.Decoder
	values  []int32
	offset  uint
}

func newInt32PageReader(typ Type, decoder encoding.Decoder, bufferSize int) *int32PageReader {
	return &int32PageReader{
		typ:     typ,
		decoder: decoder,
		values:  make([]int32, 0, atLeastOne(bufferSize/4)),
	}
}

func (r *int32PageReader) ReadValues(values []Value) (int, error) {
	i := 0
	for {
		for r.offset < uint(len(r.values)) && i < len(values) {
			values[i] = makeValueInt32(r.values[r.offset])
			r.offset++
			i++
		}

		if i == len(values) {
			return i, nil
		}

		n, err := r.decoder.DecodeInt32(r.values[:cap(r.values)])
		if n == 0 {
			return i, err
		}

		r.values = r.values[:n]
		r.offset = 0
	}
}

func (r *int32PageReader) Reset(decoder encoding.Decoder) {
	r.decoder = decoder
	r.values = r.values[:0]
	r.offset = 0
}

func (r *int32PageReader) Type() Type { return r.typ }

type int64PageReader struct {
	typ     Type
	decoder encoding.Decoder
	values  []int64
	offset  uint
}

func newInt64PageReader(typ Type, decoder encoding.Decoder, bufferSize int) *int64PageReader {
	return &int64PageReader{
		typ:     typ,
		decoder: decoder,
		values:  make([]int64, 0, atLeastOne(bufferSize/8)),
	}
}

func (r *int64PageReader) ReadValues(values []Value) (int, error) {
	i := 0
	for {
		for r.offset < uint(len(r.values)) && i < len(values) {
			values[i] = makeValueInt64(r.values[r.offset])
			r.offset++
			i++
		}

		if i == len(values) {
			return i, nil
		}

		n, err := r.decoder.DecodeInt64(r.values[:cap(r.values)])
		if n == 0 {
			return i, err
		}

		r.values = r.values[:n]
		r.offset = 0
	}
}

func (r *int64PageReader) Reset(decoder encoding.Decoder) {
	r.decoder = decoder
	r.values = r.values[:0]
	r.offset = 0
}

func (r *int64PageReader) Type() Type { return r.typ }

type int96PageReader struct {
	typ     Type
	decoder encoding.Decoder
	values  []deprecated.Int96
	offset  uint
}

func newInt96PageReader(typ Type, decoder encoding.Decoder, bufferSize int) *int96PageReader {
	return &int96PageReader{
		typ:     typ,
		decoder: decoder,
		values:  make([]deprecated.Int96, 0, atLeastOne(bufferSize/12)),
	}
}

func (r *int96PageReader) ReadValues(values []Value) (int, error) {
	i := 0
	for {
		for r.offset < uint(len(r.values)) && i < len(values) {
			values[i] = makeValueInt96(r.values[r.offset])
			r.offset++
			i++
		}

		if i == len(values) {
			return i, nil
		}

		n, err := r.decoder.DecodeInt96(r.values[:cap(r.values)])
		if n == 0 {
			return i, err
		}

		r.values = r.values[:n]
		r.offset = 0
	}
}

func (r *int96PageReader) Reset(decoder encoding.Decoder) {
	r.decoder = decoder
	r.values = r.values[:0]
	r.offset = 0
}

func (r *int96PageReader) Type() Type { return r.typ }

type floatPageReader struct {
	typ     Type
	decoder encoding.Decoder
	values  []float32
	offset  uint
}

func newFloatPageReader(typ Type, decoder encoding.Decoder, bufferSize int) *floatPageReader {
	return &floatPageReader{
		typ:     typ,
		decoder: decoder,
		values:  make([]float32, 0, atLeastOne(bufferSize/4)),
	}
}

func (r *floatPageReader) ReadValues(values []Value) (int, error) {
	i := 0
	for {
		for r.offset < uint(len(r.values)) && i < len(values) {
			values[i] = makeValueFloat(r.values[r.offset])
			r.offset++
			i++
		}

		if i == len(values) {
			return i, nil
		}

		n, err := r.decoder.DecodeFloat(r.values[:cap(r.values)])
		if n == 0 {
			return i, err
		}

		r.values = r.values[:n]
		r.offset = 0
	}
}

func (r *floatPageReader) Reset(decoder encoding.Decoder) {
	r.decoder = decoder
	r.values = r.values[:0]
	r.offset = 0
}

func (r *floatPageReader) Type() Type { return r.typ }

type doublePageReader struct {
	typ     Type
	decoder encoding.Decoder
	values  []float64
	offset  uint
}

func newDoublePageReader(typ Type, decoder encoding.Decoder, bufferSize int) *doublePageReader {
	return &doublePageReader{
		typ:     typ,
		decoder: decoder,
		values:  make([]float64, 0, atLeastOne(bufferSize/8)),
	}
}

func (r *doublePageReader) ReadValues(values []Value) (int, error) {
	i := 0
	for {
		for r.offset < uint(len(r.values)) && i < len(values) {
			values[i] = makeValueDouble(r.values[r.offset])
			r.offset++
			i++
		}

		if i == len(values) {
			return i, nil
		}

		n, err := r.decoder.DecodeDouble(r.values[:cap(r.values)])
		if n == 0 {
			return i, err
		}

		r.values = r.values[:n]
		r.offset = 0
	}
}

func (r *doublePageReader) Reset(decoder encoding.Decoder) {
	r.decoder = decoder
	r.values = r.values[:0]
	r.offset = 0
}

func (r *doublePageReader) Type() Type { return r.typ }

type byteArrayPageReader struct {
	typ     Type
	decoder encoding.Decoder
	values  encoding.ByteArrayList
	index   int
}

func newByteArrayPageReader(typ Type, decoder encoding.Decoder, bufferSize int) *byteArrayPageReader {
	return &byteArrayPageReader{
		typ:     typ,
		decoder: decoder,
		values:  encoding.MakeByteArrayList(atLeastOne(bufferSize / 16)),
	}
}

func (r *byteArrayPageReader) ReadValues(values []Value) (int, error) {
	i := 0
	for {
		for r.index < r.values.Len() && i < len(values) {
			values[i] = makeValueBytes(ByteArray, r.values.Index(r.index)).Clone()
			r.index++
			i++
		}

		if i == len(values) {
			return i, nil
		}

		r.values.Reset()
		n, err := r.decoder.DecodeByteArray(&r.values)
		if err != nil && n == 0 {
			return i, err
		}

		r.index = 0
	}
}

func (r *byteArrayPageReader) Reset(decoder encoding.Decoder) {
	r.decoder = decoder
	r.values.Reset()
	r.index = 0
}

func (r *byteArrayPageReader) Type() Type { return r.typ }

type fixedLenByteArrayPageReader struct {
	typ     Type
	decoder encoding.Decoder
	values  []byte
	offset  uint
	size    uint
}

func newFixedLenByteArrayPageReader(typ Type, decoder encoding.Decoder, bufferSize int) *fixedLenByteArrayPageReader {
	size := typ.Length()
	return &fixedLenByteArrayPageReader{
		typ:     typ,
		decoder: decoder,
		size:    uint(size),
		values:  make([]byte, 0, atLeast((bufferSize/size)*size, size)),
	}
}

func (r *fixedLenByteArrayPageReader) ReadValues(values []Value) (int, error) {
	i := 0
	for {
		for (r.offset+r.size) <= uint(len(r.values)) && i < len(values) {
			values[i] = makeValueBytes(FixedLenByteArray, copyBytes(r.values[r.offset:r.offset+r.size]))
			r.offset += r.size
			i++
		}

		if i == len(values) {
			return i, nil
		}

		n, err := r.decoder.DecodeFixedLenByteArray(int(r.size), r.values[:cap(r.values)])
		if n == 0 {
			return i, err
		}

		r.values = r.values[:uint(n)*r.size]
		r.offset = 0
	}
}

func (r *fixedLenByteArrayPageReader) Reset(decoder encoding.Decoder) {
	r.decoder = decoder
	r.values = r.values[:0]
	r.offset = 0
}

func (r *fixedLenByteArrayPageReader) Type() Type { return r.typ }

func atLeastOne(size int) int {
	return atLeast(size, 1)
}

func atLeast(size, least int) int {
	if size < least {
		return least
	}
	return size
}

var (
	_ ValueReader = (*DataPageReader)(nil)
	_ PageReader  = (*int32PageReader)(nil)
	_ PageReader  = (*int64PageReader)(nil)
	_ PageReader  = (*int96PageReader)(nil)
	_ PageReader  = (*floatPageReader)(nil)
	_ PageReader  = (*doublePageReader)(nil)
	_ PageReader  = (*byteArrayPageReader)(nil)
	_ PageReader  = (*fixedLenByteArrayPageReader)(nil)
)
