package grpc

import (
	"context"
	v1 "github.com/SirawichDev/grpc-crud/pkg/api/v1"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
)

func Runserver(ctx context.Context, v1API v1.TodoServiceServer, port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	server := grpc.NewServer()
	v1.RegisterTodoServiceServer(server, v1API)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			log.Println("shutting down Grpc server..")

			server.GracefulStop()
			<-ctx.Done()
		}
	}()

	log.Println("strarting grpc server...")
	return server.Serve(lis)
}
