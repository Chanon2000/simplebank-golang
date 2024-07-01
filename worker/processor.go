package worker // file นี้จะทำกสนเอา task จาก redis queue แล้ว process มัน

import (
	"context"

	db "github.com/chanon2000/simplebank/db/sqlc"
	"github.com/hibiken/asynq"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
)

type TaskProcessor interface {
	Start() error
	ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error
}

type RedisTaskProcessor struct {
	server *asynq.Server
	store  db.Store
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, store db.Store) TaskProcessor { // redisOpt เพื่อ connect กับ redis
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{}, // asynq.Config{} เอาไว้ควบคุม parameters ต่างๆของ asynq server // ตอนนี้ว่างๆไปก่อน
	)

	return &RedisTaskProcessor{
		server: server,
		store:  store,
	}
}

func (processor *RedisTaskProcessor) Start() error { // เพื่อบอกให้ asynq นั้น start และบอกว่าให้รัน handler function อันใหน
	mux := asynq.NewServeMux()

	mux.HandleFunc(TaskSendVerifyEmail, processor.ProcessTaskSendVerifyEmail) // นั้นคือการ task ส่ง email ให้รัน ProcessTaskSendVerifyEmail function

	return processor.server.Start(mux)
}
