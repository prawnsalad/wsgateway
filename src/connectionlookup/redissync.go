package connectionlookup

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type RedisSync struct {
	client *redis.Client
}

func NewRedisSync(url string) (*RedisSync, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	log.Printf("Connecting to redis for connection data at %s", opts.Addr)
	rClient := redis.NewClient(opts)
	res := rClient.Ping(ctx)
	if res.Err() != nil {
		return nil, fmt.Errorf("error conencting to redis: %v", res.Err().Error())
	}

	sync := &RedisSync{
		client: rClient,
	}

	return sync, nil
}

func (s *RedisSync) RemoveConnection(con *Connection) {
	s.client.Del(
		ctx,
		fmt.Sprintf("connection:%s", con.Id),
		fmt.Sprintf("connection:%s:tags", con.Id),
	)
}

func (s *RedisSync) AddConnection(con *Connection) {
	res := s.client.HSet(ctx, fmt.Sprintf("connection:%s", con.Id), map[string]string{
		"addr": con.Socket.RemoteAddr().String(),
	})
	if res.Err() != nil {
		log.Println("redis error:", res.Err().Error())
	}
}

func (s *RedisSync) UpdateConnectionTags(con *Connection, tags map[string]string) {
	res := s.client.HSet(ctx, fmt.Sprintf("connection:%s:tags", con.Id), tags)
	if res.Err() != nil {
		log.Println("redis error:", res.Err().Error())
	}
}

func (s *RedisSync) RemoveConnectionTags(con *Connection, tags []string) {
	res := s.client.HDel(ctx, fmt.Sprintf("connection:%s:tags", con.Id), tags...)
	if res.Err() != nil {
		log.Println("redis error:", res.Err().Error())
	}
}
