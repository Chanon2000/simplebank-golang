package worker

import (
	"context"
	"encoding/json"
	"fmt"

	db "github.com/chanon2000/simplebank/db/sqlc"
	"github.com/chanon2000/simplebank/util"
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

	verifyEmail, err := processor.store.CreateVerifyEmail(ctx, db.CreateVerifyEmailParams{ // สร้าง VerifyEmail เก็บลง database
		Username:   user.Username,
		Email:      user.Email,
		SecretCode: util.RandomString(32), // SecretCode ยิ่งยาวยิ่งดีเพื่อป้องกัน brute-force attacks
	})
	if err != nil {
		return fmt.Errorf("failed to create verify email: %w", err)
	}

	// กำหนดสิ่งต่างๆใน email
	subject := "Welcome to Simple Bank"
	verifyUrl := fmt.Sprintf("http://localhost:8080/v1/verify_email?email_id=%d&secret_code=%s",
		verifyEmail.ID, verifyEmail.SecretCode) // TODO: replace this URL with an environment variable that points to a front-end page
	content := fmt.Sprintf(`Hello %s,<br/>
	Thank you for registering with us!<br/>
	Please <a href="%s">click here</a> to verify your email address.<br/>
	`, user.FullName, verifyUrl)
	to := []string{user.Email}
	
	// ส่ง verify email
	err = processor.mailer.SendEmail(subject, content, to, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to send verify email: %w", err)
	}

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("email", user.Email).Msg("processed task")
	return nil
}
