package redis

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	redisV8 "github.com/go-redis/redis/v8"
)

type Subscriber interface {
	Run()
	Stop()
}

type notifier struct {
	logger log.Logger

	key       string
	processor func(string)

	done   chan struct{}
	ctx    context.Context
	cancel func()
}

func (n *notifier) Run() {
	timeout := 5 * time.Second

	n.done = make(chan struct{})
	defer close(n.done)
	n.ctx, n.cancel = context.WithCancel(context.Background())

	for {
		select {
		case <-n.ctx.Done():
			return
		default:
			_, v, err := FetchQueue(n.ctx, timeout, n.key)
			if err != nil {
				if err == redisV8.Nil || err == context.Canceled {
					continue
				}
				level.Warn(n.logger).Log("msg", "FetchQueue error", "type", n.key, "err", err)
				continue
			}
			n.processor(v)
		}
	}
}

func (n *notifier) Stop() {
	if n.cancel == nil {
		return
	}
	n.cancel()
	n.cancel = nil
	<-n.done
}

func Subscribe(logger log.Logger, key string, processor func(string)) Subscriber {
	n := notifier{
		logger: logger,
		key: key,
		processor: processor,
	}
	return &n
}
