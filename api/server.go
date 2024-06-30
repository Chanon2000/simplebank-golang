package api

import (
	"fmt"

	db "github.com/chanon2000/simplebank/db/sqlc"
	"github.com/chanon2000/simplebank/token"
	"github.com/chanon2000/simplebank/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10" // เราจะไม่ใช้ validator อันนี้กับ gRPC // และจริงๆแล้ว error message ที่ได้จาก package นี้นั้นมันก็ดูไม่ค่อย human-friendly และ computer-friendly เท่าไหร่
	// เช่น "error": "Key: 'createUserRequest.Username' Error:Field validation for 'Username' failed on the 'alphanum' tag" ซึ่งถ้าเราจะหาว่า fields ใหน error ใน message นี้ คุณต้องทำ analysis บน message ถึงจะรู้
)

// Server serves HTTP requests for our banking service.
type Server struct {
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
}

// NewServer creates a new HTTP server and set up routing.
func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	
	if err != nil {
		return nil, fmt.Errorf("connot create token maker: %w", err)
	}
	server := &Server{
		config: config,
		store: store,
		tokenMaker: tokenMaker,
	}
	

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	} 

	server.setupRouter()
	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()

	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)
	router.POST("/tokens/renew_access", server.renewAccessToken)

	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker))

	authRoutes.POST("/accounts", server.createAccount)
	authRoutes.GET("/accounts/:id", server.getAccount)
	authRoutes.GET("/accounts", server.listAccount)

	authRoutes.POST("/transfers", server.createTransfer)

	server.router = router
}

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}