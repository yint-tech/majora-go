package client

import (
	"runtime"
	"time"

	"github.com/adamweixuan/getty"
	"iinti.cn/majora-go/common"
	"iinti.cn/majora-go/log"
	"iinti.cn/majora-go/protocol"
)

type MajoraEventListener struct {
	client *Client
}

func (m *MajoraEventListener) OnOpen(session getty.Session) error {
	m.client.session = session
	packet := protocol.TypeRegister.CreatePacket()
	packet.Extra = m.client.config.ClientID
	extraMap := make(map[string]string, 1)
	extraMap[common.ExtrakeyUser] = m.client.config.Extra.Account
	packet.Data = protocol.EncodeExtra(extraMap)
	if _, _, err := session.WritePkg(packet, time.Second*10); err != nil {
		log.Error().Errorf("register to server error %+v", err)
		return err
	}
	log.Run().Infof("[OnOpen] register to %s success", m.client.config.TunnelAddr)

	return nil
}

func (m *MajoraEventListener) OnClose(session getty.Session) {
	log.Error().Errorf("OnClose-> session closed %v", session.IsClosed())
	m.client.CloseAll()
}

func (m *MajoraEventListener) OnError(session getty.Session, err error) {
	log.Error().Errorf("OnError %s. session is closed:%v", err.Error(), session.IsClosed())
	m.client.CloseAll()
}

func (m *MajoraEventListener) OnCron(_ getty.Session) {
}

func (m *MajoraEventListener) OnMessage(session getty.Session, input interface{}) {
	defer func() {
		if err := recover(); err != nil {
			var buf [4096]byte
			n := runtime.Stack(buf[:], false)
			log.Error().Errorf("goroutine panic OnMessage :%s,err:%+v", string(buf[:n]), err)
		}
	}()

	majoraPacket := input.(*protocol.MajoraPacket)
	log.Run().Debugf("receive packet from server %d->%s", majoraPacket.SerialNumber, majoraPacket.Ttype.ToString())

	switch majoraPacket.Ttype {
	case protocol.TypeHeartbeat:
		m.client.handleHeartbeat(session)
	case protocol.TypeConnect:
		m.client.handleConnect(majoraPacket, session)
	case protocol.TypeTransfer:
		m.client.handleTransfer(majoraPacket, session)
	case protocol.TypeDisconnect:
		m.client.handleDisconnectMessage(session, majoraPacket)
	case protocol.TypeControl:
		m.client.handleControlMessage(session, majoraPacket)
	case protocol.TypeDestroy:
		m.client.handleDestroyMessage()
	}
}
