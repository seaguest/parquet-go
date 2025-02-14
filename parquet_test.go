package parquet_test

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"reflect"

	"github.com/segmentio/parquet-go"
	"github.com/segmentio/parquet-go/deprecated"
)

type Contact struct {
	Name        string `parquet:"name"`
	PhoneNumber string `parquet:"phoneNumber,optional,zstd"`
}

type AddressBook struct {
	Owner             string    `parquet:"owner,zstd"`
	OwnerPhoneNumbers []string  `parquet:"ownerPhoneNumbers,gzip"`
	Contacts          []Contact `parquet:"contacts"`
}

func forEachLeafColumn(col *parquet.Column, do func(*parquet.Column) error) error {
	children := col.Columns()

	if len(children) == 0 {
		return do(col)
	}

	for _, child := range children {
		if err := forEachLeafColumn(child, do); err != nil {
			return err
		}
	}

	return nil
}

func forEachPage(pages parquet.PageReader, do func(parquet.Page) error) error {
	for {
		p, err := pages.ReadPage()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return err
		}
		if err := do(p); err != nil {
			return err
		}
	}
}

func forEachValue(values parquet.ValueReader, do func(parquet.Value) error) error {
	buffer := [3]parquet.Value{}
	for {
		n, err := values.ReadValues(buffer[:])
		for _, v := range buffer[:n] {
			if err := do(v); err != nil {
				return err
			}
		}
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return err
		}
	}
}

func forEachColumnPage(col *parquet.Column, do func(*parquet.Column, parquet.Page) error) error {
	return forEachLeafColumn(col, func(leaf *parquet.Column) error {
		return forEachPage(leaf.Pages(), func(page parquet.Page) error { return do(leaf, page) })
	})
}

func forEachColumnValue(col *parquet.Column, do func(*parquet.Column, parquet.Value) error) error {
	return forEachColumnPage(col, func(leaf *parquet.Column, page parquet.Page) error {
		return forEachValue(page.Values(), func(value parquet.Value) error { return do(leaf, value) })
	})
}

func forEachColumnChunk(file *parquet.File, do func(*parquet.Column, parquet.ColumnChunk) error) error {
	return forEachLeafColumn(file.Root(), func(leaf *parquet.Column) error {
		for i, n := 0, file.NumRowGroups(); i < n; i++ {
			if err := do(leaf, file.RowGroup(i).Column(int(leaf.Index()))); err != nil {
				return err
			}
		}
		return nil
	})
}

func createParquetFile(rows rows, options ...parquet.WriterOption) (*parquet.File, error) {
	buffer := new(bytes.Buffer)

	if err := writeParquetFile(buffer, rows, options...); err != nil {
		return nil, err
	}

	reader := bytes.NewReader(buffer.Bytes())
	return parquet.OpenFile(reader, reader.Size())
}

func writeParquetFile(w io.Writer, rows rows, options ...parquet.WriterOption) error {
	writer := parquet.NewWriter(w, options...)

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return writer.Close()
}

func writeParquetFileWithBuffer(w io.Writer, rows rows, options ...parquet.WriterOption) error {
	buffer := parquet.NewBuffer()
	for _, row := range rows {
		if err := buffer.Write(row); err != nil {
			return err
		}
	}

	writer := parquet.NewWriter(w, options...)
	numRows, err := parquet.CopyRows(writer, buffer.Rows())
	if err != nil {
		return err
	}
	if numRows != int64(len(rows)) {
		return fmt.Errorf("wrong number of rows written from buffer to file: want=%d got=%d", len(rows), numRows)
	}
	return writer.Close()
}

type rows []interface{}

func makeRows(any interface{}) rows {
	if v, ok := any.([]interface{}); ok {
		return rows(v)
	}
	value := reflect.ValueOf(any)
	slice := make([]interface{}, value.Len())
	for i := range slice {
		slice[i] = value.Index(i).Interface()
	}
	return rows(slice)
}

func randValueFuncOf(t parquet.Type) func(*rand.Rand) parquet.Value {
	switch k := t.Kind(); k {
	case parquet.Boolean:
		return func(r *rand.Rand) parquet.Value {
			return parquet.ValueOf(r.Float64() < 0.5)
		}

	case parquet.Int32:
		return func(r *rand.Rand) parquet.Value {
			return parquet.ValueOf(r.Int31())
		}

	case parquet.Int64:
		return func(r *rand.Rand) parquet.Value {
			return parquet.ValueOf(r.Int63())
		}

	case parquet.Int96:
		return func(r *rand.Rand) parquet.Value {
			return parquet.ValueOf(deprecated.Int96{
				0: r.Uint32(),
				1: r.Uint32(),
				2: r.Uint32(),
			})
		}

	case parquet.Float:
		return func(r *rand.Rand) parquet.Value {
			return parquet.ValueOf(r.Float32())
		}

	case parquet.Double:
		return func(r *rand.Rand) parquet.Value {
			return parquet.ValueOf(r.Float64())
		}

	case parquet.ByteArray:
		return func(r *rand.Rand) parquet.Value {
			n := r.Intn(49) + 1
			b := make([]byte, n)
			return parquet.ValueOf(b)
		}

	default:
		panic("NOT IMPLEMENTED")
	}
}
