package global

type ServerType int

const (
	ServerNone ServerType = iota
	ServerGate
	ServerGame
	ServerLogin
	ServerDatabase
)

type ServerNameType struct {
	Name string
	Typ  ServerType
}

type ServerMapAddress struct {
	Servers map[string]*ServerNameType
}

func NewServerMapAddress() *ServerMapAddress {
	return &ServerMapAddress{
		Servers: make(map[string]*ServerNameType),
	}
}

func (s *ServerMapAddress) MapAddrToServerName(addr string, svrType string, name string) {
	if s.Servers != nil && svrType != "" {
		if _, ok := s.Servers[addr]; !ok {
			s.Servers[addr] = &ServerNameType{
				Name: name,
				Typ:  serviceString[svrType],
			}
		}
	}
}

func (s *ServerMapAddress) GetSvrTypeByAddr(addr string) ServerType {
	if s.Servers != nil {
		return s.Servers[addr].Typ
	} else {
		return ServerNone
	}
}
