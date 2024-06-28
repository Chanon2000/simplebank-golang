package gapi

// file นี้สำหรับ CreateUser RPC โดยเฉพาะเลย // แนะนำให้แยกแต่ละ RPC ออกเป็นคนละ file

import (
	"context"

	db "github.com/chanon2000/simplebank/db/sqlc"
	"github.com/chanon2000/simplebank/pb"
	"github.com/chanon2000/simplebank/util"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// คือทำการ implement CreateUser ใน server ของเราเอง // เขียน implement ของ CreateUser จริงๆที่นี่แหละ
func (server *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	// จะเห็นว่าใน Gin นั้นเราต้องทำการ bind input parameter ลง request object ด้วยตัวเอง แต่ใน gRPC นั้นไม่ต้อง เพราะมันจัดการให้เราแล้วใน framework

	hashedPassword, err := util.HashPassword(req.GetPassword()) // ใช้ GetPassword() ดีกว่า เรียก req.Password ตรงๆเพราะว่ามันทำ safety check ก่อนด้วย
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %s", err)
	}

	arg := db.CreateUserParams{
		Username:       req.GetUsername(), // GetUsername GetFullName GetEmail มันเป็น function ที่ proto สร้างมาให้เรา เลยเอามาใช้ได้เลย และดีกว่าเรียก req.Username เพราะมันมีการ check value ใน function ให้ด้วย
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

	rsp := &pb.CreateUserResponse{
		User: convertUser(user),
	}
	return rsp, nil
}


