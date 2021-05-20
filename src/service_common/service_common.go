package service_common

import (
	"GoGameServer/src/MsgHandler"
	"GoGameServer/src/lib"
	"github.com/panjf2000/gnet"
	"github.com/streadway/amqp"
)

type Service interface {
	Start() (err error)
	Stop()
	Run()
}

type ServerCommon struct {
	Name      string
	Id        int
	CloseChan chan int
	*gnet.EventServer
}

func FailOnError(err error, msg string) {
	if err != nil {
		lib.Logger.Fatal(msg)
	}
}

func (s *ServerCommon) StartRabbit() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	FailOnError(err, "Connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	FailOnError(err, "Rabbit Make channel error")
	err = ch.ExchangeDeclare(
		s.Name,
		"fanout",
		false,
		false,
		false,
		false,
		nil,
	)
	FailOnError(err, "Declare Exchange error")

}

func (s *ServerCommon) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {

	return
}

func (s *ServerCommon) Encode(msg MsgHandler.Message) (err error) {
	//TODO protobuf marshal
	return
}

func (s *ServerCommon) Decode(data []byte) (msg MsgHandler.Message, err error) {
	//TODO protobuf unmarshal
	return
}

func (s *ServerCommon) HandleMessage(msg MsgHandler.Message) {

}
