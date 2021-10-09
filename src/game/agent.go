package game

import "GoGameServer/src/lib"

const MaxAgent = 10000

type Agent struct {
	OpenId    string
	P         *Player
	CloseChan chan int
}

func (a *Agent) Run() {
	for {
		select {
		case <-a.CloseChan:
			return
		}
	}
}

type AgentManager struct {
	Agents map[int64]*Agent
}

func NewAgentManager() *AgentManager {
	return &AgentManager{
		Agents: make(map[int64]*Agent),
	}
}

func (am *AgentManager) GetAgent(playerId int64) *Agent {
	if am.Agents != nil {
		return am.Agents[playerId]
	}
	return nil
}

func (am *AgentManager) GetPlayer(playerId int64) *Player {
	if am.Agents != nil {
		return am.Agents[playerId].P
	}
	return nil
}

func (am *AgentManager) NewAgent(s *lib.Session, playerId int64) *Agent {
	agent := Agent{}
	am.Agents[playerId] = &agent
	return &agent
}
