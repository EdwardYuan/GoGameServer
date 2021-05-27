package global

type ServerType int

const (
	ServerNone ServerType = iota
	ServerGate
	ServerGame
	ServerLogin
	ServerDatabase
)

type ServerMapAddress struct {
	Servers map[string]ServerType
}

func NewServerMapAddress() *ServerMapAddress {
	return &ServerMapAddress{
		Servers: make(map[string]ServerType),
	}
}

func (s *ServerMapAddress) MapAddrToServerName(addr string, svrType ServerType) {
	if s.Servers != nil {
		if _, ok := s.Servers[addr]; !ok {
			s.Servers[addr] = svrType
		}
	}
}

func (s *ServerMapAddress) GetSvrTypeByAddr(addr string) ServerType {
	if s.Servers != nil {
		return s.Servers[addr]
	} else {
		return ServerNone
	}
}
