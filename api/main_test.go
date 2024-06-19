package api

import (
	"os"
	"testing"
	"time"

	db "github.com/chanon2000/simplebank/db/sqlc"
	"github.com/chanon2000/simplebank/util"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// สร้าง server สำหรับ test
func newTestServer(t *testing.T, store db.Store) *Server {
	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := NewServer(config, store)
	require.NoError(t, err)

	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode) // ครั้งนี้สร้าง main_test.go มาเพื่อจะกำหนด SetMode ในกับ Gin เป็น TestMode แค่นั้นแหละ เพื่อให้ไม่ต้องแสดงข้อฒุลที่ terminal เยอะ
	os.Exit(m.Run())
}