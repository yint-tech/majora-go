package client

import (
	"fmt"
	"github.com/adamweixuan/getty"
	gxsync "github.com/adamweixuan/gostnops/sync"
	"net"
	"runtime"

	"iinti.cn/majora-go/common"
	"iinti.cn/majora-go/log"
)

var (
	taskPool = gxsync.NewTaskPoolSimple(runtime.GOMAXPROCS(-1) * 100)
)

func (client *Client) connect() {

	reConnect := client.config.ReconnInterval

	gettyCli := getty.NewTCPClient(
		getty.WithServerAddress(fmt.Sprintf("%s:%d", client.host, client.port)),
		getty.WithConnectionNumber(1),
		getty.WithClientTaskPool(taskPool),
		getty.WithReconnectInterval(int(reConnect.Milliseconds())),
		getty.WithLocalAddressClient(client.localAddr))
	gettyCli.RunEventLoop(NewClientSession(client))
	client.natTunnel = gettyCli
}

func NewClientSession(client *Client) func(getty.Session) error {
	return func(session getty.Session) error {
		return InitialSession(session, client)
	}
}

func InitialSession(session getty.Session, client *Client) (err error) {
	tcpConn, ok := session.Conn().(*net.TCPConn)
	if !ok {
		panic(fmt.Sprintf("invalid session %+v", session.Conn()))
	}

	if err = tcpConn.SetNoDelay(true); err != nil {
		return err
	}
	if err = tcpConn.SetKeepAlive(true); err != nil {
		return err
	}
	if err = tcpConn.SetKeepAlivePeriod(common.KeepAliveTimeout); err != nil {
		return err
	}
	if err = tcpConn.SetReadBuffer(common.MB); err != nil {
		return err
	}
	if err = tcpConn.SetWriteBuffer(common.MB); err != nil {
		return err
	}

	session.SetName(common.SessionName)
	session.SetMaxMsgLen(common.KB8)
	session.SetReadTimeout(common.ReadTimeout)
	session.SetWriteTimeout(common.WriteTimeout)
	session.SetWaitTime(common.WaitTimeout)

	session.SetPkgHandler(PkgCodec)
	session.SetEventListener(&MajoraEventListener{
		client: client,
	})
	return nil
}

func (client *Client) Redial(tag string) {
	log.Run().Infof("[Redial %s] start, can redial? %v", tag, client.config.Redial.Valid())
	if !client.config.Redial.Valid() {
		return
	}
	log.Run().Infof("[Redial %s] Send offline message", tag)
	if _, _, err := client.session.WritePkg(OfflinePacket, 0); err != nil {
		log.Run().Errorf("[Redial %s] write offline to server error %s", tag, err.Error())
	}
	log.Run().Info("[Redial %s %s] start close local session", client.host, tag)

}
