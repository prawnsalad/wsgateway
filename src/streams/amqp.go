package streams

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"com.wsgateway/connectionlookup"
	"github.com/streadway/amqp"
)

type StreamAmqp struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	exchangeName string
	exchangeType string
	routingKey   string
}

func NewStreamAmqp(amqpUrl, exchangeName, exchangeType, routingKey string) (*StreamAmqp, error) {
	log.Printf("Connecting via AMQP for streaming at %s", amqpUrl)
	amqpConn, err := amqp.Dial(amqpUrl)
	if err != nil {
		return nil, fmt.Errorf("error conencting via AMQP: %v", err)
	}

	amqpChan, err := amqpConn.Channel()
	if err != nil {
		return nil, fmt.Errorf("error creating AMQP channel: %v", err)
	}

	if exchangeType != "" {
		err = amqpChan.ExchangeDeclare(exchangeName, exchangeType, true, false, false, false, nil)
		if err != nil {
			return nil, fmt.Errorf("error declaring AMQP exchange: %v", err)
		}
	}

	sync := &StreamAmqp{
		conn:         amqpConn,
		channel:      amqpChan,
		exchangeName: exchangeName,
		exchangeType: exchangeType,
		routingKey:   routingKey,
	}

	return sync, nil
}

func (s *StreamAmqp) PublishConnection(con *connectionlookup.Connection, event StreamEvent) {
	routingKey := replaceConnectionVars(s.routingKey, "", *con.JsonExtractVars, con.TagsAsMap())

	body, err := json.Marshal(map[string]string{
		"connection": con.Id,
		"action":     event.String(),
		"tags":       makeTagString(con),
	})
	if err != nil {
		log.Println("AMQP JSON encoding error:", err)
	}

	err = s.channel.Publish(s.exchangeName, routingKey, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Timestamp:    time.Now(),
		Body:         body,
	})
	if err != nil {
		log.Println("AMQP publish error:", err)
	}
}

func (s *StreamAmqp) PublishMessage(con *connectionlookup.Connection, messageType MessageType, message []byte) {
	msgStr := string(message)
	routingKey := replaceConnectionVars(s.routingKey, msgStr, *con.JsonExtractVars, con.TagsAsMap())

	body, err := json.Marshal(map[string]string{
		"connection": con.Id,
		"action":     EventMessage.String(),
		"type":       messageType.String(),
		"tags":       makeTagString(con),
		"message":    msgStr,
	})
	if err != nil {
		log.Println("AMQP JSON encoding error:", err)
	}

	err = s.channel.Publish(s.exchangeName, routingKey, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Timestamp:    time.Now(),
		Body:         body,
	})
	if err != nil {
		log.Println("AMQP publish error:", err)
	}
}
