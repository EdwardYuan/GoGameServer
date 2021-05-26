package global

import "net"

type ServerType int

const (
	ServerNone ServerType = iota
	ServerGate
	ServerGame
	ServerLogin
	ServerDatabase
)

type ServerMapAddress struct {
	Servers map[net.Addr]ServerType
}

func NewServerMapAddress() *ServerMapAddress {
	return &ServerMapAddress{
		Servers: make(map[net.Addr]ServerType),
	}
}

func (s *ServerMapAddress) MapAddrToServerName(addr net.Addr, svrType ServerType) {
	if s.Servers != nil {
		if _, ok := s.Servers[addr]; !ok {
			s.Servers[addr] = svrType
		}
	}
}

func (s *ServerMapAddress) GetSvrTypeByAddr(addr net.Addr) ServerType {
	if s.Servers != nil {
		return s.Servers[addr]
	} else {
		return ServerNone
	}
}
