package gapi
// เราจะสร้าง package หรือ folder แยกให้กับ API ที่ใช้ gRPC framework ไปเลย (เหมือนที่เราใช้ Gin framework แล้วเอา code ไว้ที่ api package หรือ folder) โดยจะตั้งชื่อ package นี้ว่า gapi นั้นเอง

import (
	"fmt"

	db "github.com/chanon2000/simplebank/db/sqlc"
	"github.com/chanon2000/simplebank/pb"
	"github.com/chanon2000/simplebank/token"
	"github.com/chanon2000/simplebank/util"
)

// Server serves gRPC requests for our banking service.
type Server struct {
	pb.UnimplementedSimpleBankServer // UnimplementedSimpleBankServer เป็น interface ที่ proto สร้างมาให้ ซึ่งมี RPC functions เตรียมมาให้ // โดย embed มันเข้า Server struct ของเราด้วยนั้นเอง เพื่อเอาความสามารถของมันมา เพื่อเช่นสามารถเรียก CreateUser, loginUser RPC
	config          util.Config
	store           db.Store
	tokenMaker      token.Maker
}

// NewServer creates a new gRPC server. // ซึ่งมันจะเป็น gRPC server ก็ต่อเมื่อเราเอา SimpleBankServer interface (คือ interface จากที่ proto มัน generate มาให้) มา implement นั้นเอง
func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		config:          config,
		store:           store,
		tokenMaker:      tokenMaker,
	}

	return server, nil
}
