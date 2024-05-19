package ttheader

import (
	"encoding/binary"
	"errors"
	"reflect"
	"testing"
)

func TestNewHeaderWithInfo(t *testing.T) {
	intInfo := map[uint16]string{1: "a", 2: "b"}
	strInfo := map[string]string{"c": "d"}
	h := NewHeaderWithInfo(intInfo, strInfo)
	assert(t, reflect.DeepEqual(h.IntInfo(), intInfo))
	assert(t, reflect.DeepEqual(h.StrInfo(), strInfo))
}

func TestNewHeader(t *testing.T) {
	h := NewHeader()
	assert(t, reflect.DeepEqual(h, &Header{}))
}

func TestHeader_Size(t *testing.T) {
	h := NewHeader()
	assert(t, h.Size() == 0)
	h.SetSize(10)
	assert(t, h.Size() == 10)

	defer func() {
		if err := recover(); err == nil {
			t.Errorf("should panic")
		} else {
			t.Logf("expected panic")
		}
	}()
	h.SetSize(uint16max + 1)
}

func TestHeader_Flags(t *testing.T) {
	h := NewHeader()
	assert(t, h.Flags() == 0)
	h.SetFlags(10)
	assert(t, h.Flags() == 10)
}

func TestHeader_SeqID(t *testing.T) {
	h := NewHeader()
	assert(t, h.SeqID() == 0)
	h.SetSeqID(10)
	assert(t, h.SeqID() == 10)
}

func TestHeader_ProtocolID(t *testing.T) {
	h := NewHeader()
	assert(t, h.ProtocolID() == 0)
	h.SetProtocolID(10)
	assert(t, h.ProtocolID() == 10)
}

func TestHeader_Token(t *testing.T) {
	h := NewHeader()
	assert(t, h.Token() == "")
	h.SetToken("test")
	assert(t, h.Token() == "test")
}

func TestHeader_IntInfo(t *testing.T) {
	h := NewHeader()
	assert(t, h.IntInfo() == nil)
	intInfo := map[uint16]string{1: "a", 2: "b"}
	h.SetIntInfo(intInfo)
	assert(t, reflect.DeepEqual(h.IntInfo(), intInfo))
}

func TestHeader_StrInfo(t *testing.T) {
	h := NewHeader()
	assert(t, h.StrInfo() == nil)
	strInfo := map[string]string{"c": "d"}
	h.SetStrInfo(strInfo)
	assert(t, reflect.DeepEqual(h.StrInfo(), strInfo))
}

func TestHeader_SetIsStreaming(t *testing.T) {
	h := NewHeader()
	assert(t, h.IsStreaming() == false)
	h.SetIsStreaming()
	assert(t, h.IsStreaming() == true)
}

func TestHeader_writeInfo(t *testing.T) {
	t.Run("int:short-write", func(t *testing.T) {
		intInfo := map[uint16]string{1: "a"}
		strInfo := map[string]string{"b": "c"}
		h := NewHeaderWithInfo(intInfo, strInfo)
		h.SetToken("token")
		buf := make([]byte, 0)
		err := h.writeInfo(buf)
		assert(t, err != nil)
	})

	t.Run("str:short-write", func(t *testing.T) {
		intInfo := map[uint16]string{1: "a"}
		strInfo := map[string]string{"b": "c"}
		h := NewHeaderWithInfo(intInfo, strInfo)
		h.SetToken("token")
		buf := make([]byte, intInfoSize(intInfo))
		err := h.writeInfo(buf)
		assert(t, err != nil)
		assert(t, buf[0] == InfoIDIntKeyValue, buf[0])
	})

	t.Run("token:short-write", func(t *testing.T) {
		intInfo := map[uint16]string{1: "a"}
		strInfo := map[string]string{"b": "c"}
		h := NewHeaderWithInfo(intInfo, strInfo)
		h.SetToken("token")
		buf := make([]byte, intInfoSize(intInfo)+strInfoSize(strInfo))
		err := h.writeInfo(buf)
		assert(t, err != nil)
		assert(t, buf[0] == InfoIDIntKeyValue, buf[0])
		assert(t, buf[intInfoSize(intInfo)] == InfoIDKeyValue, buf[intInfoSize(intInfo)])
	})

	t.Run("int+str+token", func(t *testing.T) {
		intInfo := map[uint16]string{1: "a"}
		strInfo := map[string]string{"b": "c"}
		h := NewHeaderWithInfo(intInfo, strInfo)
		h.SetToken("token")

		size := intInfoSize(intInfo) + strInfoSize(strInfo) + tokenSize("token")
		buf := make([]byte, size)
		err := h.writeInfo(buf)
		assert(t, err == nil)

		assert(t, buf[0] == InfoIDIntKeyValue, buf[0])
		assert(t, buf[intInfoSize(intInfo)] == InfoIDKeyValue, buf[intInfoSize(intInfo)])
		assert(t, buf[intInfoSize(intInfo)+strInfoSize(strInfo)] == InfoIDACLToken,
			buf[intInfoSize(intInfo)+strInfoSize(strInfo)])
	})

	t.Run("padding", func(t *testing.T) {
		h := NewHeader()
		size := 2
		buf := []byte{1, 2}
		err := h.writeInfo(buf[:size])
		assert(t, err == nil, err)
		assert(t, buf[0] == 0, buf[0])
		assert(t, buf[1] == 0, buf[0])
	})

	t.Run("token+padding", func(t *testing.T) {
		h := NewHeader()
		h.SetToken("token")
		size := tokenSize(h.Token()) + 2
		buf := make([]byte, size)
		buf[size-2] = 1
		buf[size-1] = 2

		err := h.writeInfo(buf[:size])

		assert(t, err == nil, err)
		assert(t, buf[size-1] == 0, buf[0])
		assert(t, buf[size-2] == 0, buf[0])
	})
}

func TestHeader_readInfo(t *testing.T) {
	intInfo := map[uint16]string{1: "a"}
	strInfo := map[string]string{"b": "c"}
	token := "token"

	t.Run("int+str+token", func(t *testing.T) {
		hw := NewHeaderWithInfo(intInfo, strInfo)
		hw.SetToken(token)
		buf := make([]byte, intInfoSize(intInfo)+strInfoSize(strInfo)+tokenSize(token))
		err := hw.writeInfo(buf)
		assert(t, err == nil)

		h := NewHeader()
		err = h.readInfo(newBytesReader(buf))

		assert(t, err == nil)
		assert(t, h.IntInfo()[1] == "a")
		assert(t, h.StrInfo()["b"] == "c")
		assert(t, h.Token() == token)
	})

	t.Run("int+str", func(t *testing.T) {
		hw := NewHeaderWithInfo(intInfo, strInfo)
		buf := make([]byte, intInfoSize(intInfo)+strInfoSize(strInfo))
		err := hw.writeInfo(buf)
		assert(t, err == nil)

		h := NewHeader()
		err = h.readInfo(newBytesReader(buf))

		assert(t, err == nil)
		assert(t, h.IntInfo()[1] == "a")
		assert(t, h.StrInfo()["b"] == "c")
		assert(t, h.token == "")
	})

	t.Run("int+token", func(t *testing.T) {
		hw := NewHeaderWithInfo(intInfo, nil)
		hw.SetToken(token)
		buf := make([]byte, intInfoSize(intInfo)+tokenSize(token))
		err := hw.writeInfo(buf)
		assert(t, err == nil)

		h := NewHeader()
		err = h.readInfo(newBytesReader(buf))

		assert(t, err == nil)
		assert(t, h.IntInfo()[1] == "a")
		assert(t, h.token == token)
		assert(t, len(h.StrInfo()) == 0)
	})

	t.Run("str+token", func(t *testing.T) {
		hw := NewHeaderWithInfo(nil, strInfo)
		hw.SetToken(token)
		buf := make([]byte, strInfoSize(strInfo)+tokenSize(token))
		err := hw.writeInfo(buf)
		assert(t, err == nil)

		h := NewHeader()
		err = h.readInfo(newBytesReader(buf))

		assert(t, err == nil)
		assert(t, len(h.IntInfo()) == 0)
		assert(t, h.token == token)
		assert(t, h.StrInfo()["b"] == "c")
	})

	t.Run("short-read", func(t *testing.T) {
		buf := make([]byte, 1)
		buf[0] = InfoIDIntKeyValue

		h := NewHeader()
		err := h.readInfo(newBytesReader(buf))

		assert(t, err != nil)
	})

	t.Run("padding", func(t *testing.T) {
		buf := make([]byte, 1)
		buf[0] = InfoIDPadding

		h := NewHeader()
		err := h.readInfo(newBytesReader(buf))

		assert(t, err == nil)
		assert(t, h.token == "")
		assert(t, len(h.IntInfo()) == 0)
		assert(t, len(h.StrInfo()) == 0)
	})

	t.Run("invalid-info-id", func(t *testing.T) {
		buf := make([]byte, 1)
		buf[0] = 0xff

		h := NewHeader()
		err := h.readInfo(newBytesReader(buf))

		assert(t, err != nil)
	})
}

func TestHeader_WriteWithSize(t *testing.T) {
	t.Run("fixed", func(t *testing.T) {
		h := NewHeader()
		h.SetIsStreaming()
		h.SetSeqID(0x12345678)
		h.SetProtocolID(0x1)

		bufSize, err := h.BytesLength()
		assert(t, err == nil, err)
		assert(t, bufSize == 14, bufSize)

		buf := make([]byte, bufSize)
		err = h.WriteWithSize(buf, bufSize)
		assert(t, err == nil)

		assert(t, binary.BigEndian.Uint16(buf[0:2]) == FrameHeaderMagic, buf[0:2])
		assert(t, binary.BigEndian.Uint16(buf[2:4]) == h.flags, buf[2:4])
		assert(t, binary.BigEndian.Uint32(buf[4:8]) == uint32(h.seqID), buf[4:8])
		assert(t, binary.BigEndian.Uint16(buf[8:10]) == uint16(bufSize-10)/PaddingSize, buf[8:10])
		assert(t, buf[OffsetProtocol] == h.protocolID, buf[OffsetProtocol])
		assert(t, buf[OffsetNTransform] == 0x00, buf[OffsetNTransform:])                // nTransform: not supported yet
		assert(t, buf[OffsetVariable] == 0x00 && buf[13] == 0x00, buf[OffsetVariable:]) // padding
	})

	t.Run("fixed:short-write", func(t *testing.T) {
		h := NewHeader()
		h.SetIsStreaming()
		h.SetSeqID(0x12345678)
		h.SetProtocolID(0x1)

		bufSize, err := h.BytesLength()
		assert(t, err == nil, err)
		assert(t, bufSize == 14, bufSize)

		buf := make([]byte, bufSize-1)
		err = h.WriteWithSize(buf, bufSize)
		assert(t, err != nil)
	})

	t.Run("fixed+token", func(t *testing.T) {
		h := NewHeader()
		h.SetIsStreaming()
		h.SetSeqID(0x12345678)
		h.SetProtocolID(0x1)
		h.SetToken("test")

		bufSize, err := h.BytesLength()
		assert(t, err == nil, err)
		assert(t, bufSize == 22, bufSize)

		buf := make([]byte, bufSize)
		err = h.WriteWithSize(buf, bufSize)
		assert(t, err == nil)

		assert(t, binary.BigEndian.Uint16(buf[0:2]) == FrameHeaderMagic, buf[0:2])
		assert(t, binary.BigEndian.Uint16(buf[2:4]) == h.flags, buf[2:4])
		assert(t, binary.BigEndian.Uint32(buf[4:8]) == uint32(h.seqID), buf[4:8])
		assert(t, binary.BigEndian.Uint16(buf[8:10]) == uint16(bufSize-10)/PaddingSize, buf[8:10])
		assert(t, buf[OffsetProtocol] == h.protocolID, buf[OffsetProtocol])
		assert(t, buf[OffsetNTransform] == 0x00, buf[OffsetNTransform:]) // nTransform: not supported yet
		assert(t, buf[OffsetVariable] == InfoIDACLToken, buf[OffsetVariable])
		assert(t, binary.BigEndian.Uint16(buf[13:15]) == uint16(len("test")), buf[13:15])
		assert(t, string(buf[15:19]) == "test", buf[15:19])
	})

	t.Run("nTransport", func(t *testing.T) {
		h := NewHeader()
		h.nTransform = 1
		buf := make([]byte, 100)
		err := h.WriteWithSize(buf, 14)
		assert(t, err != nil, err)
	})
}

func TestHeader_BytesLength(t *testing.T) {
	t.Run("fixed", func(t *testing.T) {
		h := NewHeader()
		length, err := h.BytesLength()
		assert(t, err == nil, err)
		assert(t, length == 14, length)
	})
	t.Run("fixed+token", func(t *testing.T) {
		h := NewHeader()
		h.SetToken("test")
		length, err := h.BytesLength()
		assert(t, err == nil, err)
		assert(t, length == 22, length) // fixed(12) + token(7) + padding(3)
	})
	t.Run("too-long", func(t *testing.T) {
		h := NewHeader()
		token := make([]byte, 65536)
		h.SetToken(string(token))
		_, err := h.BytesLength()
		assert(t, err != nil, err)
	})
	t.Run("nTransform", func(t *testing.T) {
		h := NewHeader()
		h.nTransform = 1
		_, err := h.BytesLength()
		assert(t, err != nil, err)
	})
}

func TestHeader_Bytes(t *testing.T) {
	t.Run("fixed", func(t *testing.T) {
		h := NewHeader()
		h.SetIsStreaming()
		h.SetSeqID(0x12345678)
		h.SetProtocolID(0x1)
		buf, err := h.Bytes()
		assert(t, err == nil, err)
		assert(t, len(buf) == 14, len(buf))
	})

	t.Run("fixed+token", func(t *testing.T) {
		h := NewHeader()
		h.SetIsStreaming()
		h.SetSeqID(0x12345678)
		h.SetProtocolID(0x1)
		h.SetToken("test")
		buf, err := h.Bytes()
		assert(t, err == nil, err)
		assert(t, len(buf) == 22, len(buf))
	})

	t.Run("too-long", func(t *testing.T) {
		h := NewHeader()
		token := make([]byte, 65536)
		h.SetToken(string(token))
		_, err := h.Bytes()
		assert(t, err != nil, err)
	})
}

func TestHeader_Read(t *testing.T) {
	t.Run("fixed:short-read", func(t *testing.T) {
		h1 := NewHeader()
		err := h1.Read(nil)
		assert(t, err != nil, err)
	})

	t.Run("invalid-magic", func(t *testing.T) {
		buf := make([]byte, 14)
		h1 := NewHeader()
		err := h1.Read(buf)
		assert(t, errors.Is(err, ErrInvalidMagic), err)
	})

	t.Run("fixed+token", func(t *testing.T) {
		hw := NewHeader()
		hw.SetIsStreaming()
		hw.SetSeqID(0x12345678)
		hw.SetProtocolID(0x1)
		hw.SetToken("test")
		buf, err := hw.Bytes()
		assert(t, err == nil, err)

		hr := NewHeader()
		err = hr.Read(buf)
		assert(t, err == nil, err)
		assert(t, hr.flags == hw.flags, hr.flags)
		assert(t, hr.seqID == hw.seqID, hr.seqID)
		assert(t, hr.protocolID == hw.protocolID, hr.protocolID)
		assert(t, hr.nTransform == hw.nTransform, hr.nTransform)
		assert(t, hr.Token() == hw.Token(), hr.Token())
		assert(t, hr.Size() == 12, hr.Size()) // protocol(1) + nTransform(1) + token(7) + padding(3)
	})

	t.Run("nTransport", func(t *testing.T) {
		hw := NewHeader()
		buf, err := hw.Bytes()
		assert(t, err == nil, err)
		buf[OffsetNTransform] = 1 // nTransport

		hr := NewHeader()
		err = hr.Read(buf)
		assert(t, err != nil, err)
	})
}
