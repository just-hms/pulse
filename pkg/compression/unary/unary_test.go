package unary_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"testing"

	"github.com/just-hms/pulse/pkg/compression/unary"
	"github.com/stretchr/testify/require"
)

func TestWriter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		values []uint64
		exp    []byte
	}{
		{"Single Value", []uint64{5}, []byte{0b1111_1011}},
		{"Multiple Values", []uint64{1, 2, 3}, []byte{0b10110111, 0b01111111}},
		{"Large Value", []uint64{9}, []byte{0b11111111, 0b10111111}},
		{"Empty Input", []uint64{}, []byte{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := require.New(t)
			// Setup a f as the underlying writer
			f := &bytes.Buffer{}
			buf := make([]byte, 8)

			w := unary.NewWriter(f, 0)

			// Write the test values
			for _, value := range tt.values {
				binary.LittleEndian.PutUint64(buf, value)
				_, err := w.Write(buf)
				req.NoError(err)
			}

			// Flush the writer to ensure all data is written
			err := w.Flush()
			req.NoError(err)

			got := f.Bytes()
			if got == nil {
				got = []byte{}
			}
			req.Equal(tt.exp, got)
		})
	}
}

func TestWriterBytesCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		values []uint64
	}{
		{"Single Value", []uint64{5}},
		{"Multiple Values", []uint64{1, 2, 3}},
		{"Large Value", []uint64{20}},
		{"Empty Input", []uint64{}},
		{"Already aligned", []uint64{12, 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := require.New(t)
			// Setup a f as the underlying writer
			f := &bytes.Buffer{}
			buf := make([]byte, 8)

			w := unary.NewWriter(f, 0)

			sum := 0
			for _, v := range tt.values {
				binary.LittleEndian.PutUint64(buf, v)
				n, err := w.Write(buf)
				req.NoError(err)
				sum += n
			}
			err := w.Flush()
			req.NoError(err)

			req.Equal(len(f.Bytes()), sum)
		})
	}
}

func TestReader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data []byte
		exp  []uint64
	}{
		{"Single Value", []byte{0b11111011}, []uint64{5}},
		{"Multiple Values", []byte{0b10110111, 0b01111111}, []uint64{1, 2, 3}},
		{"Large Value", []byte{0b11111111, 0b10111111}, []uint64{9}},
		{"Empty Input", []byte{}, []uint64{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := require.New(t)

			// Setup a f as the underlying reader
			f := &bytes.Buffer{}
			buf := make([]byte, 8)

			_, err := f.Write(tt.data)
			req.NoError(err)

			r := unary.NewReader(f, 0)

			got := []uint64{}

			for {
				_, err := r.Read(buf)
				if errors.Is(err, io.EOF) {
					break
				}
				req.NoError(err)

				v := binary.LittleEndian.Uint64(buf)
				got = append(got, v)

			}
			req.Equal(tt.exp, got)
		})
	}
}

func TestDifferenSize(t *testing.T) {
	t.Parallel()
	req := require.New(t)

	f := &bytes.Buffer{}
	buf := make([]byte, 4)

	w := unary.NewWriter(f, 0)
	exp := uint32(12)
	binary.LittleEndian.PutUint32(buf, exp)
	_, err := w.Write(buf)
	req.NoError(err)

	err = w.Flush()
	req.NoError(err)

	req.Equal([]byte{0b11111111, 0b11110111}, f.Bytes())

	r := unary.NewReader(f, 0)
	_, err = r.Read(buf)
	req.NoError(err)

	got := binary.LittleEndian.Uint32(buf)
	req.Equal(exp, got)
}

func TestMiniumum(t *testing.T) {
	t.Parallel()
	req := require.New(t)

	// test with 0 as minimum
	{
		f := &bytes.Buffer{}
		buf := make([]byte, 8)

		exp := uint64(9)
		w := unary.NewWriter(f, 0)
		binary.LittleEndian.PutUint64(buf, exp)
		n, err := w.Write(buf)
		req.NoError(err)
		req.Equal(2, n)

		err = w.Flush()
		req.NoError(err)

		req.Equal([]byte{0b11111111, 0b10111111}, f.Bytes())

		r := unary.NewReader(f, 0)
		r.Read(buf)
		v := binary.LittleEndian.Uint64(buf)
		req.Equal(exp, v)
	}

	// test with 1 as minimum
	{
		f := &bytes.Buffer{}
		buf := make([]byte, 8)

		exp := uint64(9)
		w := unary.NewWriter(f, 1)
		binary.LittleEndian.PutUint64(buf, exp)
		n, err := w.Write(buf)
		req.NoError(err)
		req.Equal(2, n)

		err = w.Flush()
		req.NoError(err)

		req.Equal([]byte{0b11111111, 0b01111111}, f.Bytes())

		r := unary.NewReader(f, 1)
		r.Read(buf)
		v := binary.LittleEndian.Uint64(buf)
		req.Equal(exp, v)
	}

}
