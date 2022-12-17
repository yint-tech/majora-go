package common

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"time"
)

const (
	MagicSize        = 8
	FrameSize        = 4
	HeaderSize       = FrameSize + FrameSize
	TypeSize         = 1
	ExtraSize        = 1
	SerialNumberSize = 8
	MAGIC            = int64(0x6D616A6F72613031)
)

const (
	TCP = "tcp"
)

const (
	DNSServer = "114.114.114.114:53"
)

const (
	UpdateServer     = "http://81.70.224.147:10010"
	Latest           = "latest.txt"
	UpdateBinaryPath = "/version/"
	// VersionTpl majora-v0.0.1-linux-arm64 name-version-os-arch
	VersionTpl = "%s-%s-%s-%s"
)

const (
	DefaultMode = 0o755
)

const (
	MB      = 1024 * 1024
	KB8     = 1024 * 8
	BufSize = 1024 * 4
)

const (
	ReadTimeout      = time.Minute
	WriteTimeout     = time.Minute
	WaitTimeout      = time.Minute
	KeepAliveTimeout = time.Second * 10

	SessionTimeout   = time.Minute * 5
	HeartBeatTimeout = time.Second * 30
	UpstreamTimeout  = time.Minute
)

const (
	SessionName = "majora-cli"
)

const (
	DefNatServerHost = "majora.iinti.cn"
	DefNatServerPort = 5879
	DefNatAddr       = "majora.iinti.cn:5879"
	PprofAddr        = "127.0.0.1:6060"
)

const (
	ExtrakeyUser = "majora.key.user"
)

var (
	ErrInvalidSize  = errors.New("invalid size")
	ErrInvalidMagic = errors.New("invalid magic")
	ErrNilPacket    = errors.New("nil packet")
)

func ConvertInt32ToBytes(input int32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(input))
	return buf
}

func ConvertInt64ToBytes(input int64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(input))
	return buf
}

func ReadInt32(conn io.Reader) (int32, error) {
	buf := make([]byte, 4)
	readSize, err := conn.Read(buf)

	if readSize != 4 || err != nil {
		return 0, err
	}
	return int32(binary.BigEndian.Uint32(buf)), nil
}

func ReadInt64(conn io.Reader) (int64, error) {
	buf := make([]byte, 8)
	readSize, err := conn.Read(buf)

	if readSize != 8 || err != nil {
		return 0, err
	}
	return int64(binary.BigEndian.Uint64(buf)), nil
}

func ReadByte(conn io.Reader) (byte, error) {
	buf := make([]byte, 1)
	readSize, err := conn.Read(buf)

	if readSize != 1 || err != nil {
		return 0, err
	}
	oneByte := uint8(0)
	err = binary.Read(bytes.NewBuffer(buf), binary.BigEndian, &oneByte)
	return oneByte, err
}

func ReadN(size int, conn io.Reader) ([]byte, error) {
	buf := make([]byte, size)
	readSize, err := conn.Read(buf)

	if readSize != size || err != nil {
		return nil, err
	}
	return buf, nil
}

func ReadMagic(buf []byte) bool {
	magic := int64(binary.BigEndian.Uint64(buf))
	return magic == MAGIC
}
