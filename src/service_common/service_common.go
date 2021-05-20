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
	Name       string
	Id         int
	CloseChan  chan int
	RabbitChan chan bool
	*gnet.EventServer
}

func FailOnError(err error, msg string) {
	if err != nil {
		lib.Logger.Fatal(msg)
	}
}

func (s *ServerCommon) StartRabbit() {
	s.RabbitChan = make(chan bool)
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	defer conn.Close()
	FailOnError(err, "Connect to RabbitMQ")

	<-s.RabbitChan
	ch, err := conn.Channel()
	defer ch.Close()
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
	queue, err := ch.QueueDeclare(s.Name, false, true, false, false, nil)
	FailOnError(err, "Declare RabbitMQ queue Error")

	err = ch.ExchangeBind(
		queue.Name,
		queue.Name,
		queue.Name,
		false,
		nil,
	)
	FailOnError(err, "Declare Bind Error")

	ch.Consume(queue.Name, "", true, false, false, false, nil)

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
