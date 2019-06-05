package v1

import (
	"context"
	"database/sql"
	v1 "github.com/SirawichDev/grpc-crud/pkg/api/v1"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	apiVer = "v1"
)

type toDoServiceServer struct {
	db *sql.DB
}

func NewTaskServiceServer(db *sql.DB) v1.TodoServiceServer {
	return &toDoServiceServer{db}
}

func (s *toDoServiceServer) checkHealth(api string) error {
	if len(api) > 0 {
		if apiVer != api {
			return status.Errorf(codes.Unimplemented, "unsupport api version")
		}
	}
	return nil
}
func (s *toDoServiceServer) connect(ctx context.Context) (*sql.Conn, error) {
	e, err := s.db.Conn(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "Failed to Connect database")
	}
	return e, nil
}
func (s *toDoServiceServer) Create(ctx context.Context, req *v1.MakeCreateRequest) (*v1.MakeCreateResponse, error) {
	if err := s.checkHealth(req.Api); err != nil {
		return nil, err
	}
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	timestamp, err := ptypes.Timestamp(req.Todo.Timestamp)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid timestamp"+err.Error())

	}
	res, err := c.ExecContext(ctx, "INSERT INTO ToDo(`Title`,`Description`,`TimeStamp`) VALUE(?,?,?) ", req.Todo.Title, req.Todo.Description, timestamp)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to insert to todo"+err.Error())
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to retrieve id for created todo"+err.Error())
	}
	return &v1.MakeCreateResponse{
		Api: apiVer,
		Id:  id,
	}, nil

}
