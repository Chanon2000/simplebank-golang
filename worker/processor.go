package worker // file นี้จะทำกสนเอา task จาก redis queue แล้ว process มัน

import (
	"context"

	db "github.com/chanon2000/simplebank/db/sqlc"
	"github.com/chanon2000/simplebank/mail"
	"github.com/go-redis/redis/v8"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
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
	mailer mail.EmailSender // EmailSender interface ซึ่งมี method ที่เอาไว้ส่ง email
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, store db.Store, mailer mail.EmailSender) TaskProcessor { // redisOpt เพื่อ connect กับ redis
	logger := NewLogger()
	redis.SetLogger(logger) // เนื่องจากมี logs ที่เกิดจาก go-redis นี้ขึ้น เมื่อเกิด error (เช่นไม่สามารถ connect redis ได้) ซึ่ง format log มันเป็น format ของ package มันเลย เราเลยทำการ format มันให้เหมือนกันให้หมดด้วยใช้การ custom มันโดยใช้ redis.SetLogger นั้นเอง
	// go-redis เป็น package ที่ asynq ใช้ในเบื้องหลังอีกที
	
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{ // asynq.Config{} เอาไว้ควบคุม parameters ต่างๆของ asynq server  
			Queues: map[string]int{
				QueueCritical: 10, // 10 คือ priority values นั้นคือ ให้ task processor เอา task จาก Critical Queue ก่อน Default Queue ก่อน (เพราะ priority สูงกว่า)
				QueueDefault:  5,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) { 
				// function ที่กำหนดใน ErrorHandler จะถูก executed เมื่อ task เกิด error // ในที่นี้ก็คือถ้าเกิดรัน code ใน ProcessTaskSendVerifyEmail function ขึ้น มันก็จะเข้ามารัน function ของ ErrorHandler นั้นเอง
				// ErrorHandler เพื่อกำหนดว่าจะแสดง error ยังไง โดยกำหนด asynq.ErrorHandlerFunc function เข้าไป
				log.Error().Err(err).Str("type", task.Type()). // print log ตรงนี้
					Bytes("payload", task.Payload()).Msg("process task failed")
			}),
			Logger: logger, // Logger field ทำให้เราสามารถกำหนด custom logger ให้กับ Asynq server ได้
			// แต่คุณจะเห็นว่า asynq error logs นั้น มันไม่ได้ print ใน format เดียวกับ zerolog (ทดสอบด้วยการลองหยุดรัน redis ตอนที่ process task อยู่ก็ได้) ซึ่งนี้อาจทำให้ยากต่อการ index logs เข้า monitoring หรือ searching ได้
		},
	)

	return &RedisTaskProcessor{
		server: server,
		store:  store,
		mailer: mailer,
	}
}

func (processor *RedisTaskProcessor) Start() error { // เพื่อบอกให้ asynq นั้น start และบอกว่าให้รัน handler function อันใหน
	mux := asynq.NewServeMux()

	mux.HandleFunc(TaskSendVerifyEmail, processor.ProcessTaskSendVerifyEmail) // นั้นคือการ task ส่ง email ให้รัน ProcessTaskSendVerifyEmail function

	return processor.server.Start(mux)
}
