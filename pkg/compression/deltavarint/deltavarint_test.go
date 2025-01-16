package deltavarint_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"testing"

	"github.com/just-hms/pulse/pkg/compression/deltavarint"
	"github.com/stretchr/testify/require"
)

func TestWriter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		values []uint64
		exp    []byte
	}{
		{"Single Value", []uint64{5}, []byte{5}},
		{"Multiple Values", []uint64{1, 2, 3}, []byte{1, 1, 1}},
		{"Bigger multiple Values", []uint64{100, 102, 103}, []byte{100, 2, 1}},
		{
			"Big values",
			[]uint64{300, 500, 590},
			[]byte{
				0b10101100, // 172
				0b00000010, // 2
				0b11001000, // 200
				0b00000001, // 1
				0b01011010, // 90
			}},
		{"Empty Input", []uint64{}, []byte{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := require.New(t)
			// Setup a f as the underlying writer
			f := &bytes.Buffer{}
			buf := make([]byte, 8)

			w := deltavarint.NewWriter(f)

			// Write the test values
			for _, value := range tt.values {
				binary.LittleEndian.PutUint64(buf, value)
				_, err := w.Write(buf)
				req.NoError(err)
			}

			got := f.Bytes()
			if got == nil {
				got = []byte{}
			}
			req.Equal(tt.exp, got)
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
		{"Single Value", []byte{5}, []uint64{5}},
		{"Multiple Values", []byte{1, 1, 1}, []uint64{1, 2, 3}},
		{"Bigger multiple Values", []byte{100, 2, 1}, []uint64{100, 102, 103}},
		{"Big values", []byte{172, 2, 200, 1, 90}, []uint64{300, 500, 590}},
		{"Empty Input", []byte{}, []uint64{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := require.New(t)
			// Setup a f as the underlying writer
			f := &bytes.Buffer{}
			buf := make([]byte, 8)

			_, err := f.Write(tt.data)
			req.NoError(err)

			w := deltavarint.NewReader(f)

			got := []uint64{}
			// Write the test values
			for {
				_, err := w.Read(buf)
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

	w := deltavarint.NewWriter(f)
	exp := uint32(12)
	binary.LittleEndian.PutUint32(buf, exp)
	_, err := w.Write(buf)
	req.NoError(err)

	r := deltavarint.NewReader(f)
	_, err = r.Read(buf)
	req.NoError(err)

	got := binary.LittleEndian.Uint32(buf)
	req.Equal(exp, got)
}
