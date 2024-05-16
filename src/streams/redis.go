package streams

import (
	"context"
	"fmt"
	"log"

	"com.wsgateway/connectionlookup"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type StreamRedis struct {
	client     *redis.Client
	streamName string
}

func NewStreamRedis(redisUrl string, streamName string) (*StreamRedis, error) {
	opts, err := redis.ParseURL(redisUrl)
	if err != nil {
		return nil, err
	}

	log.Printf("Connecting to redis for streaming at %s", opts.Addr)
	rClient := redis.NewClient(opts)
	res := rClient.Ping(ctx)
	if res.Err() != nil {
		return nil, fmt.Errorf("error conencting to redis: %v", res.Err().Error())
	}

	sync := &StreamRedis{
		client:     rClient,
		streamName: streamName,
	}

	return sync, nil
}

func (s *StreamRedis) PublishConnection(con *connectionlookup.Connection, event StreamEvent) {
	res := s.client.XAdd(ctx, &redis.XAddArgs{
		Stream: "connectionevents",
		Values: map[string]string{
			"connection": con.Id,
			"action":     event.String(),
			"tags":       makeTagString(con),
		},
	})

	if res.Err() != nil {
		log.Println("redis stream error:", res.Err().Error())
	}
}

func (s *StreamRedis) PublishMessage(con *connectionlookup.Connection, messageType MessageType, message []byte) {
	msgStr := string(message)
	streamName := replaceConnectionVars(s.streamName, msgStr, *con.JsonExtractVars)

	res := s.client.XAdd(ctx, &redis.XAddArgs{
		Stream: streamName,
		Values: map[string]string{
			"connection": con.Id,
			"action":     EventMessage.String(),
			"type":       messageType.String(),
			"tags":       makeTagString(con),
			"message":    msgStr,
		},
	})

	if res.Err() != nil {
		log.Println("redis stream error:", res.Err().Error())
	}
}
