package streams

import (
	"container/list"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"com.wsgateway/connectionlookup"
	amqp "github.com/rabbitmq/amqp091-go"
)

type AmqpEvent struct {
	routingKey string
	message    amqp.Publishing
}
type StreamAmqp struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	exchangeName string
	routingKey   string
	backlogChan  chan AmqpEvent
	backlog      *list.List
}

func NewStreamAmqp(amqpUrl, exchangeName, queueName, routingKey string) (*StreamAmqp, error) {
	log.Printf("Connecting via AMQP for streaming at %s", amqpUrl)
	amqpConn, err := amqp.Dial(amqpUrl)
	if err != nil {
		return nil, fmt.Errorf("error conencting via AMQP: %v", err)
	}

	amqpChan, err := amqpConn.Channel()
	if err != nil {
		return nil, fmt.Errorf("error creating AMQP channel: %v", err)
	}

	if exchangeName != "" {
		err = amqpChan.ExchangeDeclare(exchangeName, "topic", true, false, false, false, nil)
		if err != nil {
			return nil, fmt.Errorf("error declaring AMQP exchange: %v", err)
		}
	}

	if queueName != "" {
		queue, err := amqpChan.QueueDeclare(queueName, true, false, false, false, nil)
		if err != nil {
			return nil, fmt.Errorf("error declaring AMQP queue: %v", err)
		}

		err = amqpChan.QueueBind(queue.Name, "#", exchangeName, false, nil)
		if err != nil {
			return nil, fmt.Errorf("error binding AMQP queue to exchange: %v", err)
		}
	}

	sync := &StreamAmqp{
		conn:         amqpConn,
		channel:      amqpChan,
		exchangeName: exchangeName,
		routingKey:   routingKey,
		backlogChan:  make(chan AmqpEvent, 100),
	}

	go sync.Publisher()

	return sync, nil
}

func (s *StreamAmqp) Publisher() {
	backlogLog := sync.RWMutex{}
	backlog := list.New()
	retryCnt := 0

	go func() {
		for event := range s.backlogChan {
			backlogLog.Lock()
			backlog.PushBack(event)
			backlogLog.Unlock()
			
		}
	}()

	for {
		backlogLog.RLock()
		e := backlog.Front()
		backlogLog.RUnlock()
		if e == nil {
			// TODO: Find a way to get rid of this sleep
			time.Sleep(1 * time.Millisecond)
			continue
		}

		event := e.Value.(AmqpEvent)
		println("Publishing event", s.exchangeName, event.routingKey, string(event.message.Body))
		for {
			err := s.channel.Publish(s.exchangeName, event.routingKey, false, false, event.message)
			if err != nil {
				retrySec := backoff(float64(retryCnt), 10)

				backlogLog.Lock()
				backlogLen := backlog.Len()
				backlogLog.Unlock()

				log.Printf("AMQP publish error, retrying in %v seconds with %d outstanding events: %s", retrySec, backlogLen, err.Error())
				time.Sleep(time.Second * time.Duration(retrySec))
				retryCnt++
				continue
			}

			retryCnt = 0
			break
		}

		backlogLog.Lock()
		backlog.Remove(e)
		backlogLog.Unlock()
	}
}

func backoff(retryCnt, maxLen float64) float64 {
	len := math.Pow(2, float64(retryCnt))
	if len > maxLen {
		len = maxLen
	}
	return len
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

	s.backlogChan <- AmqpEvent{
		routingKey: routingKey,
		message: amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Timestamp:    time.Now(),
			Body:         body,
		},
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

	s.backlogChan <- AmqpEvent{
		routingKey: routingKey,
		message: amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Timestamp:    time.Now(),
			Body:         body,
		},
	}
}
