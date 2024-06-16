package api

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode) // ครั้งนี้สร้าง main_test.go มาเพื่อจะกำหนด SetMode ในกับ Gin เป็น TestMode แค่นั้นแหละ เพื่อให้ไม่ต้องแสดงข้อฒุลที่ terminal เยอะ
	os.Exit(m.Run())
}