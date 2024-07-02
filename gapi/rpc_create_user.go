package gapi

import (
	"context"
	"time"

	db "github.com/chanon2000/simplebank/db/sqlc"
	"github.com/chanon2000/simplebank/pb"
	"github.com/chanon2000/simplebank/util"
	"github.com/chanon2000/simplebank/val"
	"github.com/chanon2000/simplebank/worker"
	"github.com/hibiken/asynq"
	"github.com/lib/pq"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	violations := validateCreateUserRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	hashedPassword, err := util.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %s", err)
	}

	arg := db.CreateUserParams{
		Username:       req.GetUsername(),
		HashedPassword: hashedPassword,
		FullName:       req.GetFullName(),
		Email:          req.GetEmail(),
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				return nil, status.Errorf(codes.AlreadyExists, "username already exists: %s", err)
			}
		}
		return nil, status.Errorf(codes.Internal, "failed to create user: %s", err)
	}

	// TODO: use db transaction (ตรงนี้เราควรจะ create user และ ส่ง task ใน transaction เดียว เพื่อที่ว่าถ้า operation ใหน fail ก็จะได้ rollback ทั้ง 2 เลยได้) (ทำให้ lecture หน้า)

	taskPayload := &worker.PayloadSendVerifyEmail{
		Username: user.Username,
	}
	opts := []asynq.Option{
		asynq.MaxRetry(10), // retry ไม่เกิน 10 ครั้งถ้ามันเกิด fails
		asynq.ProcessIn(10 * time.Second), // คือเพิ่ม deley ให้กับ task
		// asynq.Queue("critical"), // ถ้าคุณมี multiple tasks ที่มี priority level ที่แตกต่างกัน คุณสามารถใช้ asynq.Queue() option เพื่อกำหนด queues ส่งไป เช่นในครั้งนี้กำหนดเป็น "critical" // เรากำหนดเป็น "critical" คุณต้องไปบอก task processor ด้วยว่าให้เอา task จาก critical queues (มันจะเอาจาก "default" queue เป็น default)
		asynq.Queue(worker.QueueCritical), // ใส่เป็น const แทนใส่ string ตรงๆนี้กว่า
	}
	err = server.taskDistributor.DistributeTaskSendVerifyEmail(ctx, taskPayload, opts...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to distribute task to send verify email: %s", err)
	}

	rsp := &pb.CreateUserResponse{
		User: convertUser(user),
	}
	return rsp, nil
}

func validateCreateUserRequest(req *pb.CreateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}

	if err := val.ValidatePassword(req.GetPassword()); err != nil {
		violations = append(violations, fieldViolation("password", err))
	}

	if err := val.ValidateFullName(req.GetFullName()); err != nil {
		violations = append(violations, fieldViolation("full_name", err))
	}

	if err := val.ValidateEmail(req.GetEmail()); err != nil {
		violations = append(violations, fieldViolation("email", err))
	}

	return violations
}
