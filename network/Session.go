package network

type Session struct {
	id        int
	network   *Network
	processor *NetworkProcessor
}
