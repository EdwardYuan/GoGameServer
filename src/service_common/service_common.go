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
	rabbitConn *amqp.Connection
	RabbitChan *amqp.Channel
	*gnet.EventServer
}

func FailOnError(err error, msg string) {
	if err != nil {
		lib.Logger.Fatal(msg)
	}
}

func (s *ServerCommon) Stop() {
	defer s.RabbitChan.Close()
	defer s.rabbitConn.Close()

}

func (s *ServerCommon) StartRabbit() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	defer conn.Close()
	FailOnError(err, "Connect to RabbitMQ")

	s.RabbitChan, err = conn.Channel()
	FailOnError(err, "Rabbit Make channel error")

	err = s.RabbitChan.ExchangeDeclare(
		s.Name,
		"fanout",
		false,
		false,
		false,
		false,
		nil,
	)
	FailOnError(err, "Declare Exchange error")
	queue, err := s.RabbitChan.QueueDeclare(s.Name, false, true, false, false, nil)
	FailOnError(err, "Declare RabbitMQ queue Error")

	err = s.RabbitChan.ExchangeBind(
		queue.Name,
		queue.Name,
		queue.Name,
		false,
		nil,
	)
	FailOnError(err, "Declare Bind Error")
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
