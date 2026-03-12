package network

import (
	"GoGameServer/src/lib"
	"go.uber.org/zap"
	"net"
	"time"
)

// SessionCloseReason 会话关闭原因
type SessionCloseReason string

const (
	NETWORK_SESSION_CLOSED_BY_LOCAL             SessionCloseReason = "local_closed"
	NETWORK_SESSION_CLOSED_BY_REMOTE            SessionCloseReason = "remote_closed"
	NETWORK_SESSION_CLOSED_BY_HEARTBEAT_TIMEOUT SessionCloseReason = "heartbeat_timeout"
	NETWORK_SESSION_CLOSED_BY_READ_ERROR        SessionCloseReason = "read_error"
	NETWORK_SESSION_CLOSED_BY_WRITE_ERROR       SessionCloseReason = "write_error"
	NETWORK_SESSION_CLOSED_BY_SHUTDOWN          SessionCloseReason = "shutdown"
)

type Session struct {
	id            int
	network       *Network
	processor     *Processor
	conn          net.Conn
	closed        *lib.AtomBool
	lastReadTime  int64
	lastWriteTime int64
	closeChan     chan SessionCloseReason
	//config        map[string]string
	logger *zap.SugaredLogger
}

func NewSession(id int, network *Network, c net.Conn) *Session {
	session := &Session{
		id:            id,
		conn:          c,
		lastReadTime:  time.Now().Unix(),
		lastWriteTime: time.Now().Unix(),
		closed:        &lib.AtomBool{},
		network:       network,
	}
	session.processor = NewProcessor(session)
	return session
}
