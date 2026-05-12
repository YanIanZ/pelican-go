package jobs

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Queue struct {
	Client    *asynq.Client
	Server    *asynq.Server
	Inspector *asynq.Inspector
	DB        *gorm.DB
}

func NewQueue(rdb *redis.Client, db *gorm.DB) *Queue {
	opt := asynq.RedisClientOpt{Addr: rdb.Options().Addr}
	client := asynq.NewClient(opt)
	server := asynq.NewServer(opt, asynq.Config{Concurrency: 10})
	inspector := asynq.NewInspector(opt)
	return &Queue{Client: client, Server: server, Inspector: inspector, DB: db}
}

func (q *Queue) RegisterHandlers(mux *asynq.ServeMux) {
	mux.HandleFunc("server:install", q.HandleServerInstall)
	mux.HandleFunc("backup:process", q.HandleBackupProcess)
	mux.HandleFunc("webhook:dispatch", q.HandleWebhookDispatch)
}

func (q *Queue) Start() error {
	mux := asynq.NewServeMux()
	q.RegisterHandlers(mux)
	return q.Server.Run(mux)
}

func (q *Queue) Enqueue(ctx context.Context, typename string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = q.Client.Enqueue(asynq.NewTask(typename, data))
	return err
}
