package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	dot "github.com/bartwild/WSO-scaling-and-ha-in-stateless-apps/internal"
	pb "github.com/bartwild/WSO-scaling-and-ha-in-stateless-apps/proto/dot_product"
)

type server struct {
	pb.UnimplementedDotProductServiceServer
}

var limiter = rate.NewLimiter(rate.Every(time.Second/1000), 1000)

func (s *server) Calculate(ctx context.Context, in *pb.DotProduct) (*pb.DotProductResult, error) {
	id := in.GetId()
	inputs := in.GetInput()

	results := make([]float32, len(inputs))
	for i, input := range inputs {
		res := dot.DotProductAVX2FMA(input.GetV1(), input.GetV2())
		results[i] = res
	}

	res := &pb.DotProductResult{}
	res.SetId(id)
	res.SetResult(results)
	return res, nil
}

func startManagementServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("READY"))
	})

	mux.HandleFunc("/panic", func(w http.ResponseWriter, r *http.Request) {
		log.Warn().Msg("Otrzymano żądanie /panic. Wyłączam proces!")
		os.Exit(1)
	})

	log.Info().Msg("Serwer zarządzający (HTTP) nasłuchuje na porcie :8081")
	if err := http.ListenAndServe(":8081", mux); err != nil {
		log.Fatal().Err(err).Msg("Błąd serwera HTTP")
	}
}

func unaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if !limiter.Allow() {
		return nil, status.Error(codes.ResourceExhausted, "zbyt dużo zapytań!")
	}
	return handler(ctx, req)
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	go startManagementServer()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}

	s := grpc.NewServer(grpc.MaxConcurrentStreams(128), grpc.NumStreamWorkers(16), grpc.UnaryInterceptor(unaryInterceptor))
	pb.RegisterDotProductServiceServer(s, &server{})
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(s, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	reflection.Register(s)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-quit
		log.Warn().Msg("Otrzymano sygnał zamknięcia, graceful shutdown...")
		healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		time.Sleep(5 * time.Second)
		s.GracefulStop()
	}()

	log.Info().Msg("Serwer gRPC nasłuchuje na porcie :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatal().Err(err).Msg("failed to serve")
	}
}
