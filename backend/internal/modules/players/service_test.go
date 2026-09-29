package players_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"football-api/internal/modules/players"
	"football-api/internal/modules/teams"
)

type fakePlayerRepo struct {
	store    map[string]*players.Player
	nextID   int
	forceErr error
}

func newFakePlayerRepo() *fakePlayerRepo {
	return &fakePlayerRepo{store: make(map[string]*players.Player)}
}

func (f *fakePlayerRepo) Create(_ context.Context, teamID, name string, height, weight float64, position string, jerseyNumber int) (*players.Player, error) {
	if f.forceErr != nil {
		return nil, f.forceErr
	}
	for _, p := range f.store {
		if p.TeamID == teamID && p.JerseyNumber == jerseyNumber {
			return nil, players.ErrJerseyNumberTaken
		}
	}
	f.nextID++
	id := "player-" + string(rune('0'+f.nextID))
	p := &players.Player{
		ID: id, TeamID: teamID, Name: name,
		Height: height, Weight: weight, Position: position,
		JerseyNumber: jerseyNumber,
		CreatedAt:    time.Now(), UpdatedAt: time.Now(),
	}
	f.store[id] = p
	return p, nil
}

func (f *fakePlayerRepo) GetByID(_ context.Context, id string) (*players.Player, error) {
	if f.forceErr != nil {
		return nil, f.forceErr
	}
	p, ok := f.store[id]
	if !ok {
		return nil, players.ErrPlayerNotFound
	}
	return p, nil
}

func (f *fakePlayerRepo) List(_ context.Context, teamID *string, limit, offset int) ([]*players.Player, int, error) {
	if f.forceErr != nil {
		return nil, 0, f.forceErr
	}
	var all []*players.Player
	for _, p := range f.store {
		if teamID != nil && p.TeamID != *teamID {
			continue
		}
		all = append(all, p)
	}
	total := len(all)
	if offset >= total {
		return []*players.Player{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return all[offset:end], total, nil
}

func (f *fakePlayerRepo) Update(_ context.Context, id string, name *string, height, weight *float64, position *string, jerseyNumber *int) (*players.Player, error) {
	if f.forceErr != nil {
		return nil, f.forceErr
	}
	p, ok := f.store[id]
	if !ok {
		return nil, players.ErrPlayerNotFound
	}
	if name != nil {
		p.Name = *name
	}
	if height != nil {
		p.Height = *height
	}
	if weight != nil {
		p.Weight = *weight
	}
	if position != nil {
		p.Position = *position
	}
	if jerseyNumber != nil {
		for _, other := range f.store {
			if other.ID != id && other.TeamID == p.TeamID && other.JerseyNumber == *jerseyNumber {
				return nil, players.ErrJerseyNumberTaken
			}
		}
		p.JerseyNumber = *jerseyNumber
	}
	p.UpdatedAt = time.Now()
	return p, nil
}

func (f *fakePlayerRepo) SoftDelete(_ context.Context, id string) error {
	if f.forceErr != nil {
		return f.forceErr
	}
	if _, ok := f.store[id]; !ok {
		return players.ErrPlayerNotFound
	}
	delete(f.store, id)
	return nil
}

type fakeTeamRepo struct {
	exists map[string]bool
}

func newFakeTeamRepo(ids ...string) *fakeTeamRepo {
	f := &fakeTeamRepo{exists: make(map[string]bool)}
	for _, id := range ids {
		f.exists[id] = true
	}
	return f
}

func (f *fakeTeamRepo) Create(_ context.Context, _ string, _ *string, _ int, _, _ string) (*teams.Team, error) {
	return nil, nil
}

func (f *fakeTeamRepo) GetByID(_ context.Context, id string) (*teams.Team, error) {
	if !f.exists[id] {
		return nil, teams.ErrTeamNotFound
	}
	return &teams.Team{ID: id, Name: "Fake Team"}, nil
}

func (f *fakeTeamRepo) List(_ context.Context, _, _ int) ([]*teams.Team, int, error) {
	return nil, 0, nil
}

func (f *fakeTeamRepo) Update(_ context.Context, _ string, _ *string, _ *string, _ *int, _ *string, _ *string) (*teams.Team, error) {
	return nil, nil
}

func (f *fakeTeamRepo) SoftDelete(_ context.Context, _ string) error {
	return nil
}

func ptr[T any](v T) *T { return &v }

const validTeamID = "team-aaa"

func TestCreatePlayer(t *testing.T) {
	tests := []struct {
		name    string
		req     players.CreatePlayerRequest
		wantErr error
	}{
		{
			name: "valid request creates player",
			req: players.CreatePlayerRequest{
				TeamID: validTeamID, Name: "Budi", Height: 175.5, Weight: 70.0,
				Position: "striker", JerseyNumber: 9,
			},
		},
		{
			name: "non-existent team returns ErrTeamNotFound",
			req: players.CreatePlayerRequest{
				TeamID: "team-unknown", Name: "Andi", Height: 170, Weight: 65,
				Position: "midfielder", JerseyNumber: 10,
			},
			wantErr: players.ErrTeamNotFound,
		},
		{
			name: "duplicate jersey number returns ErrJerseyNumberTaken",
			req: players.CreatePlayerRequest{
				TeamID: validTeamID, Name: "Citra", Height: 168, Weight: 63,
				Position: "defender", JerseyNumber: 9, // same as first test
			},
			wantErr: players.ErrJerseyNumberTaken,
		},
	}

	repo := newFakePlayerRepo()
	checker := newFakeTeamRepo(validTeamID)
	svc := players.NewService(repo, checker)

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
			if resp.TeamID != tc.req.TeamID || resp.JerseyNumber != tc.req.JerseyNumber {
				t.Errorf("Create() = %+v, unexpected fields", resp)
			}
		})
	}
}

func TestGetPlayerByID(t *testing.T) {
	repo := newFakePlayerRepo()
	checker := newFakeTeamRepo(validTeamID)
	svc := players.NewService(repo, checker)

	created, _ := svc.Create(context.Background(), players.CreatePlayerRequest{
		TeamID: validTeamID, Name: "Doni", Height: 180, Weight: 75,
		Position: "goalkeeper", JerseyNumber: 1,
	})

	tests := []struct {
		name    string
		id      string
		wantErr error
	}{
		{name: "existing id returns player", id: created.ID},
		{name: "unknown id returns ErrPlayerNotFound", id: "nonexistent", wantErr: players.ErrPlayerNotFound},
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

func TestUpdatePlayer(t *testing.T) {
	repo := newFakePlayerRepo()
	checker := newFakeTeamRepo(validTeamID)
	svc := players.NewService(repo, checker)

	created, _ := svc.Create(context.Background(), players.CreatePlayerRequest{
		TeamID: validTeamID, Name: "Eko", Height: 172, Weight: 68,
		Position: "midfielder", JerseyNumber: 7,
	})

	tests := []struct {
		name     string
		id       string
		req      players.UpdatePlayerRequest
		wantErr  error
		wantName string
		wantPos  string
	}{
		{
			name:     "partial update only changes sent fields",
			id:       created.ID,
			req:      players.UpdatePlayerRequest{Position: ptr("defender")},
			wantName: "Eko",
			wantPos:  "defender",
		},
		{
			name:    "unknown id returns ErrPlayerNotFound",
			id:      "nonexistent",
			req:     players.UpdatePlayerRequest{Name: ptr("X")},
			wantErr: players.ErrPlayerNotFound,
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
			if resp.Position != tc.wantPos {
				t.Errorf("Update() position = %v, want %v", resp.Position, tc.wantPos)
			}
		})
	}
}

func TestSoftDeletePlayer(t *testing.T) {
	repo := newFakePlayerRepo()
	checker := newFakeTeamRepo(validTeamID)
	svc := players.NewService(repo, checker)

	created, _ := svc.Create(context.Background(), players.CreatePlayerRequest{
		TeamID: validTeamID, Name: "Fajar", Height: 176, Weight: 72,
		Position: "striker", JerseyNumber: 11,
	})

	tests := []struct {
		name    string
		id      string
		wantErr error
	}{
		{name: "existing player is soft deleted", id: created.ID},
		{name: "already deleted returns ErrPlayerNotFound", id: created.ID, wantErr: players.ErrPlayerNotFound},
		{name: "unknown id returns ErrPlayerNotFound", id: "nonexistent", wantErr: players.ErrPlayerNotFound},
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
