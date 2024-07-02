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

	// คือทำการส่ง task ไปที่ redis กับ create user นั้นอยู่ใน db transaction เดียวกัน
	// เพราะถ้ามัน CreateUser สำเร็จแต่ดันส่ง email ไม่ได้ (อาจเพราะว่าต่อ redis ไม่ได้) มันจะเกิดปัญหาขึ้น และแก้ค่อนข้างยาก 
	// ทำให้สามารถแก้ด้วยการ send task ไปที่ redis ใน db transaction เดียวกับที่ทำการ insert new user ลง database ทำให้ถ้าเรา fail ในการส่ง task นั้นก็จะทำให้ transaction ถูก rollback เลย ทำให้ client สามารถส่ง retry มาโดยไม่เกิด issue ได้นั้นเอง
	arg := db.CreateUserTxParams{ // อันนี้คือการเตรียม input data สำหรับ transaction นะ ซึ่งก็กำหนด AfterCreate callback function ด้วย
		CreateUserParams: db.CreateUserParams{
			Username:       req.GetUsername(),
			HashedPassword: hashedPassword,
			FullName:       req.GetFullName(),
			Email:          req.GetEmail(),
		},
		AfterCreate: func(user db.User) error {
			taskPayload := &worker.PayloadSendVerifyEmail{
				Username: user.Username,
			}
			opts := []asynq.Option{
				asynq.MaxRetry(10), // retry ไม่เกิน 10 ครั้งถ้ามันเกิด fails
				asynq.ProcessIn(10 * time.Second), // คือเพิ่ม deley ให้กับ task // คือกำหนด delay เป็น 10s หมายความว่า task จะถูก picked up โดย worker หลังจากผ่านไป 10s // 
				// asynq.Queue("critical"), // ถ้าคุณมี multiple tasks ที่มี priority level ที่แตกต่างกัน คุณสามารถใช้ asynq.Queue() option เพื่อกำหนด queues ส่งไป เช่นในครั้งนี้กำหนดเป็น "critical" // เรากำหนดเป็น "critical" คุณต้องไปบอก task processor ด้วยว่าให้เอา task จาก critical queues (มันจะเอาจาก "default" queue เป็น default)
				asynq.Queue(worker.QueueCritical), // ใส่เป็น const แทนใส่ string ตรงๆนี้กว่า
			}

			return server.taskDistributor.DistributeTaskSendVerifyEmail(ctx, taskPayload, opts...) // return error ไปที่ caller ของ callback function นี้เลย
		},
	}

	txResult, err := server.store.CreateUserTx(ctx, arg) // AfterCreate จะถูก call หลังจากที่ create user เสร็จ (ดูใน code ของ CreateUserTx ได้)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				return nil, status.Errorf(codes.AlreadyExists, "username already exists: %s", err)
			}
		}
		return nil, status.Errorf(codes.Internal, "failed to create user: %s", err)
	}

	rsp := &pb.CreateUserResponse{
		User: convertUser(txResult.User),
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
