package redis

import (
	"fmt"
	"strings"

	red "github.com/go-redis/redis"
)

const (
	ClusterType = "cluster"
	NodeType    = "node"
)

type Redis struct {
	Addr   string
	Type   string
	Pass   string
	client red.Cmdable
	closer interface {
		Close() error
	}
}

func NewRedis(redisAddr, redisType string, redisPass ...string) *Redis {
	pass := ""
	if len(redisPass) > 0 {
		pass = redisPass[0]
	}
	if redisType == "" {
		redisType = NodeType
	}

	store := &Redis{
		Addr: redisAddr,
		Type: redisType,
		Pass: pass,
	}

	switch redisType {
	case ClusterType:
		client := red.NewClusterClient(&red.ClusterOptions{
			Addrs:    splitAddrs(redisAddr),
			Password: pass,
		})
		store.client = client
		store.closer = client
	default:
		client := red.NewClient(&red.Options{
			Addr:     redisAddr,
			Password: pass,
			DB:       0,
		})
		store.client = client
		store.closer = client
	}

	return store
}

func (s *Redis) Hkeys(key string) ([]string, error) {
	return s.client.HKeys(key).Result()
}

func (s *Redis) Hset(key, field, value string) error {
	return s.client.HSet(key, field, value).Err()
}

func (s *Redis) Hmset(key string, fieldsAndValues map[string]string) error {
	values := make(map[string]interface{}, len(fieldsAndValues))
	for field, value := range fieldsAndValues {
		values[field] = value
	}
	return s.client.HMSet(key, values).Err()
}

func (s *Redis) Hgetall(key string) (map[string]string, error) {
	return s.client.HGetAll(key).Result()
}

func (s *Redis) Sadd(key string, values ...interface{}) (int, error) {
	expanded := expandValues(values)
	if len(expanded) == 0 {
		return 0, nil
	}
	count, err := s.client.SAdd(key, expanded...).Result()
	return int(count), err
}

func (s *Redis) SpopN(key string, count int64) ([]string, error) {
	return s.client.SPopN(key, count).Result()
}

func (s *Redis) Ping() bool {
	val, err := s.client.Ping().Result()
	return err == nil && val == "PONG"
}

func (s *Redis) Close() error {
	if s.closer == nil {
		return nil
	}
	return s.closer.Close()
}

func splitAddrs(addrs string) []string {
	var out []string
	for _, addr := range strings.Split(addrs, ",") {
		addr = strings.TrimSpace(addr)
		if addr != "" {
			out = append(out, addr)
		}
	}
	if len(out) == 0 {
		out = append(out, addrs)
	}
	return out
}

func expandValues(values []interface{}) []interface{} {
	var expanded []interface{}
	for _, value := range values {
		switch v := value.(type) {
		case []string:
			for _, item := range v {
				expanded = append(expanded, item)
			}
		case []interface{}:
			expanded = append(expanded, v...)
		default:
			expanded = append(expanded, value)
		}
	}
	return expanded
}

func (s *Redis) String() string {
	return fmt.Sprintf("redis://%s/%s", s.Type, s.Addr)
}
