package worker // file นี้เพื่อ create tasks และ distribute ไปที่ workers ผ่าน redis queue

import (
	"context"

	"github.com/hibiken/asynq"
)

// TaskDistributor, RedisTaskDistributor นั้นจะคล้ายๆกับที่เราทำให้ database layer ที่เรามี Store interface และ SQLStore struct
type TaskDistributor interface {
	DistributeTaskSendVerifyEmail( // ใส่ DistributeTaskSendVerifyEmail ลงใน TaskDistributor interface
		ctx context.Context,
		payload *PayloadSendVerifyEmail,
		opts ...asynq.Option,
	) error
}

type RedisTaskDistributor struct {
	client *asynq.Client
}

func NewRedisTaskDistributor(redisOpt asynq.RedisClientOpt) TaskDistributor { // เพื่อสร้าง new redis task distributor
	client := asynq.NewClient(redisOpt)
	return &RedisTaskDistributor{ // จะเห็นว่าเรา return เป็น RedisTaskDistributor แต่ กำหนด return type เป็น TaskDistributor interface 
		// เป็นวิธีเพื่อที่จะทำให้ ถ้า RedisTaskDistributor ไม่ implement ทุก required function ใน TaskDistributor interface ก็จะทำให้ compiler แจ้ง error ทันที
		client: client,
	}
}
