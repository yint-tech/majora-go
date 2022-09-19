package protocol

type MajoraPacket struct {
	Ttype        MajoraPacketType // 消息类型
	SerialNumber int64            // 流水号
	Data         []byte           // 核心数据
	Extra        string
}

type (
	MajoraPacketType byte
)

const (
	TypeRegister     MajoraPacketType = 0x01
	TypeHeartbeat    MajoraPacketType = 0x02
	TypeConnect      MajoraPacketType = 0x03
	TypeDisconnect   MajoraPacketType = 0x04
	TypeTransfer     MajoraPacketType = 0x05
	TypeControl      MajoraPacketType = 0x06
	TypeConnectReady MajoraPacketType = 0x07
	TypeDestroy      MajoraPacketType = 0x08
	TypeOffline      MajoraPacketType = 0x09
)

func (mpt MajoraPacketType) CreatePacket() *MajoraPacket {
	return &MajoraPacket{Ttype: mpt}
}

func (mpt MajoraPacketType) ToString() string {
	switch mpt {
	case TypeHeartbeat:
		return "Heartbeat"
	case TypeRegister:
		return "TypeRegister"
	case TypeConnect:
		return "TypeConnect"
	case TypeDisconnect:
		return "TypeDisconnect"
	case TypeTransfer:
		return "TypeTransfer"
	case TypeControl:
		return "TypeControl"
	case TypeConnectReady:
		return "TypeConnectReady"
	case TypeDestroy:
		return "TypeDestroy"
	case TypeOffline:
		return "TypeOffline"
	}
	return "Unknown"
}
