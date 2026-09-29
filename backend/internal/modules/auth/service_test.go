package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"football-api/internal/modules/admin"
	"football-api/internal/modules/auth"

	"golang.org/x/crypto/bcrypt"
)

type fakeAdminRepo struct {
	admins  map[string]*admin.Admin
	byID    map[string]*admin.Admin
	findErr error
}

func newFakeAdminRepo(admins ...*admin.Admin) *fakeAdminRepo {
	f := &fakeAdminRepo{
		admins: make(map[string]*admin.Admin),
		byID:   make(map[string]*admin.Admin),
	}
	for _, a := range admins {
		f.admins[a.Email] = a
		f.byID[a.ID] = a
	}
	return f
}

func (f *fakeAdminRepo) FindByEmail(_ context.Context, email string) (*admin.Admin, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	a, ok := f.admins[email]
	if !ok {
		return nil, admin.ErrNotFound
	}
	return a, nil
}

func (f *fakeAdminRepo) FindByID(_ context.Context, id string) (*admin.Admin, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	a, ok := f.byID[id]
	if !ok {
		return nil, admin.ErrNotFound
	}
	return a, nil
}

func (f *fakeAdminRepo) Create(_ context.Context, _, _, _ string) (*admin.Admin, error) {
	return nil, nil
}

func (f *fakeAdminRepo) ExistsAny(_ context.Context) (bool, error) { return false, nil }

type fakeAuthRepo struct {
	revoked   map[string]bool
	insertErr error
}

func newFakeAuthRepo() *fakeAuthRepo {
	return &fakeAuthRepo{revoked: make(map[string]bool)}
}

func (f *fakeAuthRepo) InsertRevokedToken(_ context.Context, jti string, _ time.Time) error {
	if f.insertErr != nil {
		return f.insertErr
	}
	f.revoked[jti] = true
	return nil
}

func (f *fakeAuthRepo) IsRevoked(_ context.Context, jti string) (bool, error) {
	return f.revoked[jti], nil
}

func (f *fakeAuthRepo) DeleteExpiredTokens(_ context.Context) error { return nil }

func makeHash(pw string) string {
	h, _ := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.MinCost)
	return string(h)
}

const testSecret = "test-secret-key"

func TestLogin(t *testing.T) {
	validAdmin := &admin.Admin{
		ID:           "admin-1",
		Email:        "admin@test.com",
		Name:         "Admin",
		PasswordHash: makeHash("correct-password"),
	}

	tests := []struct {
		name      string
		email     string
		password  string
		wantErr   error
		wantToken bool
	}{
		{
			name:      "valid credentials returns token",
			email:     "admin@test.com",
			password:  "correct-password",
			wantToken: true,
		},
		{
			name:     "wrong password returns ErrInvalidCredentials",
			email:    "admin@test.com",
			password: "wrong-password",
			wantErr:  auth.ErrInvalidCredentials,
		},
		{
			name:     "unknown email returns ErrInvalidCredentials",
			email:    "noone@test.com",
			password: "whatever",
			wantErr:  auth.ErrInvalidCredentials,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := auth.NewService(newFakeAdminRepo(validAdmin), newFakeAuthRepo(), testSecret)

			resp, err := svc.Login(context.Background(), tc.email, tc.password)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("Login() error = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Login() unexpected error: %v", err)
			}
			if resp.Token == "" {
				t.Error("Login() returned empty token")
			}
			if resp.ExpiresAt.Before(time.Now()) {
				t.Error("Login() ExpiresAt is in the past")
			}
		})
	}
}

func TestLogout(t *testing.T) {
	validAdmin := &admin.Admin{
		ID:           "admin-1",
		Email:        "admin@test.com",
		Name:         "Admin",
		PasswordHash: makeHash("password"),
	}

	t.Run("logout inserts jti into blacklist", func(t *testing.T) {
		authRepo := newFakeAuthRepo()
		svc := auth.NewService(newFakeAdminRepo(validAdmin), authRepo, testSecret)

		loginResp, err := svc.Login(context.Background(), "admin@test.com", "password")
		if err != nil {
			t.Fatalf("Login() error: %v", err)
		}

		claims := auth.MustParseClaims(t, loginResp.Token, testSecret)

		if err := svc.Logout(context.Background(), claims); err != nil {
			t.Fatalf("Logout() error: %v", err)
		}

		revoked, _ := authRepo.IsRevoked(context.Background(), claims.ID)
		if !revoked {
			t.Error("jti should be in the blacklist after logout")
		}
	})

	t.Run("logout with invalid claims returns ErrTokenInvalid", func(t *testing.T) {
		svc := auth.NewService(newFakeAdminRepo(validAdmin), newFakeAuthRepo(), testSecret)
		err := svc.Logout(context.Background(), &auth.Claims{})
		if !errors.Is(err, auth.ErrTokenInvalid) {
			t.Fatalf("Logout() error = %v, want ErrTokenInvalid", err)
		}
	})

	t.Run("repo error is propagated", func(t *testing.T) {
		authRepo := newFakeAuthRepo()
		authRepo.insertErr = errors.New("db error")
		svc := auth.NewService(newFakeAdminRepo(validAdmin), authRepo, testSecret)

		loginResp, _ := svc.Login(context.Background(), "admin@test.com", "password")
		claims := auth.MustParseClaims(t, loginResp.Token, testSecret)

		if err := svc.Logout(context.Background(), claims); err == nil {
			t.Error("expected error from Logout when repo fails")
		}
	})
}

func TestGetMe(t *testing.T) {
	validAdmin := &admin.Admin{
		ID:    "admin-1",
		Email: "admin@test.com",
		Name:  "Admin",
	}

	tests := []struct {
		name    string
		adminID string
		wantErr error
	}{
		{
			name:    "existing admin returns MeResponse",
			adminID: "admin-1",
		},
		{
			name:    "unknown admin ID returns ErrNotFound",
			adminID: "nonexistent",
			wantErr: admin.ErrNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := auth.NewService(newFakeAdminRepo(validAdmin), newFakeAuthRepo(), testSecret)

			resp, err := svc.GetMe(context.Background(), tc.adminID)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("GetMe() error = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("GetMe() unexpected error: %v", err)
			}
			if resp.ID != validAdmin.ID || resp.Email != validAdmin.Email || resp.Name != validAdmin.Name {
				t.Errorf("GetMe() = %+v, want %+v", resp, validAdmin)
			}
		})
	}
}
