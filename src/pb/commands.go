package pb

const (
	CMD_INTERNAL_PLAYER_LOGIN = iota + 2000000
	CMD_INTERNAL_PLAYER_LOGOUT
	CMD_INTERNAL_PLAYER_TO_GAME_MESSAGE
)

type PbInternalCmd int32

const (
	InternalGateToProxy = iota
	InternalProxyToGate
	InternalProxyToGame
	InternalGameToProxy
	InternalProxySync
	InternalRPC
	InternalGameToService
	InternalServiceToGame
)
