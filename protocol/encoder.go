package protocol

import (
	"bytes"

	"iinti.cn/majora-go/common"
)

type Encoder interface {
	Encode(*MajoraPacket) []byte
}

type MajoraPacketEncoder struct{}

func (s *MajoraPacketEncoder) Encode(packet *MajoraPacket) []byte {
	if packet == nil {
		return nil
	}
	bodyLength := common.TypeSize + common.SerialNumberSize + common.ExtraSize

	bodyLength += len(packet.Data)

	bodyLength += len([]byte(packet.Extra))

	var (
		innerBuf = make([]byte, 0, bodyLength+8+4)
		// todo 池化提高性能
		buffer = bytes.NewBuffer(innerBuf)
	)

	// magic 8byte
	buffer.Write(common.ConvertInt64ToBytes(common.MAGIC))
	// body length 4byte
	buffer.Write(common.ConvertInt32ToBytes(int32(bodyLength)))
	// type 1byte
	buffer.WriteByte(byte(packet.Ttype))
	// serial num 4byte
	buffer.Write(common.ConvertInt64ToBytes(packet.SerialNumber))

	// extra
	if len(packet.Extra) > 0 {
		extraBs := []byte(packet.Extra)
		buffer.WriteByte(byte(len(extraBs)))
		buffer.Write(extraBs)
	} else {
		buffer.WriteByte(0x00)
	}

	if len(packet.Data) > 0 {
		buffer.Write(packet.Data)
	}
	return buffer.Bytes()
}
