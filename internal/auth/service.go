package auth

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"invitely-api/pkg/slug"
	"invitely-api/pkg/uuid"
)

type Service struct {
	supabase *SupabaseClient
	db       *sql.DB
}

func NewService(supabase *SupabaseClient, db *sql.DB) *Service {
	return &Service{supabase: supabase, db: db}
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (AuthResponse, error) {
	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		return AuthResponse{}, errors.New("email and password are required")
	}

	session, err := s.supabase.Login(ctx, req.Email, req.Password)
	if err != nil {
		return AuthResponse{}, err
	}

	user, err := s.ensureUser(ctx, session.User, "", "owner")
	if err != nil {
		return AuthResponse{}, err
	}

	return authResponse(session, user), nil
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (AuthResponse, error) {
	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		return AuthResponse{}, errors.New("email and password are required")
	}

	role := req.Role
	if role == "" {
		role = "owner"
	}

	session, err := s.supabase.Register(ctx, req.Email, req.Password)
	if err != nil {
		return AuthResponse{}, err
	}

	user, err := s.ensureUser(ctx, session.User, req.Name, role)
	if err != nil {
		return AuthResponse{}, err
	}

	return authResponse(session, user), nil
}

func (s *Service) EnsureUserFromToken(ctx context.Context, token string) (User, error) {
	supabaseUser, err := s.supabase.User(ctx, token)
	if err != nil {
		return User{}, err
	}

	return s.ensureUser(ctx, supabaseUser, "", "owner")
}

func (s *Service) ensureUser(ctx context.Context, supabaseUser SupabaseUser, name string, role string) (User, error) {
	var user User
	err := s.db.QueryRowContext(ctx, `
		select id, coalesce(tenant_id::text, ''), coalesce(supabase_user_id::text, ''), email, name, role
		from users
		where supabase_user_id = $1 or email = $2
	`, supabaseUser.ID, strings.ToLower(supabaseUser.Email)).Scan(
		&user.ID,
		&user.TenantID,
		&user.SupabaseUserID,
		&user.Email,
		&user.Name,
		&user.Role,
	)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return User{}, err
	}

	if name == "" {
		name = strings.Split(supabaseUser.Email, "@")[0]
	}

	tenantID := ""
	if role != "platform_admin" {
		var err error
		tenantID, err = uuid.New()
		if err != nil {
			return User{}, err
		}
		tenantName := name
		if tenantName == "" {
			tenantName = supabaseUser.Email
		}
		if _, err := s.db.ExecContext(ctx, `
			insert into tenants (id, name, slug)
			values ($1, $2, $3)
			on conflict (slug) do nothing
		`, tenantID, tenantName, slug.Make(tenantName+"-"+tenantID[:8])); err != nil {
			return User{}, err
		}
	}

	userID, err := uuid.New()
	if err != nil {
		return User{}, err
	}

	_, err = s.db.ExecContext(ctx, `
		insert into users (id, tenant_id, supabase_user_id, email, name, role)
		values ($1, nullif($2, '')::uuid, $3, $4, $5, $6)
	`, userID, tenantID, supabaseUser.ID, strings.ToLower(supabaseUser.Email), name, role)
	if err != nil {
		return User{}, err
	}

	return User{
		ID:             userID,
		TenantID:       tenantID,
		SupabaseUserID: supabaseUser.ID,
		Email:          strings.ToLower(supabaseUser.Email),
		Name:           name,
		Role:           role,
	}, nil
}

func authResponse(session SupabaseAuthResponse, user User) AuthResponse {
	return AuthResponse{
		AccessToken:  session.AccessToken,
		RefreshToken: session.RefreshToken,
		TokenType:    session.TokenType,
		ExpiresIn:    session.ExpiresIn,
		User:         user,
	}
}
