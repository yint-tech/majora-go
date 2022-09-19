package trace

import (
	"runtime"
	"time"

	"iinti.cn/majora-go/log"
	"iinti.cn/majora-go/safe"
)

var (
	sessionEventChan  = make(chan *sessionEvent, runtime.GOMAXPROCS(-1)*100)
	ConnectEvent      = "ConnectEvent"
	TransferEvent     = "TransferEvent"
	MajoraSessionName = "MajoraSessionId"
	MajoraTraceEnable = "traceEnable"
	MajoraSessionUser = "user"
	UpStreamEvent     = "ReadUpStream"
	DisconnectEvent   = "Disconnect"
	DnsResolveEvent   = "DnsResolve"
	sessionIdNop      = "session_id_not_set"
)

func init() {
	safe.SafeGo(func() {
		for {
			e := <-sessionEventChan
			if e.Err != nil {
				log.Trace().Errorf("[%s] [%s] [%s] [%s] [%s] %s error:%+v",
					e.natHost, e.sessionId, e.user, e.Timestamp.Format("2006-01-02 15:04:05.000000"), e.EventName, e.Message, e.Err)
			} else {
				log.Trace().Infof("[%s] [%s] [%s] [%s] [%s] %s",
					e.natHost, e.sessionId, e.user, e.Timestamp.Format("2006-01-02 15:04:05.000000"), e.EventName, e.Message)
			}
		}
	})
}

// Event 事件
type Event struct {
	// 发生时间
	Timestamp time.Time

	// 事件名称
	EventName string

	// 事件消息
	Message string

	// 错误，如果存在
	Err error
}

type sessionEvent struct {
	user      string
	sessionId string
	natHost   string
	*Event
}

type Recorder interface {
	RecordEvent(eventName string, message string)

	RecordErrorEvent(eventName string, message string, err error)

	Enable() bool
}

type nopRecorder struct{}

func (n *nopRecorder) RecordEvent(eventName string, message string) {

}

func (n *nopRecorder) RecordErrorEvent(eventName string, message string, err error) {

}

func (n *nopRecorder) Enable() bool {
	return false
}

type recorderImpl struct {
	user      string
	sessionId string
	host      string
}

func (r *recorderImpl) RecordEvent(eventName string, message string) {
	r.RecordErrorEvent(eventName, message, nil)
}

func (r *recorderImpl) RecordErrorEvent(eventName string, message string, err error) {
	event := &Event{
		Timestamp: time.Now(),
		EventName: eventName,
		Message:   message,
		Err:       err,
	}
	sessionEvent := &sessionEvent{
		user:      r.user,
		sessionId: r.sessionId,
		natHost:   r.host,
		Event:     event,
	}

	// 当 trace 日志 channel 超过 90% 时放弃 trace 记录，防止阻塞主业务
	sessionChanCap := cap(sessionEventChan)
	sessionChanLen := len(sessionEventChan)
	if sessionChanLen > sessionChanCap*9/10 {
		log.Run().Errorf("sessionEventChan data to many -> cap:%d len:%d", sessionChanCap, sessionChanLen)
		return
	}
	sessionEventChan <- sessionEvent
}

func (r *recorderImpl) Enable() bool {
	return true
}

var defaultNopRecorder = nopRecorder{}

func acquireRecorder(sessionId string, host, user string, enable bool) Recorder {
	if enable {
		return &recorderImpl{
			user:      user,
			sessionId: sessionId,
			host:      host,
		}
	} else {
		return &defaultNopRecorder
	}

}

type Session struct {
	Recorder Recorder
}

func NewSession(sessionId string, host string, user string, enable bool) *Session {
	if len(sessionId) == 0 {
		sessionId = sessionIdNop
	}
	return &Session{
		Recorder: acquireRecorder(sessionId, host, user, enable),
	}
}
