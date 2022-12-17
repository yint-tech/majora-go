package protocol

import (
	"bufio"
)

var Codec ICodec = NewDefCodec()

type (
	ICodec interface {
		Encode(packet *MajoraPacket) []byte
		Decode(reader *bufio.Reader) (*MajoraPacket, error)
	}

	DefCodec struct {
		Encoder Encoder
		Decoder Decoder
	}
)

func NewDefCodec() *DefCodec {
	return &DefCodec{
		Encoder: &MajoraPacketEncoder{},
		Decoder: &MajoraPacketDecoder{},
	}
}

func (d *DefCodec) Encode(packet *MajoraPacket) []byte {
	return d.Encoder.Encode(packet)
}

func (d *DefCodec) Decode(reader *bufio.Reader) (*MajoraPacket, error) {
	return d.Decoder.Decode(reader)
}
