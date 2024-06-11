package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq" // เนื่องจากเราไม่ได้ใช้ package นี้ตรงๆใน file นี้ เลยเติม _ ไว้ข้างหน้า
)

const (
	dbDriver = "postgres"
	dbSource = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
)

var testQueries *Queries
var testDB *sql.DB // เอามาประกาศเป็น global แทนเพื่อให้ใช้ได้หลายที่และใน file อื่นได้

func TestMain(m *testing.M) {
	var err error

	testDB, err = sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("connot connect to db:", err)
	}

	testQueries = New(testDB)

	os.Exit(m.Run()) // m.Run() เพื่อรัน unit test โดยมันจะ return exit code (ซึ่งเป็นตัวบอกว่า test นั้น pass หรือ fail ) แล้วเราก็ report กลับไปที่ test runner ผ่าน os.Exit()
}