package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq" // import ที่ตรงนี้ด้วยเพื่อให้สามารถใช้งานเบื้องหลังใน file นี้ได้
	"github.com/chanon2000/simplebank/api"
	db "github.com/chanon2000/simplebank/db/sqlc"
)

// main.go คือเป็น entry point ของ server

const (
	dbDriver = "postgres"
	dbSource = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
	serverAddress = "0.0.0.0:8080"
)

func main() {
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("connot connect to db:", err)
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)

	err = server.Start(serverAddress) // คือสั่ง start server ตรงนี้
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}

