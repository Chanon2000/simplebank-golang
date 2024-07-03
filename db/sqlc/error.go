package db

import (
	"github.com/jackc/pgx/v5"
)



var ErrRecordNotFound = pgx.ErrNoRows // กำหนดลงตัวแปรไปเลยเพื่อจะได้ใช้ได้หลายๆที่ // เวลาแก้ไขจะได้แก้ที่นี้ี่เดียว
