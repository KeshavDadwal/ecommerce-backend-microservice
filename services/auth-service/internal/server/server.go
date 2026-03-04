package server

import (
	"context"
	"database/sql"

	authv1 "ecommerce-backend-microservice/proto/auth/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthServer implements authv1.AuthServiceServer.
type AuthServer struct {
	authv1.UnimplementedAuthServiceServer
	DB *sql.DB
}

// NewAuthServer returns a new AuthServer that uses the given DB.
func NewAuthServer(db *sql.DB) *AuthServer {
	return &AuthServer{DB: db}
}

// RegisterAdmin registers an admin user and returns JWT access and refresh tokens.
func (s *AuthServer) RegisterAdmin(ctx context.Context, req *authv1.RegisterRequest) (*authv1.AuthResponse, error) {
	if req.Email == "" || req.Password == "" || req.FullName == "" {
		return nil, status.Error(codes.InvalidArgument, "email, password, and full_name are required")
	}
	userID, roles, perms, err := s.createUserAndAssignRole(ctx, req.Email, req.Password, req.FullName)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	access, err := s.GenerateAccessToken(userID, roles, perms)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate access token")
	}
	refresh, err := s.GenerateRefreshToken(userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate refresh token")
	}
	return &authv1.AuthResponse{AccessToken: access, RefreshToken: refresh}, nil
}

// LoginAdmin authenticates an admin and returns JWT access and refresh tokens.
func (s *AuthServer) LoginAdmin(ctx context.Context, req *authv1.LoginRequest) (*authv1.AuthResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}
	userID, roles, perms, err := s.authenticateUser(ctx, req.Email, req.Password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	access, err := s.GenerateAccessToken(userID, roles, perms)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate access token")
	}
	refresh, err := s.GenerateRefreshToken(userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate refresh token")
	}
	return &authv1.AuthResponse{AccessToken: access, RefreshToken: refresh}, nil
}

// ValidateToken validates the access token and returns user_id, roles, and permissions from the JWT.
func (s *AuthServer) ValidateToken(ctx context.Context, req *authv1.ValidateRequest) (*authv1.ValidateResponse, error) {
	if req.AccessToken == "" {
		return nil, status.Error(codes.InvalidArgument, "access_token is required")
	}
	userID, roles, permissions, err := s.ParseAccessToken(req.AccessToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return &authv1.ValidateResponse{
		UserId:      userID,
		Roles:       roles,
		Permissions: permissions,
	}, nil
}

// createUserAndAssignRole inserts a user and assigns default ADMIN role; returns userID, roles, permissions.
func (s *AuthServer) createUserAndAssignRole(ctx context.Context, email, password, fullName string) (string, []string, []string, error) {
	hash, err := hashPassword(password)
	if err != nil {
		return "", nil, nil, err
	}
	userID, err := insertUser(ctx, s.DB, email, hash, fullName)
	if err != nil {
		return "", nil, nil, err
	}
	roles, perms, err := ensureAdminRoleAndAssign(ctx, s.DB, userID)
	if err != nil {
		return "", nil, nil, err
	}
	return userID, roles, perms, nil
}

// authenticateUser finds user by email, verifies password, returns userID and roles/permissions.
func (s *AuthServer) authenticateUser(ctx context.Context, email, password string) (string, []string, []string, error) {
	return findUserAndVerify(ctx, s.DB, email, password)
}
