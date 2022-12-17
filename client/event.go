package client

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/adamweixuan/getty"

	"iinti.cn/majora-go/common"
	"iinti.cn/majora-go/global"
	"iinti.cn/majora-go/log"
	"iinti.cn/majora-go/protocol"
	"iinti.cn/majora-go/safe"
	"iinti.cn/majora-go/trace"
)

var ( //nolint:gofumpt
	HeartbeatPacket = protocol.TypeHeartbeat.CreatePacket()
	OfflinePacket   = protocol.TypeOffline.CreatePacket()
)

func (client *Client) handleHeartbeat(session getty.Session) {
	if _, _, err := session.WritePkg(HeartbeatPacket, common.HeartBeatTimeout); err != nil {
		log.Error().Errorf("handleHeartbeat error %+v %v", err, session.IsClosed())
	} else {
		log.Run().Infof("handleHeartbeat success")
	}
}

func (client *Client) handleConnect(packet *protocol.MajoraPacket, session getty.Session) {
	m := decodeMap(packet.Data)
	if len(m) == 0 {
		log.Error().Errorf("Get map data from connect packet failed (sn:%d)", packet.SerialNumber)
	}
	sessionID, ok := m[trace.MajoraSessionName]
	if !ok {
		log.Error().Errorf("Get sessionId from connect packet failed (sn:%d)", packet.SerialNumber)
	}
	enableTrace, ok := m[trace.MajoraTraceEnable]
	if !ok {
		log.Error().Errorf("Get enableTrace from connect packet failed (sn:%d)", packet.SerialNumber)
	}
	user, ok := m[trace.MajoraSessionUser]
	if !ok {
		log.Error().Errorf("Get user from connect packet failed (sn:%d)", packet.SerialNumber)
	}
	traceSession := trace.NewSession(sessionID, client.host, user, enableTrace == "true")
	client.AddSession(packet, traceSession)
	traceSession.Recorder.RecordEvent(trace.ConnectEvent, fmt.Sprintf("Start handle connect to %s (sn:%d)",
		packet.Extra, packet.SerialNumber))

	if session.IsClosed() {
		log.Run().Warnf("[handleConnect] %d -> nat server is closed", packet.SerialNumber)
		traceSession.Recorder.RecordErrorEvent(trace.ConnectEvent, fmt.Sprintf("NatServer is closed (sn:%d)", packet.SerialNumber), nil)
		client.closeVirtualConnection(session, packet.SerialNumber)
		return
	}
	hostPort := strings.Split(packet.Extra, ":")
	if len(packet.Extra) == 0 || len(hostPort) != 2 {
		log.Error().Errorf("[handleConnect] invalid extra %s", packet.Extra)
		traceSession.Recorder.RecordErrorEvent(trace.ConnectEvent,
			fmt.Sprintf("Connect extra invalid %s (%d)", packet.Extra, packet.SerialNumber), nil)
		client.closeVirtualConnection(session, packet.SerialNumber)
		return
	}

	dialer := net.Dialer{
		Timeout:   common.UpstreamTimeout,
		LocalAddr: client.localAddr,
	}

	var target string
	ip, err := client.dnsCache.Get([]byte(hostPort[0]))
	if err != nil {
		traceSession.Recorder.RecordEvent(trace.DNSResolveEvent, fmt.Sprintf("Dns cache miss %s ", hostPort[0]))
		hosts, dnsErr := net.LookupHost(hostPort[0])
		if dnsErr != nil {
			traceSession.Recorder.RecordErrorEvent(trace.DNSResolveEvent, fmt.Sprintf("Resolve %s ip error", hostPort[0]), dnsErr)
			client.closeVirtualConnection(session, packet.SerialNumber)
			return
		}
		err := client.dnsCache.Set([]byte(hostPort[0]), []byte(hosts[0]), int(global.Config.DNSCacheDuration.Seconds()))
		if err != nil {
			traceSession.Recorder.RecordErrorEvent(trace.DNSResolveEvent, fmt.Sprintf("Dns cache set error %s", hostPort[0]), err)
		}
		target = hosts[0]
	} else {
		target = string(ip)
	}

	traceSession.Recorder.RecordEvent(trace.DNSResolveEvent, fmt.Sprintf("Dns cache hit %s -> %s", hostPort[0], target))

	// ipv6
	conn, err := dialer.Dial(common.TCP, net.JoinHostPort(target, hostPort[1]))
	if err != nil {
		log.Error().Errorf("[handleConnect] %d->connect to %s->%s", packet.SerialNumber, packet.Extra, err.Error())
		traceSession.Recorder.RecordErrorEvent(trace.ConnectEvent,
			fmt.Sprintf("Connect to %s failed (sn:%d)", packet.Extra, packet.SerialNumber), err)
		client.closeVirtualConnection(session, packet.SerialNumber)
		return
	}
	tcpConn := conn.(*net.TCPConn)
	_ = tcpConn.SetNoDelay(true)
	_ = tcpConn.SetKeepAlive(true)
	client.AddConnection(packet, tcpConn, packet.Extra)
	traceSession.Recorder.RecordEvent(trace.ConnectEvent, fmt.Sprintf("Connect to %s success, local: %s -> remote:%s (sn:%d)",
		packet.Extra, tcpConn.LocalAddr(), tcpConn.RemoteAddr(), packet.SerialNumber))

	traceSession.Recorder.RecordEvent(trace.ConnectEvent, fmt.Sprintf("Start replay natServer connect ready (sn:%d)", packet.SerialNumber))
	majoraPacket := protocol.TypeConnectReady.CreatePacket()
	majoraPacket.SerialNumber = packet.SerialNumber
	majoraPacket.Extra = client.config.ClientID
	if session.IsClosed() {
		log.Run().Warnf("[handleConnect] %d -> nat server is closed", packet.SerialNumber)
		traceSession.Recorder.RecordErrorEvent(trace.ConnectEvent, fmt.Sprintf("NatServer is closed (sn:%d)", packet.SerialNumber),
			nil)
		client.closeVirtualConnection(session, packet.SerialNumber)
		return
	}
	if _, _, err := session.WritePkg(majoraPacket, 0); err != nil {
		log.Error().Errorf("[handleConnect] %d->write pkg to nat server with error %s", packet.SerialNumber,
			err.Error())
		traceSession.Recorder.RecordErrorEvent(trace.ConnectEvent, fmt.Sprintf("Write pkg to natServer failed (sn:%d)",
			packet.SerialNumber), err)
		client.closeVirtualConnection(session, packet.SerialNumber)
		return
	}
	safe.Go("handleConnect", func() {
		client.handleUpStream(tcpConn, packet, session)
	})
	log.Run().Debugf("[handleConnect] %d->connect success to %s ", packet.SerialNumber, packet.Extra)
	traceSession.Recorder.RecordEvent(trace.ConnectEvent, fmt.Sprintf("Replay natServer connect ready success (sn:%d)", packet.SerialNumber))
}

func decodeMap(data []byte) map[string]string {
	result := make(map[string]string, 1)
	var headerSize int8
	err := binary.Read(bytes.NewBuffer(data[:1]), binary.BigEndian, &headerSize)
	data = data[1:]
	if err != nil {
		return result
	}
	var i int8
	for i = 0; i < headerSize; i++ {
		var keyLength int8
		_ = binary.Read(bytes.NewBuffer(data[:1]), binary.BigEndian, &keyLength)
		data = data[1:]

		key := make([]byte, keyLength)
		_ = binary.Read(bytes.NewBuffer(data[:keyLength]), binary.BigEndian, &key)
		data = data[keyLength:]

		var valueLength int8
		_ = binary.Read(bytes.NewBuffer(data[:1]), binary.BigEndian, &valueLength)
		data = data[1:]

		value := make([]byte, valueLength)
		_ = binary.Read(bytes.NewBuffer(data[:valueLength]), binary.BigEndian, &value)
		data = data[valueLength:]

		result[string(key)] = string(value)
	}
	return result
}

func (client *Client) handleTransfer(packet *protocol.MajoraPacket, session getty.Session) {
	traceRecorder := client.GetRecorderFromSession(packet.SerialNumber)
	traceRecorder.RecordEvent(trace.TransferEvent,
		fmt.Sprintf("Receive transfer packet from natServer,start to be forward to target, len:%d (%d)", len(packet.Data), packet.SerialNumber))

	load, ok := client.connStore.Load(packet.SerialNumber)
	if !ok {
		log.Error().Errorf("[handleTransfer] %d-> can not find connection", packet.SerialNumber)
		traceMessage := fmt.Sprintf("Find upstream connection failed (%d)", packet.SerialNumber)
		traceRecorder.RecordErrorEvent(trace.TransferEvent, traceMessage, nil)
		client.closeVirtualConnection(session, packet.SerialNumber)
		return
	}
	conn := load.(*net.TCPConn)
	cnt, err := conn.Write(packet.Data)
	if err != nil {
		log.Error().Errorf("[handleTransfer] %d->write to upstream fail for %s", packet.SerialNumber, err)
		traceMessage := fmt.Sprintf("Write to upstream failed (%d)", packet.SerialNumber)
		traceRecorder.RecordErrorEvent(trace.TransferEvent, traceMessage, err)
		client.closeVirtualConnection(session, packet.SerialNumber)
		return
	}

	if cnt != len(packet.Data) {
		log.Error().Errorf("[handleTransfer] %d-> write not all data for expect->%d/%d",
			packet.SerialNumber, len(packet.Data), cnt)
		traceMessage := fmt.Sprintf("Write not all data for expect -> %d/%d (sn:%d)", len(packet.Data), cnt, packet.SerialNumber)
		traceRecorder.RecordErrorEvent(trace.TransferEvent, traceMessage, nil)
		client.closeVirtualConnection(session, packet.SerialNumber)
		return
	}
	log.Run().Debugf("[handleTransfer] %d-> success dataLen: %d", packet.SerialNumber, len(packet.Data))
	traceMessage := fmt.Sprintf("transfer data success (%d)", packet.SerialNumber)
	traceRecorder.RecordEvent(trace.TransferEvent, traceMessage)
}

func (client *Client) handleUpStream(conn *net.TCPConn, packet *protocol.MajoraPacket, session getty.Session) {
	traceRecorder := client.GetRecorderFromSession(packet.SerialNumber)
	traceRecorder.RecordEvent(trace.UpStreamEvent, fmt.Sprintf("Ready read from upstream (sn:%d)", packet.SerialNumber))
	log.Run().Debugf("[handleUpStream] %d-> handleUpStream start...", packet.SerialNumber)
	for {
		buf := make([]byte, common.BufSize)
		cnt, err := conn.Read(buf)
		if err != nil {
			opErr, ok := err.(*net.OpError)
			if ok && opErr.Err.Error() == "i/o timeout" {
				recorderMessage := fmt.Sprintf("Upstream deadDeadline start close (sn:%d)", packet.SerialNumber)
				traceRecorder.RecordEvent(trace.UpStreamEvent, recorderMessage)
			} else {
				log.Run().Debugf("[handleUpStream] %d->read with error:%+v,l:%s->r:%s",
					packet.SerialNumber, err, conn.LocalAddr(), conn.RemoteAddr())
				recorderMessage := fmt.Sprintf("Read with l:%s->r:%s (sn:%d) ",
					conn.LocalAddr(), conn.RemoteAddr(), packet.SerialNumber)
				traceRecorder.RecordErrorEvent(trace.UpStreamEvent, recorderMessage, err)
			}
			client.OnClose(session, conn, packet.SerialNumber)
			break
		}
		traceRecorder.RecordEvent(trace.UpStreamEvent, fmt.Sprintf("read count: %d (sn:%d)",
			cnt, packet.SerialNumber))

		traceRecorder.RecordEvent(trace.UpStreamEvent, fmt.Sprintf("Start write to natServer (sn:%d)", packet.SerialNumber))
		pack := protocol.TypeTransfer.CreatePacket()
		pack.Data = buf[0:cnt]
		pack.SerialNumber = packet.SerialNumber
		if _, _, err := session.WritePkg(pack, 0); err != nil {
			log.Error().Errorf("[handleUpStream] %d-> write to server fail %+v", packet.SerialNumber, err.Error())
			traceRecorder.RecordErrorEvent(trace.UpStreamEvent,
				fmt.Sprintf("Write to natServer failed (sn:%d)", packet.SerialNumber), err)
			client.OnClose(session, conn, packet.SerialNumber)
			break
		} else {
			log.Run().Debugf("[handleUpStream] %d->success dataLen:%d", packet.SerialNumber, len(packet.Data))
			traceRecorder.RecordEvent(trace.UpStreamEvent,
				fmt.Sprintf("Write to natServer success (sn:%d)", packet.SerialNumber))
		}
	}
}

func (client *Client) handleDisconnectMessage(session getty.Session, packet *protocol.MajoraPacket) {
	traceRecorder := client.GetRecorderFromSession(packet.SerialNumber)
	traceRecorder.RecordEvent(trace.DisconnectEvent, fmt.Sprintf("Start close upstream extra:%s (sn:%d)",
		packet.Extra, packet.SerialNumber))
	log.Run().Debugf("[handleDisconnectMessage] %d->session closed %v extra:%s", packet.SerialNumber, session.IsClosed())
	if conn, ok := client.connStore.Load(packet.SerialNumber); ok {
		upstreamConn := conn.(*net.TCPConn)
		readDeadLine := time.Now().Add(3 * time.Millisecond)
		traceRecorder.RecordEvent(trace.DisconnectEvent, fmt.Sprintf("Set upstream read deadline:%s (sn:%d)",
			readDeadLine.Format("2006-01-02 15:04:05.000000"), packet.SerialNumber))
		err := upstreamConn.SetReadDeadline(readDeadLine)
		if err != nil {
			traceRecorder.RecordErrorEvent(trace.DisconnectEvent,
				fmt.Sprintf("Set upstream read deadline failed (sn:%d)", packet.SerialNumber), err)
			client.OnClose(session, upstreamConn, packet.SerialNumber)
		}

	} else {
		traceRecorder.RecordEvent(trace.DisconnectEvent, fmt.Sprintf("The upstream connection is closed, do nothing (sn:%d)", packet.SerialNumber))
	}
}

func (client *Client) handleControlMessage(_ *protocol.MajoraPacket) {
	log.Run().Debugf("handleControlMessage")
}

// handleDestroyMessage 是直接关闭nat server ?
func (client *Client) handleDestroyMessage() {
	client.natTunnel.Close()
}

func (client *Client) AddSession(packet *protocol.MajoraPacket, session *trace.Session) {
	if _, ok := client.sessionStore.Load(packet.SerialNumber); ok {
		log.Error().Errorf("[AddSession] %d->error, has one", packet.SerialNumber)
	}
	client.sessionStore.Store(packet.SerialNumber, session)
	log.Run().Debugf("[AddSession] %d-> success", packet.SerialNumber)
}

func (client *Client) GetRecorderFromSession(sn int64) trace.Recorder {
	session, ok := client.sessionStore.Load(sn)
	if !ok {
		log.Run().Warnf("[GetRecorderFromSession] get session failed, maybe already closed (%d)", sn)
		session = trace.NewSession("", "", "", false)
	}
	traceSession := session.(*trace.Session)
	return traceSession.Recorder
}

func (client *Client) AddConnection(packet *protocol.MajoraPacket, conn *net.TCPConn, addr string) {
	if _, ok := client.connStore.Load(packet.SerialNumber); ok {
		log.Error().Errorf("[AddConnection] %d->error, has one", packet.SerialNumber)
	}
	client.connStore.Store(packet.SerialNumber, conn)
	log.Run().Debugf("[AddConnection] %d->%s success", packet.SerialNumber, addr)
}

// OnClose 1. 本地缓存删除 2. 关闭连接  3. 通知natserver
func (client *Client) OnClose(natSession getty.Session, upStreamSession net.Conn, serialNumber int64) {
	defer func() {
		if err := recover(); err != nil {
			var buf [4096]byte
			n := runtime.Stack(buf[:], false)
			log.Error().Errorf("goroutine panic OnClose :%s,err:%+v", string(buf[:n]), err)
		}
	}()
	client.closeVirtualConnection(natSession, serialNumber)
	_ = upStreamSession.Close()
}

// closeVirtualConnection disconnect to server
func (client *Client) closeVirtualConnection(session getty.Session, serialNumber int64) {
	traceRecorder := client.GetRecorderFromSession(serialNumber)

	log.Run().Debugf("[closeVirtualConnection] %d->session closed %v", serialNumber, session.IsClosed())
	if session.IsClosed() {
		log.Run().Warnf("[closeVirtualConnection] %d->session is closed", serialNumber)
		return
	}

	traceRecorder.RecordEvent(trace.DisconnectEvent, fmt.Sprintf("Start send disconnect to natServer (sn:%d)", serialNumber))
	majoraPacket := protocol.TypeDisconnect.CreatePacket()
	majoraPacket.SerialNumber = serialNumber
	majoraPacket.Extra = client.config.ClientID
	if allCnt, sendCnt, err := session.WritePkg(majoraPacket, 0); err != nil {
		log.Run().Warnf("[closeVirtualConnection] ->%d error %s session closed %v allCnt %d sendCnt %d",
			serialNumber, err.Error(), session.IsClosed(), allCnt, sendCnt)
		traceRecorder.RecordErrorEvent(trace.DisconnectEvent,
			fmt.Sprintf("Send disconnect to natServer failed closed:%v allCnt %d sendCnt %d (sn:%d)",
				session.IsClosed(), allCnt, sendCnt, serialNumber), err)
		session.Close()
	}
	client.connStore.Delete(serialNumber)
	client.sessionStore.Delete(serialNumber)
	traceRecorder.RecordEvent(trace.DisconnectEvent, fmt.Sprintf("Send disconnect to natServer success (sn:%d)", serialNumber))
}

func (client *Client) CloseAll() {
	defer func() {
		if err := recover(); err != nil {
			var buf [4096]byte
			n := runtime.Stack(buf[:], false)
			log.Error().Errorf("goroutine-CloseAll panic.stack:%s,err:%+v", string(buf[:n]), err)
		}
	}()
	client.connStore.Range(func(key, value interface{}) bool {
		serialNumber := key.(int64)
		conn, _ := value.(*net.TCPConn)
		log.Run().Debugf("[CloseAll] close serialNumber -> %d", serialNumber)
		client.OnClose(client.session, conn, serialNumber)
		return true
	})
	client.connStore = sync.Map{}
	client.sessionStore = sync.Map{}
}
