package ttheader

import (
	"encoding/binary"
	"errors"
	"io"
	"reflect"
	"testing"
)

func Test_writeUint16(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		buf := make([]byte, 2)
		err := writeUint16(buf, 1)
		if err != nil {
			t.Errorf("expect %v, got %v", nil, err)
		}
		if buf[0] != 0 || buf[1] != 1 {
			t.Errorf("expect %v, got %v", []byte{0, 1}, buf)
		}
	})
	t.Run("insufficient buffer", func(t *testing.T) {
		buf := make([]byte, 1)
		err := writeUint16(buf, 1)
		if !errors.Is(err, io.ErrShortWrite) {
			t.Errorf("expect %v, got %v", io.ErrShortWrite, err)
		}
	})
}

func Test_writeString(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		buf := make([]byte, 4)
		err := writeString(buf, "test")
		if err != nil {
			t.Errorf("expect %v, got %v", nil, err)
		}
		if string(buf) != "test" {
			t.Errorf("expect %v, got %v", "test", buf)
		}
	})
	t.Run("insufficient buffer", func(t *testing.T) {
		buf := make([]byte, 1)
		err := writeString(buf, "test")
		if !errors.Is(err, io.ErrShortWrite) {
			t.Errorf("expect %v, got %v", io.ErrShortWrite, err)
		}
	})
}

func Test_writeLengthPrefixedString(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		buf := make([]byte, 6)
		l, err := writeLengthPrefixedString(buf, "test")
		if err != nil {
			t.Errorf("expect %v, got %v", nil, err)
		}
		if l != 6 {
			t.Errorf("expect %v, got %v", 6, l)
		}
		if string(buf[2:]) != "test" {
			t.Errorf("expect %v, got %v", "test", buf[2:])
		}
		if buf[0] != 0 || buf[1] != 4 {
			t.Errorf("expect %v, got %v", []byte{0, 4}, buf[0:2])
		}
	})
	t.Run("insufficient buffer", func(t *testing.T) {
		buf := make([]byte, 1)
		_, err := writeLengthPrefixedString(buf, "test")
		if !errors.Is(err, io.ErrShortWrite) {
			t.Errorf("expect %v, got %v", io.ErrShortWrite, err)
		}
	})
}

func Test_writeStrKVInfo(t *testing.T) {
	t.Run("empty map", func(t *testing.T) {
		size := 0
		info := map[string]string{}
		buf := make([]byte, size)
		l, err := writeStrKVInfo(buf, info)
		assert(t, err == nil, err)
		assert(t, l == size, l)
	})

	t.Run("normal", func(t *testing.T) {
		info := map[string]string{"k1": "v1", "k2": "v2"}
		size := strInfoSize(info)

		buf := make([]byte, size)
		l, err := writeStrKVInfo(buf, info)
		assert(t, err == nil, err)
		assert(t, l == size, l)
		assert(t, buf[0] == InfoIDKeyValue, buf[0])

		strInfo, err := readStrKVInfo(newBytesReader(buf[1:]))
		assert(t, err == nil, err)
		assert(t, reflect.DeepEqual(strInfo, info), strInfo)
	})

	t.Run("short write info", func(t *testing.T) {
		k, v := "key", "value"
		info := map[string]string{k: v}
		buf := make([]byte, 0)
		_, err := writeStrKVInfo(buf, info)
		assert(t, errors.Is(err, io.ErrShortWrite), err)
	})

	t.Run("short write size", func(t *testing.T) {
		k, v := "key", "value"
		info := map[string]string{k: v}
		buf := make([]byte, 1)
		_, err := writeStrKVInfo(buf, info)
		assert(t, errors.Is(err, io.ErrShortWrite), err)
	})

	t.Run("short write key", func(t *testing.T) {
		k, v := "key", "value"
		info := map[string]string{k: v}
		buf := make([]byte, 3)
		_, err := writeStrKVInfo(buf, info)
		assert(t, errors.Is(err, io.ErrShortWrite), err)
	})

	t.Run("short write value", func(t *testing.T) {
		k, v := "key", "value"
		size := len(k) + 2 + len(v) + 2
		info := map[string]string{k: v}
		buf := make([]byte, size-1)
		_, err := writeStrKVInfo(buf, info)
		assert(t, errors.Is(err, io.ErrShortWrite), err)
	})
}

func Test_writeIntKVInfo(t *testing.T) {
	t.Run("empty map", func(t *testing.T) {
		size := 0
		info := map[uint16]string{}
		buf := make([]byte, size)
		l, err := writeIntKVInfo(buf, info)
		assert(t, err == nil, err)
		assert(t, l == size, l)
	})
	t.Run("normal", func(t *testing.T) {
		info := map[uint16]string{1: "a", 2: "b"}
		size := intInfoSize(info)
		buf := make([]byte, size)
		l, err := writeIntKVInfo(buf, info)
		assert(t, err == nil, err)
		assert(t, l == size, l)
		assert(t, buf[0] == InfoIDIntKeyValue, buf[0])

		intInfo, err := readIntKVInfo(newBytesReader(buf[1:]))
		assert(t, err == nil, err)
		assert(t, reflect.DeepEqual(intInfo, info), intInfo)
	})
	t.Run("short write info", func(t *testing.T) {
		k, v := uint16(1), "test"
		info := map[uint16]string{k: v}
		buf := make([]byte, 0)
		_, err := writeIntKVInfo(buf, info)
		assert(t, errors.Is(err, io.ErrShortWrite), err)
	})
	t.Run("short write size", func(t *testing.T) {
		k, v := uint16(1), "test"
		info := map[uint16]string{k: v}
		buf := make([]byte, 1)
		_, err := writeIntKVInfo(buf, info)
		assert(t, errors.Is(err, io.ErrShortWrite), err)
	})
	t.Run("short write key", func(t *testing.T) {
		k, v := uint16(1), "test"
		info := map[uint16]string{k: v}
		buf := make([]byte, 3)
		_, err := writeIntKVInfo(buf, info)
		assert(t, errors.Is(err, io.ErrShortWrite), err)
	})
	t.Run("short write value", func(t *testing.T) {
		k, v := uint16(1), "test"
		info := map[uint16]string{k: v}
		size := 2 + 2 + len(v)
		buf := make([]byte, size-1)
		_, err := writeIntKVInfo(buf, info)
		assert(t, errors.Is(err, io.ErrShortWrite), err)
	})
}

func Test_byteSliceToString(t *testing.T) {
	expected := "test"
	buf := []byte(expected)
	str := byteSliceToString(buf)
	if str != expected {
		t.Errorf("expect %v, got %v", expected, str)
	}
}

func Test_stringToByteSlice(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		str := "test"
		buf := stringToByteSlice(str)
		assert(t, string(buf) == str)
	})
}

func Test_readString(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		expected := "test"
		buf := []byte(expected)
		reader := newBytesReader(buf)
		str, err := readString(reader, 4)
		if err != nil {
			t.Errorf("expect %v, got %v", nil, err)
			return
		}
		if str != expected {
			t.Errorf("expect %v, got %v", expected, str)
		}
	})
	t.Run("length-exceed-eof", func(t *testing.T) {
		buf := []byte("test")
		reader := newBytesReader(buf)
		_, err := readString(reader, 5)
		if err != io.EOF {
			t.Errorf("expect %v, got %v", io.EOF, err)
			return
		}
	})
	t.Run("reader-eof", func(t *testing.T) {
		buf := []byte("")
		reader := newBytesReader(buf)
		_, err := readString(reader, 4)
		if err != io.EOF {
			t.Errorf("expect %v, got %v", io.EOF, err)
		}
	})
}

func Test_readUint16(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		expected := uint16(1)
		buf := make([]byte, 2)
		binary.BigEndian.PutUint16(buf, expected)
		reader := newBytesReader(buf)
		n, err := readUint16(reader)
		if err != nil {
			t.Errorf("expect %v, got %v", nil, err)
			return
		}
		if n != expected {
			t.Errorf("expect %v, got %v", expected, n)
		}
	})
	t.Run("length-exceed-eof", func(t *testing.T) {
		reader := newBytesReader([]byte{0})
		_, err := readUint16(reader)
		if err != io.EOF {
			t.Errorf("expect %v, got %v", io.EOF, err)
			return
		}
	})
	t.Run("reader-eof", func(t *testing.T) {
		reader := newBytesReader([]byte{})
		_, err := readUint16(reader)
		if err != io.EOF {
			t.Errorf("expect %v, got %v", io.EOF, err)
		}
	})
}

func Test_readLengthPrefixedString(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		expected := "test"
		buf := make([]byte, 6)
		_, err := writeLengthPrefixedString(buf, expected)
		if err != nil {
			t.Errorf("write expect %v, got %v", nil, err)
			return
		}
		reader := newBytesReader(buf)
		str, err := readLengthPrefixedString(reader)
		if err != nil {
			t.Errorf("read expect %v, got %v", nil, err)
			return
		}
		if str != expected {
			t.Errorf("expect %v, got %v", expected, str)
		}
	})

	t.Run("read-length-err", func(t *testing.T) {
		reader := newBytesReader(nil)
		_, err := readLengthPrefixedString(reader)
		if err == nil {
			t.Errorf("read expect %v, got %v", nil, err)
			return
		}
	})

	t.Run("read-string-err", func(t *testing.T) {
		expected := "test"
		buf := make([]byte, 6)
		_, err := writeLengthPrefixedString(buf, expected)
		if err != nil {
			t.Errorf("write expect %v, got %v", nil, err)
			return
		}
		reader := newBytesReader(buf[:5])
		_, err = readLengthPrefixedString(reader)
		if err == nil {
			t.Errorf("read expect %v, got %v", nil, err)
			return
		}
	})
}

func Test_readStrKVInfo(t *testing.T) {
	expected := map[string]string{"k1": "v1", "k2": "v2"}
	size := strInfoSize(expected)
	buf := make([]byte, size)
	_, writeErr := writeStrKVInfo(buf, expected)
	assert(t, writeErr == nil, writeErr)

	t.Run("normal", func(t *testing.T) {
		reader := newBytesReader(buf[1:])
		info, err := readStrKVInfo(reader)
		assert(t, err == nil, err)
		assert(t, reflect.DeepEqual(info, expected), info)
	})

	t.Run("read-size-err", func(t *testing.T) {
		reader := newBytesReader(nil)
		_, err := readStrKVInfo(reader)
		if err == nil {
			t.Errorf("expect %v, got %v", nil, err)
		}
	})

	t.Run("read-key-err", func(t *testing.T) {
		buf := []byte{0, 2}
		reader := newBytesReader(buf)
		_, err := readStrKVInfo(reader)
		assert(t, err != nil, err)
	})

	t.Run("read-value-err", func(t *testing.T) {
		reader := newBytesReader(buf[:len(buf)-2])
		_, err := readStrKVInfo(reader)
		assert(t, err != nil, err)
	})
}

func Test_readIntKVInfo(t *testing.T) {
	expected := map[uint16]string{1: "a", 2: "b"}
	size := intInfoSize(expected)
	buf := make([]byte, size)
	_, writeErr := writeIntKVInfo(buf, expected)
	if writeErr != nil {
		t.Errorf("expect %v, got %v", nil, writeErr)
	}

	t.Run("normal", func(t *testing.T) {
		reader := newBytesReader(buf[1:])
		info, err := readIntKVInfo(reader)
		assert(t, err == nil, err)
		assert(t, reflect.DeepEqual(info, expected), info)
	})

	t.Run("read-size-err", func(t *testing.T) {
		reader := newBytesReader(nil)
		_, err := readIntKVInfo(reader)
		assert(t, err != nil, err)
	})

	t.Run("read-key-err", func(t *testing.T) {
		reader := newBytesReader([]byte{0, 2})
		_, err := readIntKVInfo(reader)
		assert(t, err != nil, err)
	})

	t.Run("read-value-err", func(t *testing.T) {
		reader := newBytesReader(buf[:len(buf)-2])
		_, err := readIntKVInfo(reader)
		assert(t, err != nil, err)
	})
}

func Test_writeByte(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		expected := byte(1)
		buf := make([]byte, 1)
		err := writeByte(buf, expected)
		if err != nil {
			t.Errorf("expect %v, got %v", nil, err)
			return
		}
		if buf[0] != expected {
			t.Errorf("expect %v, got %v", expected, buf[0])
		}
	})
	t.Run("write-err", func(t *testing.T) {
		expected := byte(1)
		buf := make([]byte, 0)
		err := writeByte(buf, expected)
		if err == nil {
			t.Errorf("expect non-nil, got %v", err)
			return
		}
	})
}

func Test_paddingSize(t *testing.T) {
	t.Run("size-0", func(t *testing.T) {
		size := paddingSize(0, PaddingSize)
		assert(t, size == 0, size)
	})
	t.Run("remain-1", func(t *testing.T) {
		size := paddingSize(1, PaddingSize)
		assert(t, size == 3, size)
	})
	t.Run("remain-2", func(t *testing.T) {
		size := paddingSize(2, PaddingSize)
		assert(t, size == 2, size)
	})
	t.Run("remain-3", func(t *testing.T) {
		size := paddingSize(3, PaddingSize)
		assert(t, size == 1, size)
	})
	t.Run("remain-0", func(t *testing.T) {
		size := paddingSize(4, PaddingSize)
		assert(t, size == 0, size)
	})
}

func Test_strInfoSize(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		expected := 0
		size := strInfoSize(nil)
		assert(t, size == expected, size)
	})
	t.Run("zero-sized", func(t *testing.T) {
		expected := 0
		size := strInfoSize(map[string]string{})
		assert(t, size == expected, size)
	})
	t.Run("normal", func(t *testing.T) {
		expected := 1 + 2 + 2 + 2 + 2 + 2 + 2 + 2 + 2 + 2
		size := strInfoSize(map[string]string{"k1": "v1", "k2": "v2"})
		assert(t, size == expected, size)
	})
}

func Test_intInfoSize(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		expected := 0
		size := intInfoSize(nil)
		assert(t, size == expected, size)
	})
	t.Run("zero-sized", func(t *testing.T) {
		expected := 0
		size := intInfoSize(map[uint16]string{})
		assert(t, size == expected, size)
	})
	t.Run("normal", func(t *testing.T) {
		expected := 1 + 2 + 2 + 2 + 1 + 2 + 2 + 1
		size := intInfoSize(map[uint16]string{1: "a", 2: "b"})
		assert(t, size == expected, size)
	})
}

func Test_tokenSize(t *testing.T) {
	t.Run("empty-token", func(t *testing.T) {
		expected := 0
		size := tokenSize("")
		assert(t, size == expected, size)
	})
	t.Run("normal", func(t *testing.T) {
		expected := 1 + 2 + 4
		size := tokenSize("test")
		assert(t, size == expected, size)
	})
}

func Test_writeToken(t *testing.T) {
	t.Run("empty-token", func(t *testing.T) {
		buf := make([]byte, 0)
		n, err := writeToken(buf, "")
		assert(t, err == nil)
		assert(t, n == 0)
	})

	t.Run("info_id:short-write", func(t *testing.T) {
		token := "test"
		buf := make([]byte, 0)
		_, err := writeToken(buf, token)
		assert(t, err != nil)
	})

	t.Run("token:short-write", func(t *testing.T) {
		token := "test"
		buf := make([]byte, 1)
		_, err := writeToken(buf, token)
		assert(t, err != nil)
		assert(t, buf[0] == InfoIDACLToken, buf[0])
	})

	t.Run("normal", func(t *testing.T) {
		token := "test"
		buf := make([]byte, tokenSize(token))
		n, err := writeToken(buf, token)
		assert(t, err == nil)
		assert(t, n == tokenSize(token))
		assert(t, buf[0] == InfoIDACLToken, buf[0])
		assert(t, binary.BigEndian.Uint16(buf[1:3]) == uint16(len(token)), binary.BigEndian.Uint16(buf[1:3]))
		assert(t, string(buf[3:]) == token, string(buf[3:]))
	})
}

func TestIsMagic(t *testing.T) {
	t.Run("normal:true", func(t *testing.T) {
		buf := []byte{0x10, 0x0}
		assert(t, IsMagic(buf))
	})
	t.Run("normal:false", func(t *testing.T) {
		buf := []byte{0x11, 0x0}
		assert(t, !IsMagic(buf))
	})
	t.Run("short-read", func(t *testing.T) {
		buf := []byte{0x10}
		assert(t, !IsMagic(buf))
	})
}

func TestIsStreaming(t *testing.T) {
	t.Run("normal:true", func(t *testing.T) {
		flags := BitMaskIsStreaming
		buf := []byte{0, 0}
		binary.BigEndian.PutUint16(buf, flags)
		assert(t, IsStreaming(buf))
	})
	t.Run("normal:false", func(t *testing.T) {
		buf := []byte{0x0, 0x0}
		assert(t, !IsStreaming(buf))
	})
	t.Run("short-read", func(t *testing.T) {
		buf := []byte{0x0}
		assert(t, !IsStreaming(buf))
	})
}

func TestParsePackageServiceMethod(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		p, s, m, err := ParsePackageServiceMethod("$package.$service/$method")
		assert(t, err == nil, err)
		assert(t, p == "$package", p)
		assert(t, s == "$service", s)
		assert(t, m == "$method", m)
	})
	t.Run("normal:dotted-pkg", func(t *testing.T) {
		p, s, m, err := ParsePackageServiceMethod("a.b.c.$service/$method")
		assert(t, err == nil, err)
		assert(t, p == "a.b.c", p)
		assert(t, s == "$service", s)
		assert(t, m == "$method", m)
	})
	t.Run("normal:slash-prefixed", func(t *testing.T) {
		p, s, m, err := ParsePackageServiceMethod("/$package.$service/$method")
		assert(t, err == nil, err)
		assert(t, p == "$package", p)
		assert(t, s == "$service", s)
		assert(t, m == "$method", m)
	})
	t.Run("normal:no-package", func(t *testing.T) {
		p, s, m, err := ParsePackageServiceMethod("$service/$method")
		assert(t, err == nil, err)
		assert(t, p == "", p)
		assert(t, s == "$service", s)
		assert(t, m == "$method", m)
	})
	t.Run("error:no-service", func(t *testing.T) {
		_, _, _, err := ParsePackageServiceMethod("$method")
		assert(t, err != nil, err)
	})
	t.Run("error:empty-string", func(t *testing.T) {
		_, _, _, err := ParsePackageServiceMethod("")
		assert(t, err != nil, err)
	})
}
