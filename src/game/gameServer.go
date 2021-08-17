package game

type GSServer struct {
	Agents map[int64]*Agent
}

func NewGsServer() *GSServer {
	return &GSServer{
		Agents: make(map[int64]*Agent),
	}
}
