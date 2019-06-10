package cmd

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"github.com/SirawichDev/grpc-crud/pkg/protocol/grpc"
	v1 "github.com/SirawichDev/grpc-crud/pkg/services/v1"
)

type Config struct {
	GRPCPort            string
	DatastoreDBHost     string
	DatastoreDBUser     string
	DatastoreDBPassword string
	DatastoreDBSchema   string
}

func Runserver() error {
	ctx := context.Background()
	var cfg Config

	flag.StringVar(&cfg.GRPCPort, "grpc-port", "", "gRPC port to bind")
	flag.StringVar(&cfg.DatastoreDBHost, "db-host", "", "/cloudsql/crud-grpc-microservice-golang:us-central1:todo")
	flag.StringVar(&cfg.DatastoreDBUser, "db-user", "", "5635512124@psu.ac.th")
	flag.StringVar(&cfg.DatastoreDBPassword, "db-password", "", "1w2e3r4t5y")
	flag.StringVar(&cfg.DatastoreDBSchema, "db-schema", "", "todos")

	flag.Parse()

	if len(cfg.GRPCPort) == 0 {
		return fmt.Errorf("invalid TCP for gRPC server: '%s'", cfg.GRPCPort)

	}

	param := "parseTime=true"

	dsn := fmt.Sprintf("%s:%s@unix(%s)?%s",
		cfg.DatastoreDBUser,
		cfg.DatastoreDBPassword,
		cfg.DatastoreDBHost,
		cfg.DatastoreDBSchema, param)
	db,err := sql.Open("mysql",dsn)
	if err != nil{
		return fmt.Errorf("failed to open database: %v",err)

	}
	defer  db.Close()

	v1API := v1.NewTaskServiceServer(db)
	return grpc.Runserver(ctx,v1API,cfg.GRPCPort)
}
