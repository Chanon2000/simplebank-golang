package main

import (
	"context"
	"database/sql"
	"os"

	// "log"
	"net"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log" // เปลี่ยนมาใช้ zerolog แทน // ใช้ zerolog ในการเขียน logs ให้หมด

	"github.com/chanon2000/simplebank/api"
	db "github.com/chanon2000/simplebank/db/sqlc"
	_ "github.com/chanon2000/simplebank/doc/statik" // ตรงนี้แหละคือทำการ point ไปที่ statik.go นั้นเอง
	"github.com/chanon2000/simplebank/gapi"
	"github.com/chanon2000/simplebank/pb"
	"github.com/chanon2000/simplebank/util"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // doc ของ migrate บอกให้ import // เพียงเพื่อ point ไปที่ database/postgres subpackage ของ migrate module
	_ "github.com/golang-migrate/migrate/v4/source/file"       // doc ของ migrate บอกให้ import // เนื่องจากเราใช้ file schema เลยใส่ /file
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"github.com/rakyll/statik/fs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)


func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server")
	}

	if config.Environment == "development" { // ย้ายมาเขียนล่าง LoadConfig เพราะเราจะ check config.Environment ว่าอยู่ env ใหน แล้วค่อยสั่ง log.Output
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}) // ใส่ line นี้ไว้ที่บนสุด เพื่อให้ logs ที่แสดงที่ server นั้นอ่านง่ายขึ้น (ศึกษาเพิ่มเติมได้ที่ doc ของ zerolog)
	}
	
	println("config.DBDriver", config.DBDriver)
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}

	// เราจะแก้ด้วยการเปลี่ยนมารัน migration ที่ main function ตรงนี้แทน  // รัน migration ใน golang สามารถดูได้ใน doc
	// run db migration
	runDBMigration(config.MigrationURL, config.DBSource)


	store := db.NewStore(conn)
	go runGatewayServer(config, store) // เราจะต้องทำการ serve ทั้ง gRPC และ HTTP requests ในเวลาเดียวกัน แต่เราไม่สามารถเรียกทั้ง 2 function ใน go routine เดียวกันได้ เพราะ first server มันจะ block อีก server นึง
	// ซึ่งเราก็แค่ใส่ go keyword เพื่อให้ runGatewayServer รันในอีก go routine นั้นเอง
	runGrpcServer(config, store)
	// ซึ่งก็จะทำให้ทั้ง gRPC server และ HTTP server นั้น start ขึ้นมาพร้อมกันนั้นเอง
}

func runDBMigration(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create new migrate instance")
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange { // err != migrate.ErrNoChange คือถ้า error คือ no change ก็จะถือให้ success ไปเลย ไม่ต้องเข้า error ตรงนี้
		log.Fatal().Err(err).Msg("failed to run migrate up")
	}

	log.Info().Msg("db migrated successfully")
}

// เอาไว้รัน gRPC server
func runGrpcServer(config util.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server")
	}

	gprcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger) // เนื่องจาก gRPC ของเราเป็นประเภท unary
	grpcServer := grpc.NewServer(gprcLogger) // ใส่ gprcLogger ลง NewServer คือการเพิ่ม interceptor เข้า gRPC server
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GRPCServerAddress) // "tcp" คือ protocol
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create listener")
	}

	log.Info().Msgf("start gRPC server at %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start gRPC server")
	}
}

func runGatewayServer(config util.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server")
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
		log.Fatal().Err(err).Msg("cannot register handler server")
	}

	mux := http.NewServeMux() // mux จะทำการรับ http requests จาก clients
	mux.Handle("/", grpcMux) // ทำการ reroute ไปที่ gRPC mux และ convert มันเป็น gRPC format

	statikFS, err := fs.New()
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create statik fs")
	}

	// fs := http.FileServer(http.Dir("./doc/swagger")) // เพื่อ serve static files หรือ api doc ของเรา
	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(statikFS))
	mux.Handle("/swagger/", swaggerHandler)

	listener, err := net.Listen("tcp", config.HTTPServerAddress)
	if err != nil {
		log.Error().Err(err).Msg("cannot create listener:")
	}

	log.Printf("start HTTP gateway server at %s", listener.Addr().String())
	err = http.Serve(listener, mux)
	if err != nil {
		log.Error().Err(err).Msg("cannot start HTTP gateway server:")
	}
}

// เอาไว้รัน Gin server
func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Error().Err(err).Msg("cannot create server:")
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Error().Err(err).Msg("cannot start server:")
	}
}