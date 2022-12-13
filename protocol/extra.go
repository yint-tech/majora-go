package protocol

import (
	"bytes"

	"iinti.cn/majora-go/log"
)

func EncodeExtra(extra map[string]string) []byte {
	if len(extra) == 0 {
		return nil
	}
	buf := bytes.Buffer{}
	buf.WriteByte(byte(len(extra)))
	for key, val := range extra {
		keybs := []byte(key)
		valbs := []byte(val)

		buf.WriteByte(byte(len(keybs)))
		buf.Write(keybs)
		buf.WriteByte(byte(len(valbs)))
		buf.Write(valbs)
	}
	return buf.Bytes()
}

func DecodeExtra(input []byte) map[string]string {
	if len(input) < 4 {
		return map[string]string{}
	}
	buffer := bytes.NewBuffer(input)

	headerSize, err := buffer.ReadByte()
	if err != nil {
		log.Run().Errorf("DecodeExtra error %+v", err)
		return map[string]string{}
	}
	data := make(map[string]string)
	for i := 0; i < int(headerSize); i++ {
		keyLen, _ := buffer.ReadByte()
		keyBs := make([]byte, keyLen)
		_, _ = buffer.Read(keyBs)
		valLen, _ := buffer.ReadByte()
		valBs := make([]byte, valLen)
		_, _ = buffer.Read(valBs)
		data[string(keyBs)] = string(valBs)
	}
	return data
}
