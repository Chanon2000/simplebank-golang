package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq" // import ที่ตรงนี้ด้วยเพื่อให้สามารถใช้งานเบื้องหลังใน file นี้ได้
	"github.com/chanon2000/simplebank/api"
	db "github.com/chanon2000/simplebank/db/sqlc"
	"github.com/chanon2000/simplebank/util"
)

// main.go คือเป็น entry point ของ server

func main() {
	config, err := util.LoadConfig(".") // "." เนื่องจาก main.go มันอยู่ location เดียวกับ app.env
	if err != nil {
		log.Fatal("connot load config:", err)
	}
	
	println("config.DBDriver", config.DBDriver)
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("connot connect to db:", err)
	}

	store := db.NewStore(conn)
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	err = server.Start(config.ServerAddress) // คือสั่ง start server ตรงนี้
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}

