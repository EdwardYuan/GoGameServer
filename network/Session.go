package network

import "net"

type Session struct {
	id        int
	network   *Network
	processor *Processor
	conn      net.Conn
	reader    MessageReader
	writer    MessageWriter
}
