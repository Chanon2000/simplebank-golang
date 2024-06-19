package api

import (
	"fmt"

	db "github.com/chanon2000/simplebank/db/sqlc"
	"github.com/chanon2000/simplebank/token"
	"github.com/chanon2000/simplebank/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10" // ใช้ version 10
)

// คือ file ที่เราจะ implement HTTP API server

// Server serves HTTP requests for our banking service.
type Server struct {
	config     util.Config
	store      db.Store // เพื่อให้เราสามารถสื่อสารกับ db เมื่อทำการ processing API request
	tokenMaker token.Maker
	router     *gin.Engine // เพื่อทำให้เราสามารถส่งแต่ละ API request ไปที่ handler function
}

// NewServer creates a new HTTP server and set up routing.
func NewServer(config util.Config, store db.Store) (*Server, error) {
	// function นี้ จะสร้าง server instance ใหม่ และ setup ทุก http api routes ให้กับ service ของเราบน server
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey) // เราจะใช้ Paseto
	// tokenMaker, err := token.NewJWTMaker(config.TokenSymmetricKey) // ถ้าจะเปลี่ยนมาใช้ JWT ก็แค่เปลี่ยนมาใช้ NewJWTMaker แค่นั้นเลย
	
	if err != nil {
		return nil, fmt.Errorf("connot create token maker: %w", err)
	}
	server := &Server{
		config: config,
		store: store,
		tokenMaker: tokenMaker,
	} // สร้าง new server instance 
	

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok { // เพื่อเอา current validator engine ที่ gin ใช้อยู่ (ซึ่ง binding เป็น sub-package ของ gin)
		v.RegisterValidation("currency", validCurrency) // เอา validCurrency function ที่เราเขียนมา register ลง validator ชื่อ "currency"
	} 

	server.setupRouter()
	return server, nil
}

// เนื่องจากการ register route มันเริ่มยาวเลยย้ายออกมาวางที่ function อื่น
func (server *Server) setupRouter() {
	router := gin.Default() // สร้าง new router โดยการเรียก gin.Default()

	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)

	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccount)

	router.POST("/transfers", server.createTransfer)

	server.router = router
}

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string) error { // Start method ของ Server struct
	return server.router.Run(address) // เนื่องจาก Gin มี Run function มาให้อยู่แล้ว เลยทำแค่ return มัน
}

// implement errorResponse ที่ server.go เพราะจะเอาไว้ใช้ที่ handlers ที่อยู่ใน files อื่นๆด้วย
func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}