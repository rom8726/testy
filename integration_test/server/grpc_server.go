package server

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"integration_test/userspb"
)

// UserServiceServer implements UserService gRPC service
type UserServiceServer struct {
	userspb.UnimplementedUserServiceServer
	db *sql.DB
}

// AdminServiceServer implements AdminService gRPC service
type AdminServiceServer struct {
	userspb.UnimplementedAdminServiceServer
	db *sql.DB
}

// NewUserServiceServer creates a new UserServiceServer
func NewUserServiceServer(connStr string) (*UserServiceServer, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &UserServiceServer{db: db}, nil
}

// NewAdminServiceServer creates a new AdminServiceServer
func NewAdminServiceServer(connStr string) (*AdminServiceServer, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &AdminServiceServer{db: db}, nil
}

// CreateUser creates a new user
func (s *UserServiceServer) CreateUser(ctx context.Context, req *userspb.CreateUserRequest) (*userspb.CreateUserResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	userID := uuid.New().String()

	query := `INSERT INTO users (id, name, email, age, created_at, updated_at) VALUES ($1, $2, $3, $4, NOW(), NOW())`
	_, err := s.db.ExecContext(ctx, query, userID, req.Name, req.Email, req.Age)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create user: %v", err))
	}

	return &userspb.CreateUserResponse{
		Id:    userID,
		Name:  req.Name,
		Email: req.Email,
		Age:   req.Age,
	}, nil
}

// GetUser retrieves a user by ID
func (s *UserServiceServer) GetUser(ctx context.Context, req *userspb.GetUserRequest) (*userspb.GetUserResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	var id, name, email string
	var age sql.NullInt32

	query := `SELECT id, name, email, age FROM users WHERE id = $1`
	err := s.db.QueryRowContext(ctx, query, req.Id).Scan(&id, &name, &email, &age)
	if err == sql.ErrNoRows {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get user: %v", err))
	}

	var ageValue int32
	if age.Valid {
		ageValue = age.Int32
	}

	return &userspb.GetUserResponse{
		Id:    id,
		Name:  name,
		Email: email,
		Age:   ageValue,
	}, nil
}

// UpdateUser updates an existing user
func (s *UserServiceServer) UpdateUser(ctx context.Context, req *userspb.UpdateUserRequest) (*userspb.UpdateUserResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `UPDATE users SET name = $1, email = $2, updated_at = NOW() WHERE id = $3`
	result, err := s.db.ExecContext(ctx, query, req.Name, req.Email, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update user: %v", err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get rows affected: %v", err))
	}

	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &userspb.UpdateUserResponse{
		Id:    req.Id,
		Name:  req.Name,
		Email: req.Email,
	}, nil
}

// DeleteUser deletes a user by ID
func (s *UserServiceServer) DeleteUser(ctx context.Context, req *userspb.DeleteUserRequest) (*userspb.DeleteUserResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query := `DELETE FROM users WHERE id = $1`
	result, err := s.db.ExecContext(ctx, query, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete user: %v", err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get rows affected: %v", err))
	}

	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &userspb.DeleteUserResponse{
		Success: true,
	}, nil
}

// GetAllUsers retrieves all users (admin only)
func (s *AdminServiceServer) GetAllUsers(ctx context.Context, req *userspb.GetAllUsersRequest) (*userspb.GetAllUsersResponse, error) {
	query := `SELECT id, name, email, age FROM users ORDER BY created_at ASC`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get users: %v", err))
	}
	defer rows.Close()

	var users []*userspb.GetUserResponse
	for rows.Next() {
		var id, name, email string
		var age sql.NullInt32

		if err := rows.Scan(&id, &name, &email, &age); err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to scan user: %v", err))
		}

		var ageValue int32
		if age.Valid {
			ageValue = age.Int32
		}

		users = append(users, &userspb.GetUserResponse{
			Id:    id,
			Name:  name,
			Email: email,
			Age:   ageValue,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error iterating users: %v", err))
	}

	return &userspb.GetAllUsersResponse{
		Users: users,
	}, nil
}

// Close closes the database connection
func (s *UserServiceServer) Close() error {
	return s.db.Close()
}

// Close closes the database connection
func (s *AdminServiceServer) Close() error {
	return s.db.Close()
}
