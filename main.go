package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"

	"github.com/chanon2000/simplebank/api"
	db "github.com/chanon2000/simplebank/db/sqlc"
	"github.com/chanon2000/simplebank/gapi"
	"github.com/chanon2000/simplebank/pb"
	"github.com/chanon2000/simplebank/util"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)


func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("connot load config:", err)
	}
	
	println("config.DBDriver", config.DBDriver)
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("connot connect to db:", err)
	}

	store := db.NewStore(conn)
	go runGatewayServer(config, store) // เราจะต้องทำการ serve ทั้ง gRPC และ HTTP requests ในเวลาเดียวกัน แต่เราไม่สามารถเรียกทั้ง 2 function ใน go routine เดียวกันได้ เพราะ first server มันจะ block อีก server นึง
	// ซึ่งเราก็แค่ใส่ go keyword เพื่อให้ runGatewayServer รันในอีก go routine นั้นเอง
	runGrpcServer(config, store)
	// ซึ่งก็จะทำให้ทั้ง gRPC server และ HTTP server นั้น start ขึ้นมาพร้อมกันนั้นเอง
}

// เอาไว้รัน gRPC server
func runGrpcServer(config util.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GRPCServerAddress) // "tcp" คือ protocol
	if err != nil {
		log.Fatal("cannot create listener:", err)
	}

	log.Printf("start gRPC server at %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot start gRPC server:", err)
	}
}

func runGatewayServer(config util.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
		// เพื่อ enable snake case ให้กับ gRPC gateway server
	})

	grpcMux := runtime.NewServeMux(jsonOption) // NewServeMux มาจาก runtime package ซึ่งคือ sub-package ของ grpc-gateway v2

	ctx, cancel := context.WithCancel(context.Background()) // สร้าง context
	defer cancel() // cancel context เพื่อป้องกันไม่ให้ system ทำงานที่ไม่จำเป็น

	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server) // RegisterSimpleBankHandlerServer คือ func ที่ protoc generate มาให้
	if err != nil {
		log.Fatal("cannot register handler server:", err)
	}

	mux := http.NewServeMux() // mux จะทำการรับ http requests จาก clients
	mux.Handle("/", grpcMux) // ทำการ reroute ไปที่ gRPC mux และ convert มันเป็น gRPC format

	fs := http.FileServer(http.Dir("./doc/swagger")) // เพื่อ serve static files หรือ api doc ของเรา
	mux.Handle("/swagger/", http.StripPrefix("/swagger/", fs))

	listener, err := net.Listen("tcp", config.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot create listener:", err)
	}

	log.Printf("start HTTP gateway server at %s", listener.Addr().String())
	err = http.Serve(listener, mux)
	if err != nil {
		log.Fatal("cannot start HTTP gateway server:", err)
	}
}

// เอาไว้รัน Gin server
func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}