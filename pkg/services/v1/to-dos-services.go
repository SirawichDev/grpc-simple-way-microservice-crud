package v1

import (
	"context"
	"database/sql"
	"fmt"
	v1 "github.com/SirawichDev/grpc-crud/pkg/api/v1"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
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

func (s *toDoServiceServer) Update(ctx context.Context, req *v1.MakeUpdateRequest) (*v1.MakeUpdateResponse, error) {
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
		return nil, status.Error(codes.InvalidArgument, "timestamp field has invalid"+err.Error())
	}

	res, err := c.ExecContext(ctx, "UPDATE Todo SET `Title`=?,`Description`=?,`Timestamp`=? WHERE `ID` =? ", req.Todo.Title, req.Todo.Description,
		timestamp, req.Todo.Id)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to update ToDo-> "+err.Error())
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to update "+err.Error())

	}
	if rows == 0 {
		return nil, status.Error(codes.NotFound, fmt.Sprint("Todo with Id = '%d' ", req.Todo.Id))
	}

	return &v1.MakeUpdateResponse{
		Api:     apiVer,
		Updated: rows,
	}, nil
}

func (s *toDoServiceServer) GetOne(ctx context.Context, req *v1.MakeGetRequest) (*v1.MakeGetResponse, error) {
	if err := s.checkHealth(req.Api); err != nil {
		return nil, err
	}
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	rows, err := c.QueryContext(ctx, "SELECT `ID`, `Title`,`Description`, `Timestamp` FROM Todo WHERE  `ID` = ?", req.Id)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to selete from todo -> "+err.Error())

	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, status.Error(codes.Unknown, "failed to get data from todo =>"+err.Error())
		}
		return nil, status.Error(codes.NotFound, fmt.Sprintf("TODO with ID = '%d' is no found", req.Id))
	}
	var td v1.Todo
	var timeStamp time.Time
	if err := rows.Scan(&td.Id, &td.Title, &td.Description, &timeStamp); err != nil {
		return nil, status.Error(codes.Unknown, "failed to get field values from todo row ->"+err.Error())
	}
	td.Timestamp, err = ptypes.TimestampProto(timeStamp)
	if err != nil {
		return nil, status.Error(codes.Unknown, "timestamp field has invalid format->"+err.Error())
	}
	if rows.Next() {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("found multiple Todo rows with ID='%d'", req.Id))
	}

	return &v1.MakeGetResponse{
		Api:  apiVer,
		Todo: &td,
	}, nil
}
func (s *toDoServiceServer) Delete(ctx context.Context, req *v1.MakeDeleteRequest) (*v1.MakeDeleteResponse, error) {
	if err := s.checkHealth(req.Api); err != nil {
		return nil, err
	}
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	res, err := c.ExecContext(ctx, "DELETE FROM Todo WHERE `ID`=?", req.Id)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to delete Todo=>"+err.Error())
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to get row affected value "+err.Error())

	}
	if rows == 0 {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Todo with ID='%d' is not found", req.Id))
	}
	return &v1.MakeDeleteResponse{
		Api:     apiVer,
		Deleted: rows,
	}, nil
}

func (s *toDoServiceServer) GetAll(ctx context.Context, req *v1.MakeCreateRequest) (*v1.MakeGetAllResponse, error) {
	if err := s.checkHealth(req.Api); err != nil {
		return nil, err
	}
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	rows, err := c.QueryContext(ctx, "SELECT `ID`,`Title`,`Description`,`Timestamp` FROM Todo")
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to select from todo ->"+err.Error())
	}
	defer rows.Close()
	var timestamp time.Time
	list := []*v1.Todo{}
	for rows.Next() {
		td := new(v1.Todo)
		if err := rows.Scan(&td.Id, &td.Title, &td.Description, &timestamp); err != nil {
			return nil, status.Error(codes.Unknown, "failed to get field values from todo row->"+err.Error())
		}
		td.Timestamp, err = ptypes.TimestampProto(timestamp)
		if err != nil {
			return nil, status.Error(codes.Unknown, "timestamp field has invalid format"+err.Error())
		}
		list = append(list, td)
	}
	if err := rows.Err(); err != nil {
		return nil, status.Error(codes.Unknown, "failed to get data from todo")
	}
	return &v1.MakeGetAllResponse{
		Api:  apiVer,
		Todo: list,
	}, nil
}
