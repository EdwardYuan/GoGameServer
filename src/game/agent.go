package game

const MaxAgent = 10000

type Agent struct {
	OpenId string
	P      *Player
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
