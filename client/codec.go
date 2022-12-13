package client

import (
	"bytes"
	"encoding/binary"

	"github.com/adamweixuan/getty"

	"iinti.cn/majora-go/common"
	"iinti.cn/majora-go/log"
	"iinti.cn/majora-go/protocol"
)

var PkgCodec = &PacketCodec{}

type PacketCodec struct{}

func (p *PacketCodec) Read(session getty.Session, data []byte) (interface{}, int, error) {
	log.Run().Debugf("[PacketCodec] length:%d", len(data))
	if len(data) < common.MagicSize+common.FrameSize {
		return nil, 0, nil
	}
	log.Run().Debugf("[PacketCodec] read magic %+v", binary.BigEndian.Uint64(data[0:8]))

	readMagic := data[0:common.MagicSize]

	if !common.ReadMagic(readMagic) {
		log.Run().Errorf("[PacketCodec] invalid magic %d|%s",
			binary.BigEndian.Uint64(readMagic), string(readMagic))
		return nil, 0, common.ErrInvalidMagic
	}

	reader := bytes.NewBuffer(data[common.MagicSize:])

	frameLen, err := common.ReadInt32(reader)
	if err != nil {
		log.Error().Errorf("[PacketCodec] frameLen error %+v", err)
		return nil, 0, err
	}

	// 缓冲区数据不够
	if reader.Len() < int(frameLen) {
		log.Run().Debugf("[PacketCodec] buf not enough %d|%d", reader.Len(), frameLen)
		return nil, 0, nil
	}

	log.Run().Debugf("[PacketCodec] read frameLen %d", frameLen)

	// type
	msgType, err := common.ReadByte(reader)
	if err != nil {
		log.Error().Errorf("[PacketCodec] read type error %+v", err)
		return nil, 0, err
	}

	log.Run().Debugf("[PacketCodec] read msgType %+v", msgType)

	pack := &protocol.MajoraPacket{}
	pack.Ttype = protocol.MajoraPacketType(msgType)

	// num
	pack.SerialNumber, err = common.ReadInt64(reader)
	if err != nil {
		log.Error().Errorf("[PacketCodec] read num error %+v", err)
		return nil, len(data), nil
	}

	log.Run().Debugf("[PacketCodec] read SerialNumber %d", pack.SerialNumber)

	// extra size
	extraSize, err := common.ReadByte(reader)
	if err != nil {
		log.Error().Errorf("[PacketCodec] read extra size error  %+v", err)
		return nil, len(data), nil
	}

	extra, err := common.ReadN(int(extraSize), reader)
	if err != nil {
		log.Error().Errorf("[PacketCodec] read extra error  %+v", err)
		return nil, len(data), nil
	}
	pack.Extra = string(extra)
	log.Run().Debugf("[PacketCodec] read extra %s", pack.Extra)

	// dataFrame
	dataSize := int(frameLen) - common.TypeSize - common.SerialNumberSize - common.ExtraSize - int(extraSize)
	if dataSize < 0 {
		log.Error().Errorf("[PacketCodec] read frameLen error  %+v", err)
		return nil, len(data), common.ErrInvalidSize
	}

	if dataSize > 0 {
		data, err := common.ReadN(dataSize, reader)
		if err != nil {
			log.Error().Errorf("[PacketCodec] read data error %+v", err)
			return nil, len(data), nil
		}
		pack.Data = data
	}
	log.Run().Debugf("[PacketCodec] read response %d", int(frameLen+12))
	return pack, int(frameLen + 12), nil
}

func (p *PacketCodec) Write(session getty.Session, packet interface{}) ([]byte, error) {
	majoraPkt, ok := packet.(*protocol.MajoraPacket)
	if !ok {
		return nil, common.ErrNilPacket
	}
	return protocol.Codec.Encode(majoraPkt), nil
}
