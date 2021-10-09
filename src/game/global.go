package game

var (
	AgentMgr *AgentManager
)

func InitGlobal() {
	AgentMgr = NewAgentManager()
}
