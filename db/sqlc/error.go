package db

import (
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/jackc/pgx/v5/pgconn"
)

const ( // ดู code number จาก https://www.postgresql.org/docs/current/errcodes-appendix.html ได้
	ForeignKeyViolation = "23503" // foreign_key_violation
	UniqueViolation     = "23505" // unique_violation
)

var ErrRecordNotFound = pgx.ErrNoRows // กำหนดลงตัวแปรไปเลยเพื่อจะได้ใช้ได้หลายๆที่ // เวลาแก้ไขจะได้แก้ที่นี้ี่เดียว

var ErrUniqueViolation = &pgconn.PgError{
	Code: UniqueViolation,
}

func ErrorCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) { // คือ แปลง err เป็น pgErr ด้วย errors.As นั้นเอง ถ้า convert นั้น success มันจะ return true นั้นเอง
		return pgErr.Code
	}
	return ""
}
