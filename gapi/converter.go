package gapi

import (
	db "github.com/chanon2000/simplebank/db/sqlc"
	"github.com/chanon2000/simplebank/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func convertUser(user db.User) *pb.User {
	return &pb.User{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: timestamppb.New(user.PasswordChangedAt), // ใช้ timestamppb เพื่อ convert timestamp type เนื่องจากใน protobuf นั้นมี timestamp type ที่ไม่เหมือน Golang's Time type 
		CreatedAt:         timestamppb.New(user.CreatedAt),
	}
}
