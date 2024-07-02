package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const TaskSendVerifyEmail = "task:send_verify_email" // เพื่อบอกว่า task ที่ asynq รับไปคืออะไร

type PayloadSendVerifyEmail struct {
	Username string `json:"username"`
}

func (distributor *RedisTaskDistributor) DistributeTaskSendVerifyEmail(
	ctx context.Context,
	payload *PayloadSendVerifyEmail,
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	task := asynq.NewTask(TaskSendVerifyEmail, jsonPayload, opts...) // สร้าง task ด้วย asynq
	info, err := distributor.client.EnqueueContext(ctx, task) // ส่ง task ที่ได้ไปที่ redis queue
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()). // logging
		Str("queue", info.Queue).Int("max_retry", info.MaxRetry).Msg("enqueued task")
	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error { // asynq นั้นจะทำการ pull task มาไว้ที่ task *asynq.Task เรียบร้อยแล้ว
	var payload PayloadSendVerifyEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil { // json.Unmarshal(task.Payload(), &payload) คือเอา payload จาก task ใส่ payload
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	user, err := processor.store.GetUser(ctx, payload.Username)
	if err != nil {
		// เนื่องจากบางครั้งถึงแม้ว่าเราจะกำหนด delay ไปตั้ง 10s แต่บาง case อาจจะไม่พอ เราเลยควรกำหนดให้มันสามารถ retry ได้
		// if err == sql.ErrNoRows {
		// 	return fmt.Errorf("user doesn't exist: %w", asynq.SkipRetry) // ในที่นี้คือ ถ้ามันหา record ไม่เจอ เราจะทำการ return asynq.SkipRetry error เลยทันที ทำให้้มันไม่มีการ retry
		// }
		return fmt.Errorf("failed to get user: %w", err) // return error ตรงนี้ไม่มี asynq.SkipRetry มันเลยจะทำการ retry หลังจาก fail ตรงนี้
	}

	// TODO: send email to user (ยังไม่ทำตอนนี้)

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("email", user.Email).Msg("processed task")
	return nil
}
