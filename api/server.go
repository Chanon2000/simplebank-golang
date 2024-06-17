package api

import (
	db "github.com/chanon2000/simplebank/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10" // ใช้ version 10
)

// คือ file ที่เราจะ implement HTTP API server

// Server serves HTTP requests for our banking service.
type Server struct {
	store      db.Store // เพื่อให้เราสามารถสื่อสารกับ db เมื่อทำการ processing API request
	router     *gin.Engine // เพื่อทำให้เราสามารถส่งแต่ละ API request ไปที่ handler function
}

// NewServer creates a new HTTP server and set up routing.
func NewServer(store db.Store) *Server {
	// function นี้ จะสร้าง server instance ใหม่ และ setup ทุก http api routes ให้กับ service ของเราบน server
	server := &Server{store: store} // สร้าง new server instance 
	router := gin.Default() // สร้าง new router โดยการเรียก gin.Default()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok { // เพื่อเอา current validator engine ที่ gin ใช้อยู่ (ซึ่ง binding เป็น sub-package ของ gin)
		v.RegisterValidation("currency", validCurrency) // เอา validCurrency function ที่เราเขียนมา register ลง validator ชื่อ "currency"
	} 

	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccount)

	router.POST("/transfers", server.createTransfer)

	server.router = router
	return server
}

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string) error { // Start method ของ Server struct
	return server.router.Run(address) // เนื่องจาก Gin มี Run function มาให้อยู่แล้ว เลยทำแค่ return มัน
}

// implement errorResponse ที่ server.go เพราะจะเอาไว้ใช้ที่ handlers ที่อยู่ใน files อื่นๆด้วย
func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}