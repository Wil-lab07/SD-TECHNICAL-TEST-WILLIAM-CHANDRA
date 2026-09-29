package teams_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"football-api/internal/modules/teams"
	"football-api/pkg/respond"
)

type fakeTeamRepo struct {
	store    map[string]*teams.Team
	nextID   int
	forceErr error
}

func newFakeTeamRepo() *fakeTeamRepo {
	return &fakeTeamRepo{store: make(map[string]*teams.Team)}
}

func (f *fakeTeamRepo) Create(_ context.Context, name string, logoURL *string, foundedYear int, address, city string) (*teams.Team, error) {
	if f.forceErr != nil {
		return nil, f.forceErr
	}
	for _, t := range f.store {
		if t.Name == name {
			return nil, teams.ErrTeamNameConflict
		}
	}
	f.nextID++
	id := "team-" + string(rune('0'+f.nextID))
	t := &teams.Team{
		ID:          id,
		Name:        name,
		LogoURL:     logoURL,
		FoundedYear: foundedYear,
		Address:     address,
		City:        city,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	f.store[id] = t
	return t, nil
}

func (f *fakeTeamRepo) GetByID(_ context.Context, id string) (*teams.Team, error) {
	if f.forceErr != nil {
		return nil, f.forceErr
	}
	t, ok := f.store[id]
	if !ok {
		return nil, teams.ErrTeamNotFound
	}
	return t, nil
}

func (f *fakeTeamRepo) List(_ context.Context, limit, offset int) ([]*teams.Team, int, error) {
	if f.forceErr != nil {
		return nil, 0, f.forceErr
	}
	all := make([]*teams.Team, 0, len(f.store))
	for _, t := range f.store {
		all = append(all, t)
	}
	total := len(all)
	if offset >= total {
		return []*teams.Team{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return all[offset:end], total, nil
}

func (f *fakeTeamRepo) Update(_ context.Context, id string, name *string, logoURL *string, foundedYear *int, address *string, city *string) (*teams.Team, error) {
	if f.forceErr != nil {
		return nil, f.forceErr
	}
	t, ok := f.store[id]
	if !ok {
		return nil, teams.ErrTeamNotFound
	}
	if name != nil {
		t.Name = *name
	}
	if logoURL != nil {
		t.LogoURL = logoURL
	}
	if foundedYear != nil {
		t.FoundedYear = *foundedYear
	}
	if address != nil {
		t.Address = *address
	}
	if city != nil {
		t.City = *city
	}
	t.UpdatedAt = time.Now()
	return t, nil
}

func (f *fakeTeamRepo) SoftDelete(_ context.Context, id string) error {
	if f.forceErr != nil {
		return f.forceErr
	}
	if _, ok := f.store[id]; !ok {
		return teams.ErrTeamNotFound
	}
	delete(f.store, id)
	return nil
}

func ptr[T any](v T) *T { return &v }

func TestCreateTeam(t *testing.T) {
	tests := []struct {
		name    string
		req     teams.CreateTeamRequest
		wantErr error
	}{
		{
			name: "valid request creates team",
			req: teams.CreateTeamRequest{
				Name:        "Persija",
				FoundedYear: 1928,
				Address:     "Jakarta",
				City:        "Jakarta",
			},
		},
		{
			name: "future founded year returns ErrInvalidFoundedYear",
			req: teams.CreateTeamRequest{
				Name:        "Future FC",
				FoundedYear: time.Now().Year() + 1,
				Address:     "Jakarta",
				City:        "Jakarta",
			},
			wantErr: teams.ErrInvalidFoundedYear,
		},
		{
			name: "duplicate name returns ErrTeamNameConflict",
			req: teams.CreateTeamRequest{
				Name:        "Persija",
				FoundedYear: 1928,
				Address:     "Jakarta",
				City:        "Jakarta",
			},
			wantErr: teams.ErrTeamNameConflict,
		},
	}

	repo := newFakeTeamRepo()
	svc := teams.NewService(repo)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := svc.Create(context.Background(), tc.req)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("Create() error = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Create() unexpected error: %v", err)
			}
			if resp.Name != tc.req.Name {
				t.Errorf("Create() name = %v, want %v", resp.Name, tc.req.Name)
			}
			if resp.FoundedYear != tc.req.FoundedYear {
				t.Errorf("Create() founded_year = %v, want %v", resp.FoundedYear, tc.req.FoundedYear)
			}
		})
	}
}

func TestGetTeamByID(t *testing.T) {
	repo := newFakeTeamRepo()
	svc := teams.NewService(repo)

	created, _ := svc.Create(context.Background(), teams.CreateTeamRequest{
		Name: "Arema", FoundedYear: 1987, Address: "Malang", City: "Malang",
	})

	tests := []struct {
		name    string
		id      string
		wantErr error
	}{
		{name: "existing id returns team", id: created.ID},
		{name: "unknown id returns ErrTeamNotFound", id: "nonexistent", wantErr: teams.ErrTeamNotFound},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := svc.GetByID(context.Background(), tc.id)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("GetByID() error = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("GetByID() unexpected error: %v", err)
			}
			if resp.ID != tc.id {
				t.Errorf("GetByID() id = %v, want %v", resp.ID, tc.id)
			}
		})
	}
}

func TestUpdateTeam(t *testing.T) {
	repo := newFakeTeamRepo()
	svc := teams.NewService(repo)

	created, _ := svc.Create(context.Background(), teams.CreateTeamRequest{
		Name: "Persib", FoundedYear: 1933, Address: "Bandung", City: "Bandung",
	})

	tests := []struct {
		name      string
		id        string
		req       teams.UpdateTeamRequest
		wantErr   error
		wantName  string
		wantCity  string
	}{
		{
			name:     "partial update only changes sent fields",
			id:       created.ID,
			req:      teams.UpdateTeamRequest{City: ptr("Kota Bandung")},
			wantName: "Persib",
			wantCity: "Kota Bandung",
		},
		{
			name:    "future founded year returns ErrInvalidFoundedYear",
			id:      created.ID,
			req:     teams.UpdateTeamRequest{FoundedYear: ptr(time.Now().Year() + 1)},
			wantErr: teams.ErrInvalidFoundedYear,
		},
		{
			name:    "unknown id returns ErrTeamNotFound",
			id:      "nonexistent",
			req:     teams.UpdateTeamRequest{City: ptr("X")},
			wantErr: teams.ErrTeamNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := svc.Update(context.Background(), tc.id, tc.req)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("Update() error = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Update() unexpected error: %v", err)
			}
			if resp.Name != tc.wantName {
				t.Errorf("Update() name = %v, want %v", resp.Name, tc.wantName)
			}
			if resp.City != tc.wantCity {
				t.Errorf("Update() city = %v, want %v", resp.City, tc.wantCity)
			}
		})
	}
}

func TestSoftDeleteTeam(t *testing.T) {
	repo := newFakeTeamRepo()
	svc := teams.NewService(repo)

	created, _ := svc.Create(context.Background(), teams.CreateTeamRequest{
		Name: "PSM", FoundedYear: 1915, Address: "Makassar", City: "Makassar",
	})

	tests := []struct {
		name    string
		id      string
		wantErr error
	}{
		{name: "existing team is soft deleted", id: created.ID},
		{name: "already deleted returns ErrTeamNotFound", id: created.ID, wantErr: teams.ErrTeamNotFound},
		{name: "unknown id returns ErrTeamNotFound", id: "nonexistent", wantErr: teams.ErrTeamNotFound},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := svc.SoftDelete(context.Background(), tc.id)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("SoftDelete() error = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("SoftDelete() unexpected error: %v", err)
			}
		})
	}
}

func TestListTeams(t *testing.T) {
	repo := newFakeTeamRepo()
	svc := teams.NewService(repo)

	for _, name := range []string{"Team A", "Team B", "Team C"} {
		_, _ = svc.Create(context.Background(), teams.CreateTeamRequest{
			Name: name, FoundedYear: 2000, Address: "Jakarta", City: "Jakarta",
		})
	}

	t.Run("returns paginated results", func(t *testing.T) {
		result, meta, err := svc.List(context.Background(), 1, 2)
		if err != nil {
			t.Fatalf("List() unexpected error: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("List() len = %v, want 2", len(result))
		}
		if meta.TotalRecords != 3 {
			t.Errorf("List() TotalRecords = %v, want 3", meta.TotalRecords)
		}
		if meta.TotalPages != 2 {
			t.Errorf("List() TotalPages = %v, want 2", meta.TotalPages)
		}
	})

	t.Run("returns correct PaginationMeta type", func(t *testing.T) {
		_, meta, err := svc.List(context.Background(), 1, 10)
		if err != nil {
			t.Fatalf("List() unexpected error: %v", err)
		}
		var _ respond.PaginationMeta = meta
	})
}
