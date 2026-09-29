package admin_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"football-api/internal/modules/admin"

	"golang.org/x/crypto/bcrypt"
)

type fakeAdminRepo struct {
	admins   map[string]*admin.Admin
	byID     map[string]*admin.Admin
	nextID   int
	forceErr error
}

func newFakeAdminRepo() *fakeAdminRepo {
	return &fakeAdminRepo{
		admins: make(map[string]*admin.Admin),
		byID:   make(map[string]*admin.Admin),
	}
}

func (f *fakeAdminRepo) FindByEmail(_ context.Context, email string) (*admin.Admin, error) {
	if f.forceErr != nil {
		return nil, f.forceErr
	}
	a, ok := f.admins[email]
	if !ok {
		return nil, admin.ErrNotFound
	}
	return a, nil
}

func (f *fakeAdminRepo) FindByID(_ context.Context, id string) (*admin.Admin, error) {
	if f.forceErr != nil {
		return nil, f.forceErr
	}
	a, ok := f.byID[id]
	if !ok {
		return nil, admin.ErrNotFound
	}
	return a, nil
}

func (f *fakeAdminRepo) Create(_ context.Context, name, email, passwordHash string) (*admin.Admin, error) {
	if f.forceErr != nil {
		return nil, f.forceErr
	}
	if _, exists := f.admins[email]; exists {
		return nil, admin.ErrEmailConflict
	}
	f.nextID++
	id := string(rune('A' + f.nextID))
	a := &admin.Admin{ID: id, Name: name, Email: email, PasswordHash: passwordHash}
	f.admins[email] = a
	f.byID[id] = a
	return a, nil
}

func (f *fakeAdminRepo) ExistsAny(_ context.Context) (bool, error) {
	if f.forceErr != nil {
		return false, f.forceErr
	}
	return len(f.admins) > 0, nil
}

func newLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func TestSeedIfEmpty(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("secret1234"), bcrypt.MinCost)

	tests := []struct {
		name        string
		seedAdmins  []*admin.Admin
		seedName    string
		seedEmail   string
		seedPw      string
		wantErr     bool
		wantCreated bool
	}{
		{
			name:        "empty table seeds admin",
			seedName:    "Admin",
			seedEmail:   "admin@test.com",
			seedPw:      "secret1234",
			wantCreated: true,
		},
		{
			name: "non-empty table skips seed",
			seedAdmins: []*admin.Admin{
				{ID: "X", Name: "Existing", Email: "existing@test.com", PasswordHash: string(hash)},
			},
			seedName:    "Admin",
			seedEmail:   "admin@test.com",
			seedPw:      "secret1234",
			wantCreated: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := newFakeAdminRepo()
			for _, a := range tc.seedAdmins {
				repo.admins[a.Email] = a
				repo.byID[a.ID] = a
			}
			svc := admin.NewService(repo, newLogger())

			err := svc.SeedIfEmpty(context.Background(), tc.seedName, tc.seedEmail, tc.seedPw)
			if (err != nil) != tc.wantErr {
				t.Fatalf("SeedIfEmpty() error = %v, wantErr %v", err, tc.wantErr)
			}

			_, created := repo.admins[tc.seedEmail]
			if created != tc.wantCreated {
				t.Errorf("admin created = %v, want %v", created, tc.wantCreated)
			}

			if tc.wantCreated {
				a := repo.admins[tc.seedEmail]
				if err := bcrypt.CompareHashAndPassword([]byte(a.PasswordHash), []byte(tc.seedPw)); err != nil {
					t.Error("password hash does not match seed password")
				}
			}
		})
	}
}
