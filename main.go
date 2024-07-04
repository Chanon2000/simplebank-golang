package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"net"
	"net/http"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"github.com/chanon2000/simplebank/api"
	db "github.com/chanon2000/simplebank/db/sqlc"
	_ "github.com/chanon2000/simplebank/doc/statik"

	"github.com/chanon2000/simplebank/gapi"
	"github.com/chanon2000/simplebank/mail"
	"github.com/chanon2000/simplebank/pb"
	"github.com/chanon2000/simplebank/util"
	"github.com/chanon2000/simplebank/worker"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/rakyll/statik/fs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

// เพื่อที่จะรู้ว่าควร shotdown เมื่อไหร่ เราต้อง listen บาง interrupt signals ก่อน โดยตรงนี้เราจะทำการประกาศ list of signals
var interruptSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
	syscall.SIGINT,
}

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server")
	}

	if config.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	ctx, stop := signal.NotifyContext(context.Background(), interruptSignals...) // NotifyContext เพื่อทำการ register list ของ signal ที่เราต้องการจะ listen
	defer stop()
	
	connPool, err := pgxpool.New(ctx, config.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}

	runDBMigration(config.MigrationURL, config.DBSource)

	store := db.NewStore(connPool)

	redisOpt := asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}

	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)
	
	waitGroup, ctx := errgroup.WithContext(ctx) // WithContext จาก errgroup จะ return waitGroup และ context // ซึ่งคุณจะเอาไปใช้ในแต่ละ run woker and server function ข้างล่างต่อ
	// เพื่อที่เราจะเอาไว้รันแต่ละ server และ worker ในคนละ go runtine ข้างในแต่ละ function อีกที

	// โดยตั้ง 3 function นี้ต้องรันในคนละ go routine และรอทั้ง 3 ก่อนที่จะ shutdown เมื่อได้รับ interrupt signal
	runTaskProcessor(ctx, waitGroup, config, redisOpt, store)
	runGatewayServer(ctx, waitGroup, config, store, taskDistributor)
	runGrpcServer(ctx, waitGroup, config, store, taskDistributor)

	err = waitGroup.Wait() // เพื่อรอทุก go routine ให้ compiete ก่อนที่จะ exiting main function
	if err != nil {
		log.Fatal().Err(err).Msg("error from wait group")
	}
	
}

func runDBMigration(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create new migrate instance")
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Err(err).Msg("failed to run migrate up")
	}

	log.Info().Msg("db migrated successfully")
}

// เพื่อรัน task processor
func runTaskProcessor(ctx context.Context, waitGroup *errgroup.Group, config util.Config, redisOpt asynq.RedisClientOpt, store db.Store) {
	mailer := mail.NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)
	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, store, mailer)

	log.Info().Msg("start task processor")
	err := taskProcessor.Start() // เนื่องจาก Start() มันจะ start processor ใน go routine แยกอยู่แล้ว เราเลยไม่จำเป็นต้อง create go routine เพิ่มตรงนี้ นั้นเอง
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start task processor")
	}

	// สร้าง go runtine เพื่อ listen interrupts และ gracefully shutdown task processor
	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("graceful shutdown task processor")

		taskProcessor.Shutdown()
		log.Info().Msg("task processor is stopped")

		return nil
	})
}

func runGrpcServer(ctx context.Context, waitGroup *errgroup.Group, config util.Config, store db.Store, taskDistributor worker.TaskDistributor) {
	server, err := gapi.NewServer(config, store, taskDistributor)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server")
	}

	gprcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)
	grpcServer := grpc.NewServer(gprcLogger)
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create listener")
	}

	// start server ใน go runtine
	waitGroup.Go(func() error { // waitGroup.Go เพื่อ start new go routine // โดยใส่ function เป็น input ซึ่งจะเป็น function ที่รันอีก go runtine นี้นั้นเอง
		log.Info().Msgf("start gRPC server at %s", listener.Addr().String())

		err = grpcServer.Serve(listener) // แล้วก็เอา code ตรงนี้ไว้ใน func นี้เพื่อทำการ start gRPC server ให้ go runtine แยกนั้นเอง
		if err != nil {
			if errors.Is(err, grpc.ErrServerStopped) { // เพื่อไม่ให้มันแสดง error เนื่อจาก server stop (เนื่องจากโดย default มันจะแสดง error นั้นเอง)// เนื่องจากเราไม่ต้องการให้มันแสดง error เนื่องจากทำการ interrupt ที่ทำให้ server ถูก stop เช่นตอนกด control + c ที่ terminal
				return nil
			}
			log.Error().Err(err).Msg("gRPC server failed to serve")
			return err
		}

		return nil
	})

	// ส่วนในการ listen interrupt signal และ gracefully shutdown the server // ซึ่งทำให้อีก go runtine นี้โดยการรัน waitGroup.GO
	waitGroup.Go(func() error {
		<-ctx.Done() // อ่าน value จาก context.Done() channel ซึ่ง call นี้จะถูก block จนกว่า context จะ done (ซึ่งหมายถึง interrupt signal นั้นเอง)
		log.Info().Msg("graceful shutdown gRPC server")

		grpcServer.GracefulStop() // รัน GracefulStop เพื่อ stop gRPC server
		log.Info().Msg("gRPC server is stopped")

		return nil
	})
}

func runGatewayServer(ctx context.Context, waitGroup *errgroup.Group, config util.Config, store db.Store, taskDistributor worker.TaskDistributor) {
	server, err := gapi.NewServer(config, store, taskDistributor)
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
	})

	grpcMux := runtime.NewServeMux(jsonOption)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot register handler server")
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	statikFS, err := fs.New()
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create statik fs")
	}

	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(statikFS))
	mux.Handle("/swagger/", swaggerHandler)

	httpServer := &http.Server{ // ประกาศ httpServer ใหม่
		Handler: gapi.HttpLogger(mux), // handler function
		Addr:    config.HTTPServerAddress, // address ของ server
	}

	// start server ใน go runtine
	waitGroup.Go(func() error {
		log.Info().Msgf("start HTTP gateway server at %s", httpServer.Addr)
		err = httpServer.ListenAndServe() // start server โดยใช้ ListenAndServe function
		if err != nil {
			if errors.Is(err, http.ErrServerClosed) { // เพื่อไม่ให้มันแสดง error เมื่อ http server ถูก stop
				return nil
			}
			log.Error().Err(err).Msg("HTTP gateway server failed to serve")
			return err
		}
		return nil
	})

	// ส่วนในการ listen interrupt signal และ gracefully shutdown the server
	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("graceful shutdown HTTP gateway server")

		err := httpServer.Shutdown(context.Background()) // Shutdown จะทำการ stop listening new request และจะ wait จนกว่า ทุก active request จะ completed แล้วค่อย shutdown
		// คุณสามารถใส่ timeout เพื่อบอก maximun time ที่จะ wait ได้ // แต่ครั้งนี้คุณไม่ใส่นั้นหมายความว่า no time limit สำหรับ waiting
		if err != nil {
			log.Error().Err(err).Msg("failed to shutdown HTTP gateway server")
			return err
		}

		log.Info().Msg("HTTP gateway server is stopped")
		return nil
	})

}

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