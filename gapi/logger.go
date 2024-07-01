package gapi // เราจะ implement logger interceptor ใน file นี้

import (
	"context"
	"time"

	"github.com/rs/zerolog/log" // เราจะใช้ zerolog เพื่อเขียน log ใน json format
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GrpcLogger( // definition ของ function นี้ นั้นเอามาจาก UnaryServerInterceptor interface ของ UnaryInterceptor (โดย commad + คลิก เข้าไปที่ function นั้น) แล้วก็ตั้งชื่อเป็น GrpcLogger // ซึ่งก็เพราะเราเราจะใส่ function นี้ลงใน UnaryInterceptor function อีกทีนั้นเอง
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	startTime := time.Now()
	result, err := handler(ctx, req) // คือ forward request ไปที่ handler function เพื่อ processed ต่อ แล้วก็จะได้ result ออกมา
	duration := time.Since(startTime)

	statusCode := codes.Unknown
	if st, ok := status.FromError(err); ok {
		statusCode = st.Code()
	}

	logger := log.Info() // เก็บ log.Info() ลง logger
	if err != nil { // ถ้ามี err ก็จะเป็น log.Error().Err(err) ใน logger แทน
		logger = log.Error().Err(err)
	}
	
	// log.Print("received a gRPC request") // zerolog นั้นเขียนแค่ Print ก็จะ log เป็น json เลย -> {"level":"debug","time":"2024-07-01T10:30:04+07:00","message":"received a gRPC request"}

	// log.Info().Str("protocol", "grpc"). // Info().Msg() เพื่อเขียน log ใน info-level log แทน debug level
	// 	Str("method", info.FullMethod).
	// 	Int("status_code", int(statusCode)).
	// 	Str("status_text", statusCode.String()).
	// 	Dur("duration", duration).
	// 	Msg("received a gRPC request")

	logger.Str("protocol", "grpc").
		Str("method", info.FullMethod).
		Int("status_code", int(statusCode)).
		Str("status_text", statusCode.String()).
		Dur("duration", duration).
		Msg("received a gRPC request")

	return result, err
}
