package game

import (
	"time"

	"GoGameServer/network"
	"GoGameServer/src/lib"
)

const MaxAgent = 10000

type Agent struct {
	OpenId    string
	P         *Player
	Recv      chan lib.SeqEvent
	CloseChan chan int
}

func (a *Agent) HandleEvent(se lib.SeqEvent) chan error {
	a.handle(se)
	return nil
}

func (a *Agent) handle(se lib.SeqEvent) {

}

func (a *Agent) Run() {
	for {
		select {
		case se, ok := <-a.Recv:
			timeout := time.After(10 * time.Second)
			if !ok {
				panic("agent in game, a.recv has been closed")
			}
			/*
				if a.ignoreExternal(se.Event) {
					a.Warning("ignoreExternal", se)
					continue
				}
			*/
			select {
			case <-timeout:
				// Todo 超时处理
			case err := <-a.HandleEvent(se):
				if err != nil {

				}
			}
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

func (am *AgentManager) NewAgent(s *network.Session, playerId int64) *Agent {
	agent := Agent{}
	am.Agents[playerId] = &agent
	return &agent
}
