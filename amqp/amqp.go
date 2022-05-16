package amqp

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/brunoos/cnterra-loader/config"
)

type NodeData struct {
	NodeID  int    `json:"nodeid"`
	Payload string `json:"payload"`
}

var channel *amqp.Channel

//------------------------------------------------------------------------------

func SendData(payload string) error {
	res := NodeData{
		NodeID:  config.NodeID,
		Payload: payload,
	}

	buffer, err := json.Marshal(&res)
	if err != nil {
		return err
	}

	msg := amqp.Publishing{
		ContentType: "text/json",
		Body:        buffer,
	}

	return channel.Publish(
		config.NodeEx, // exchange
		"data.out",    // routing key
		false,         // mandatory
		false,         // immediate
		msg,           // body
	)
}

//------------------------------------------------------------------------------

func Initialize() {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		config.RbUser, config.RbPassword, config.RbAddress, config.RbPort)

	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf("[ERRO] Error connecting to RabbitMQ: %s", err)
	}

	channel, err = conn.Channel()
	if err != nil {
		log.Fatalf("[ERRO] Error openning a channel: %s", err)
	}

	err = channel.ExchangeDeclare(
		config.NodeEx, // name
		"topic",       // type
		true,          // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		log.Fatalf("[ERRO] Error declaring the controller exchange: %s", err)
	}
}
