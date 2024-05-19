package ttheader

import (
	"encoding/binary"
	"io"
	"reflect"
	"testing"
)

func TestNewBufReader(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		buf := []byte{1, 2, 3}
		br := newBytesReader(buf)
		assert(t, reflect.DeepEqual(br.buf, buf))
		assert(t, br.len == 3)
		assert(t, br.idx == 0)
	})
}

func TestBytesReader_ReadByte(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		buf := []byte{1, 2}
		br := newBytesReader(buf)

		b, err := br.ReadByte()
		assert(t, err == nil)
		assert(t, b == 1)

		b, err = br.ReadByte()
		assert(t, err == nil)
		assert(t, b == 2)

		b, err = br.ReadByte()
		assert(t, err == io.EOF)
		assert(t, b == 0)
	})

	t.Run("eof", func(t *testing.T) {
		buf := []byte{}
		br := newBytesReader(buf)

		b, err := br.ReadByte()
		assert(t, err == io.EOF)
		assert(t, b == 0)
	})
}

func TestBytesReader_ReadBytes(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		buf := []byte{1, 2, 3}
		br := newBytesReader(buf)

		b, err := br.ReadBytes(2)
		assert(t, err == nil)
		assert(t, b[0] == 1)
		assert(t, b[1] == 2)
	})

	t.Run("read-all", func(t *testing.T) {
		buf := []byte{1, 2, 3}
		br := newBytesReader(buf)

		b, err := br.ReadBytes(3)
		assert(t, err == nil)
		assert(t, b[0] == 1)
		assert(t, b[1] == 2)
		assert(t, b[2] == 3)
	})

	t.Run("read-more-than-rest", func(t *testing.T) {
		buf := []byte{1, 2, 3}
		br := newBytesReader(buf)

		b, err := br.ReadBytes(4)
		assert(t, err == io.EOF)
		assert(t, len(b) == 3)
		assert(t, b[0] == 1)
		assert(t, b[1] == 2)
		assert(t, b[2] == 3)
	})

	t.Run("eof-and-read", func(t *testing.T) {
		buf := []byte{1, 2, 3}
		br := newBytesReader(buf)

		b, err := br.ReadBytes(4)
		assert(t, err == io.EOF)
		assert(t, len(b) == 3)

		b, err = br.ReadBytes(1)
		assert(t, err == io.EOF)
		assert(t, b == nil)
	})
}

func Test_bytesReader_ReadString(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		buf := []byte{'0', '1', '2', '3'}
		br := newBytesReader(buf)
		s, err := br.ReadString(4)
		assert(t, err == nil)
		assert(t, s == "0123")
	})
	t.Run("short-read", func(t *testing.T) {
		buf := []byte{'0', '1', '2', '3'}
		br := newBytesReader(buf)
		_, err := br.ReadString(5)
		assert(t, err == io.EOF)
	})
}

func Test_bytesReader_ReadInt32(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, 0xffffffff)
		br := newBytesReader(buf)
		i, err := br.ReadInt32()
		assert(t, err == nil)
		assert(t, i == -1)
	})

	t.Run("eof", func(t *testing.T) {
		buf := []byte{1, 2, 3}
		br := newBytesReader(buf)
		_, err := br.ReadInt32()
		assert(t, err == io.EOF)
	})
}

func Test_bytesReader_ReadUint32(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, 0x7fffffff)
		br := newBytesReader(buf)
		i, err := br.ReadUint32()
		assert(t, err == nil)
		assert(t, i == 0x7fffffff)
	})
	t.Run("eof", func(t *testing.T) {
		buf := []byte{1, 2, 3}
		br := newBytesReader(buf)
		_, err := br.ReadUint32()
		assert(t, err == io.EOF)
	})
}

func Test_bytesReader_ReadUint16(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		buf := make([]byte, 2)
		binary.BigEndian.PutUint16(buf, 0x7fff)
		br := newBytesReader(buf)
		i, err := br.ReadUint16()
		assert(t, err == nil)
		assert(t, i == 0x7fff)
	})
	t.Run("eof", func(t *testing.T) {
		buf := []byte{1}
		br := newBytesReader(buf)
		_, err := br.ReadUint16()
		assert(t, err == io.EOF)
	})
}
