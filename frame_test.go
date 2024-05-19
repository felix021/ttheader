package ttheader

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"testing"
)

func TestFrame_WriteWithSize(t *testing.T) {
	t.Run("too-long", func(t *testing.T) {
		f := NewFrame(nil, nil)
		err := f.WriteWithSize(nil, 0x80000000, 0x80000000)
		assert(t, err != nil, err)
	})
	t.Run("insufficient-space-for-header", func(t *testing.T) {
		h := NewHeader()
		f := NewFrame(h, nil)
		headerSize, err := h.BytesLength()
		assert(t, err == nil, err)
		buf := make([]byte, 4+headerSize-1)
		err = f.WriteWithSize(buf, headerSize, 0)
		assert(t, err != nil, err)
	})
	t.Run("header:write-err", func(t *testing.T) {
		h := NewHeader()
		token := make([]byte, 65536)
		h.SetToken(string(token))
		f := NewFrame(h, nil)

		buf := make([]byte, 100)
		err := f.WriteWithSize(buf, 50, 0)
		assert(t, err != nil, err)
	})
	t.Run("normal", func(t *testing.T) {
		h := NewHeader()
		headerSize, err := h.BytesLength()
		assert(t, err == nil, err)

		payload := []byte{1, 2, 3, 4}
		payloadSize := len(payload)

		totalSize := 4 + headerSize + payloadSize
		buf := make([]byte, totalSize)

		f := NewFrame(h, payload)
		err = f.WriteWithSize(buf, headerSize, payloadSize)
		assert(t, err == nil, err)

		assert(t, binary.BigEndian.Uint32(buf) == uint32(totalSize)-4)

		headerBytes, err := h.Bytes()
		assert(t, err == nil, err)

		assert(t, reflect.DeepEqual(buf[4:4+headerSize], headerBytes), buf[4:4+headerSize])

		assert(t, reflect.DeepEqual(buf[4+headerSize:], payload), buf[4+headerSize:])
	})
}

func TestFrame_Bytes(t *testing.T) {
	t.Run("header-length:err", func(t *testing.T) {
		h := NewHeader()
		token := make([]byte, 65536)
		h.SetToken(string(token))
		f := NewFrame(h, nil)

		_, err := f.Bytes()
		assert(t, err != nil, err)
	})
	t.Run("normal", func(t *testing.T) {
		h := NewHeader()
		payload := []byte{1, 2, 3, 4}
		f := NewFrame(h, payload)

		buf, err := f.Bytes()
		assert(t, err == nil, err)

		headerBuf, err := h.Bytes()
		assert(t, err == nil, err)

		headerSize := len(headerBuf)
		payloadSize := len(payload)

		assert(t, binary.BigEndian.Uint32(buf) == uint32(headerSize+payloadSize))
		assert(t, reflect.DeepEqual(buf[4:4+headerSize], headerBuf), buf[4:4+headerSize])
		assert(t, reflect.DeepEqual(buf[4+headerSize:], payload), buf[4+headerSize:])
	})
}

func TestFrame_ReadWithSize(t *testing.T) {
	t.Run("read-header-err", func(t *testing.T) {
		buf := make([]byte, 10)
		binary.BigEndian.PutUint32(buf, 6)
		f := NewFrame(nil, nil)
		err := f.ReadWithSize(buf, 6)
		assert(t, err != nil, err)
	})
	t.Run("normal", func(t *testing.T) {
		h := NewHeader()
		h.SetSeqID(0x12345678)
		payload := []byte{1, 2, 3, 4}
		fw := NewFrame(h, payload)
		buf, err := fw.Bytes()
		assert(t, err == nil, err)

		fr := NewFrame(nil, nil)
		err = fr.ReadWithSize(buf[4:], len(buf)-4)
		assert(t, err == nil, err)
		assert(t, fr.Header().SeqID() == 0x12345678, fr.Header().SeqID())
		assert(t, reflect.DeepEqual(fr.Payload(), payload), fr.Payload())
	})
}

func TestFrame_Read(t *testing.T) {
	t.Run("invalid-frame:no-frame-size", func(t *testing.T) {
		buf := make([]byte, 3)
		_, err := ReadFrame(bytes.NewReader(buf))
		assert(t, err != nil, err)
	})
	t.Run("invalid-frame:no-frame-data", func(t *testing.T) {
		buf := make([]byte, 5)
		binary.BigEndian.PutUint32(buf, 10)
		_, err := ReadFrame(bytes.NewReader(buf))
		assert(t, err != nil, err)
	})
	t.Run("normal", func(t *testing.T) {
		h := NewHeader()
		h.SetToken("token")
		payload := []byte{1, 2, 3, 4}
		fw := NewFrame(h, payload)
		buf, err := fw.Bytes()
		assert(t, err == nil, err)

		fr, err := ReadFrame(bytes.NewReader(buf))
		assert(t, err == nil, err)
		assert(t, fr.Header().Token() == "token", fr.Header().Token())
		assert(t, reflect.DeepEqual(fr.Payload(), payload), fr.Payload())
	})
	t.Run("normal:read-into-header", func(t *testing.T) {
		h := NewHeader()
		h.SetToken("token")
		payload := []byte{1, 2, 3, 4}
		fw := NewFrame(h, payload)
		buf, err := fw.Bytes()
		assert(t, err == nil, err)

		hr := NewHeader()
		fr := NewFrame(hr, nil)
		err = fr.Read(bytes.NewReader(buf))
		assert(t, err == nil, err)
		assert(t, fr.Header().Token() == "token", fr.Header().Token())
		assert(t, hr.Token() == "token", fr.Header().Token()) // read into pre-allocated header obj
		assert(t, reflect.DeepEqual(fr.Payload(), payload), fr.Payload())
	})
}
