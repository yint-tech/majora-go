package protocol

import (
	"bufio"

	"iinti.cn/majora-go/common"
	"iinti.cn/majora-go/log"
)

type Decoder interface {
	Decode(reader *bufio.Reader) (*MajoraPacket, error)
}

type MajoraPacketDecoder struct {
}

func (mpd *MajoraPacketDecoder) Decode(reader *bufio.Reader) (pack *MajoraPacket, err error) {
	magicbs := make([]byte, common.MagicSize)
	_, err = reader.Read(magicbs)
	if err != nil {
		return nil, err
	}

	if !common.ReadMagic(magicbs) {
		return nil, common.ErrInvalidMagic
	}

	frameLen, err := common.ReadInt32(reader)
	if err != nil {
		return nil, common.ErrInvalidSize
	}

	// type
	msgType, err := common.ReadByte(reader)
	if err != nil {
		log.Run().Errorf("read type error  %+v", err)
		return nil, common.ErrInvalidSize
	}
	pack = &MajoraPacket{}
	pack.Ttype = MajoraPacketType(msgType)

	// num
	pack.SerialNumber, err = common.ReadInt64(reader)
	if err != nil {
		log.Run().Errorf("read type error  %+v", err)
		return nil, common.ErrInvalidSize
	}

	// extra size
	extraSize, err := common.ReadByte(reader)
	if err != nil {
		log.Run().Errorf("read type error  %+v", err)
		return nil, common.ErrInvalidSize
	}

	extra, err := common.ReadN(int(extraSize), reader)
	if err != nil {
		log.Run().Errorf("read type error  %+v", err)
		return nil, common.ErrInvalidSize
	}
	pack.Extra = string(extra)

	// dataFrame
	dataSize := int(frameLen) - common.TypeSize - common.SerialNumberSize - common.ExtraSize - int(extraSize)
	if dataSize < 0 {
		log.Run().Errorf("read type error  %+v", err)
		return nil, common.ErrInvalidSize
	}

	if dataSize > 0 {
		data, err := common.ReadN(dataSize, reader)
		if err != nil {
			log.Run().Errorf("read type error  %+v", err)
		}
		pack.Data = data
	}
	return pack, nil
}
