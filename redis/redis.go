package redis

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redis_rate/v9"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/zhlls/go-common/log"
)

var (
	Client         redis.UniversalClient
	redisOPLatency prometheus.ObserverVec
)

const (
	ClusterModeSingle   = "Single"
	ClusterModeCluster  = "Cluster"
	ClusterModeSentinel = "Sentinel"
	MetricNamespace     = "notifier"
)

type Options struct {
	Registerer prometheus.Registerer

	Endpoints          []string
	Password           string
	ClusterMode        string
	SentinelMasterName string
	DB                 int
	PoolSize           int
	MinIdleConns       int
}

func NewRedisClient(opts Options) (redis.UniversalClient, error) {
	if len(opts.Endpoints) == 0 {
		return nil, errors.New("invalid redis endpoint")
	}
	redis.SetLogger(logger{})
	switch opts.ClusterMode {
	case ClusterModeSingle:
		Client = redis.NewClient(&redis.Options{
			Addr:         opts.Endpoints[0],
			Password:     opts.Password,
			DB:           opts.DB,
			PoolSize:     opts.PoolSize,
			MinIdleConns: opts.MinIdleConns,
		})
	case ClusterModeCluster:
		Client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        opts.Endpoints,
			Password:     opts.Password,
			PoolSize:     opts.PoolSize,
			MinIdleConns: opts.MinIdleConns,
		})
	case ClusterModeSentinel:
		if opts.SentinelMasterName == "" {
			return nil, errors.New("redis sentinel no master")
		}
		Client = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    opts.SentinelMasterName,
			SentinelAddrs: opts.Endpoints,
			Password:      opts.Password,
			DB:            opts.DB,
			PoolSize:      opts.PoolSize,
			MinIdleConns:  opts.MinIdleConns,
		})
	default:
		return nil, errors.New("invalid redis mode")
	}

	_, err := Client.Ping(Client.Context()).Result()
	if err != nil {
		return nil, err
	}

	//addQueueLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	//	Namespace: utils.MetricNamespace,
	//	Name:      "add_queue_latency_seconds",
	//	Help:      "The latency of alerts add to queue.",
	//	Buckets:   []float64{0.001, 0.002, 0.005, 0.01, 0.05, 0.1, 1},
	//}, []string{"op", "key"})
	redisOPLatency = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace:  MetricNamespace,
		Name:       "redis_op_latency_seconds",
		Help:       "The latency of add something to redis.",
		Objectives: map[float64]float64{0.5: 0.05, 0.95: 0.005, 0.99: 0.001},
	}, []string{"op", "key"})
	opts.Registerer.MustRegister(redisOPLatency)

	return Client, nil
}

func Close() error {
	return Client.Close()
}

func Set(key string, value interface{}, max_life int64) error {
	now := time.Now()
	cmd := Client.SetEX(Client.Context(), key, value, time.Duration(max_life)*time.Second)
	redisOPLatency.WithLabelValues("set", key).Observe(time.Since(now).Seconds())
	return cmd.Err()
}

func Get(key string) (string, error) {
	now := time.Now()
	cmd := Client.Get(Client.Context(), key)
	redisOPLatency.WithLabelValues("get", key).Observe(time.Since(now).Seconds())
	return cmd.Val(), cmd.Err()
}

func AddQueue(key string, values ...interface{}) error {
	now := time.Now()
	cmd := Client.LPush(Client.Context(), key, values...)
	redisOPLatency.WithLabelValues("lpush", key).Observe(time.Since(now).Seconds())
	return cmd.Err()
}

func FetchQueue(ctx context.Context, timeout time.Duration, keys ...string) (string, string, error) {
	cmd := Client.BRPop(ctx, timeout, keys...)
	if cmd.Err() != nil {
		return "", "", cmd.Err()
	}
	return cmd.Val()[0], cmd.Val()[1], nil
}

func QueueLen(ctx context.Context, key string) (int64, error) {
	cmd := Client.LLen(ctx, key)
	if cmd.Err() != nil {
		return 0, cmd.Err()
	}
	return cmd.Val(), nil
}

func NewRateLimiter() *redis_rate.Limiter {
	return redis_rate.NewLimiter(Client)
}

func Expire(key string, max_life int64) error {
	now := time.Now()
	cmd := Client.Expire(Client.Context(), key, time.Duration(max_life)*time.Second)
	redisOPLatency.WithLabelValues("expire", key).Observe(time.Since(now).Seconds())
	return cmd.Err()
}

func Delete(keys ...string) error {
	now := time.Now()
	cmd := Client.Del(Client.Context(), keys...)
	redisOPLatency.WithLabelValues("delete", strings.Join(keys, "/")).Observe(time.Since(now).Seconds())
	return cmd.Err()
}

type logger struct{}

func (l logger) Printf(_ context.Context, format string, v ...interface{}) {
	log.Infof(format, v...)
}
