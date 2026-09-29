package matches_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"football-api/internal/modules/matches"
	"football-api/internal/modules/players"
	"football-api/internal/modules/teams"
)

type fakeMatchRepo struct {
	store    map[string]*matches.Match
	goals    map[string][]matches.GoalItemRequest
	nextID   int
	forceErr error
}

func newFakeMatchRepo() *fakeMatchRepo {
	return &fakeMatchRepo{
		store: make(map[string]*matches.Match),
		goals: make(map[string][]matches.GoalItemRequest),
	}
}

func (f *fakeMatchRepo) Create(_ context.Context, homeTeamID, awayTeamID string, matchDate time.Time, matchTime string) (*matches.Match, error) {
	if f.forceErr != nil {
		return nil, f.forceErr
	}
	f.nextID++
	id := fmt.Sprintf("match-%d", f.nextID)
	m := &matches.Match{
		ID:         id,
		HomeTeamID: homeTeamID,
		AwayTeamID: awayTeamID,
		MatchDate:  matchDate,
		MatchTime:  matchTime,
		HomeScore:  0,
		AwayScore:  0,
		Status:     "scheduled",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	f.store[id] = m
	return m, nil
}

func (f *fakeMatchRepo) GetByID(_ context.Context, id string) (*matches.Match, error) {
	if f.forceErr != nil {
		return nil, f.forceErr
	}
	m, ok := f.store[id]
	if !ok {
		return nil, matches.ErrMatchNotFound
	}
	return m, nil
}

func (f *fakeMatchRepo) List(_ context.Context, limit, offset int) ([]*matches.Match, int, error) {
	if f.forceErr != nil {
		return nil, 0, f.forceErr
	}
	var all []*matches.Match
	for _, m := range f.store {
		all = append(all, m)
	}
	total := len(all)
	if offset >= total {
		return []*matches.Match{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return all[offset:end], total, nil
}

func (f *fakeMatchRepo) ReportResult(_ context.Context, matchID string, homeScore, awayScore int, goals []matches.GoalItemRequest) (*matches.Match, error) {
	if f.forceErr != nil {
		return nil, f.forceErr
	}
	m, ok := f.store[matchID]
	if !ok {
		return nil, matches.ErrMatchNotFound
	}
	m.HomeScore = homeScore
	m.AwayScore = awayScore
	m.Status = "finished"
	m.UpdatedAt = time.Now()
	f.goals[matchID] = goals
	return m, nil
}

func (f *fakeMatchRepo) UpdateResult(_ context.Context, matchID string, homeScore, awayScore int, goals []matches.GoalItemRequest) (*matches.Match, error) {
	return f.ReportResult(context.Background(), matchID, homeScore, awayScore, goals)
}

func (f *fakeMatchRepo) GetMatchReport(_ context.Context, matchID string) (*matches.MatchReportResponse, error) {
	m, ok := f.store[matchID]
	if !ok {
		return nil, matches.ErrMatchNotFound
	}
	return &matches.MatchReportResponse{
		MatchID:            m.ID,
		MatchDate:          m.MatchDate.Format("2006-01-02"),
		MatchTime:          m.MatchTime,
		HomeTeamName:       "Home Team",
		AwayTeamName:       "Away Team",
		FinalScore:         fmt.Sprintf("%d - %d", m.HomeScore, m.AwayScore),
		HomeScore:          m.HomeScore,
		AwayScore:          m.AwayScore,
		Status:             "Home Win",
		TopScorers:         []matches.TopScorerResponse{},
		CumulativeHomeWins: 1,
		CumulativeAwayWins: 0,
	}, nil
}

func (f *fakeMatchRepo) SoftDelete(_ context.Context, id string) error {
	if f.forceErr != nil {
		return f.forceErr
	}
	if _, ok := f.store[id]; !ok {
		return matches.ErrMatchNotFound
	}
	delete(f.store, id)
	return nil
}

type fakeTeamRepo struct {
	teams map[string]*teams.Team
}

func newFakeTeamRepo(activeIDs ...string) *fakeTeamRepo {
	f := &fakeTeamRepo{teams: make(map[string]*teams.Team)}
	for _, id := range activeIDs {
		f.teams[id] = &teams.Team{ID: id, Name: "Team " + id}
	}
	return f
}

func (f *fakeTeamRepo) Create(_ context.Context, _ string, _ *string, _ int, _, _ string) (*teams.Team, error) {
	return nil, nil
}

func (f *fakeTeamRepo) GetByID(_ context.Context, id string) (*teams.Team, error) {
	t, ok := f.teams[id]
	if !ok {
		return nil, teams.ErrTeamNotFound
	}
	return t, nil
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

type fakePlayerRepo struct {
	players map[string]*players.Player
}

func newFakePlayerRepo(plrs ...*players.Player) *fakePlayerRepo {
	f := &fakePlayerRepo{players: make(map[string]*players.Player)}
	for _, p := range plrs {
		f.players[p.ID] = p
	}
	return f
}

func (f *fakePlayerRepo) Create(_ context.Context, _, _ string, _, _ float64, _ string, _ int) (*players.Player, error) {
	return nil, nil
}

func (f *fakePlayerRepo) GetByID(_ context.Context, id string) (*players.Player, error) {
	p, ok := f.players[id]
	if !ok {
		return nil, players.ErrPlayerNotFound
	}
	return p, nil
}

func (f *fakePlayerRepo) List(_ context.Context, _ *string, _, _ int) ([]*players.Player, int, error) {
	return nil, 0, nil
}

func (f *fakePlayerRepo) Update(_ context.Context, _ string, _ *string, _ *float64, _ *float64, _ *string, _ *int) (*players.Player, error) {
	return nil, nil
}

func (f *fakePlayerRepo) SoftDelete(_ context.Context, _ string) error {
	return nil
}

const (
	homeTeamID = "team-home"
	awayTeamID = "team-away"
	playerID1  = "player-1"
)

func TestCreateMatch(t *testing.T) {
	tests := []struct {
		name    string
		req     matches.CreateMatchRequest
		wantErr error
	}{
		{
			name: "valid schedule match",
			req: matches.CreateMatchRequest{
				HomeTeamID: homeTeamID,
				AwayTeamID: awayTeamID,
				MatchDate:  "2026-10-15",
				MatchTime:  "15:30:00",
			},
		},
		{
			name: "self play returns ErrSelfPlayNotAllowed",
			req: matches.CreateMatchRequest{
				HomeTeamID: homeTeamID,
				AwayTeamID: homeTeamID,
				MatchDate:  "2026-10-15",
				MatchTime:  "15:30:00",
			},
			wantErr: matches.ErrSelfPlayNotAllowed,
		},
		{
			name: "deactivated team returns ErrTeamDeactivated",
			req: matches.CreateMatchRequest{
				HomeTeamID: homeTeamID,
				AwayTeamID: "team-unknown",
				MatchDate:  "2026-10-15",
				MatchTime:  "15:30:00",
			},
			wantErr: matches.ErrTeamDeactivated,
		},
	}

	matchRepo := newFakeMatchRepo()
	teamRepo := newFakeTeamRepo(homeTeamID, awayTeamID)
	playerRepo := newFakePlayerRepo()
	svc := matches.NewService(matchRepo, teamRepo, playerRepo)

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
			if resp.HomeTeamID != tc.req.HomeTeamID || resp.AwayTeamID != tc.req.AwayTeamID {
				t.Errorf("Create() = %+v, unexpected fields", resp)
			}
		})
	}
}

func TestReportResult(t *testing.T) {
	matchRepo := newFakeMatchRepo()
	teamRepo := newFakeTeamRepo(homeTeamID, awayTeamID)
	playerRepo := newFakePlayerRepo(&players.Player{ID: playerID1, TeamID: homeTeamID, Name: "Striker 1"})
	svc := matches.NewService(matchRepo, teamRepo, playerRepo)

	created, _ := svc.Create(context.Background(), matches.CreateMatchRequest{
		HomeTeamID: homeTeamID,
		AwayTeamID: awayTeamID,
		MatchDate:  "2026-10-15",
		MatchTime:  "15:30:00",
	})

	tests := []struct {
		name    string
		matchID string
		req     matches.ReportResultRequest
		wantErr error
	}{
		{
			name:    "valid result report",
			matchID: created.ID,
			req: matches.ReportResultRequest{
				HomeScore: 1,
				AwayScore: 0,
				Goals: []matches.GoalItemRequest{
					{PlayerID: playerID1, TeamID: homeTeamID, Minute: 45},
				},
			},
		},
		{
			name:    "score mismatch returns ErrScoreMismatch",
			matchID: created.ID,
			req: matches.ReportResultRequest{
				HomeScore: 2, // 2 goals reported in score but only 1 goal item
				AwayScore: 0,
				Goals: []matches.GoalItemRequest{
					{PlayerID: playerID1, TeamID: homeTeamID, Minute: 45},
				},
			},
			wantErr: matches.ErrScoreMismatch,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m, _ := svc.Create(context.Background(), matches.CreateMatchRequest{
				HomeTeamID: homeTeamID,
				AwayTeamID: awayTeamID,
				MatchDate:  "2026-10-15",
				MatchTime:  "15:30:00",
			})

			resp, err := svc.ReportResult(context.Background(), m.ID, tc.req)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("ReportResult() error = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ReportResult() unexpected error: %v", err)
			}
			if resp.Status != "finished" || resp.HomeScore != tc.req.HomeScore {
				t.Errorf("ReportResult() = %+v, want status finished and homeScore %d", resp, tc.req.HomeScore)
			}
		})
	}

	finishedMatch, _ := svc.Create(context.Background(), matches.CreateMatchRequest{
		HomeTeamID: homeTeamID,
		AwayTeamID: awayTeamID,
		MatchDate:  "2026-10-15",
		MatchTime:  "15:30:00",
	})
	_, _ = svc.ReportResult(context.Background(), finishedMatch.ID, matches.ReportResultRequest{
		HomeScore: 1,
		AwayScore: 0,
		Goals: []matches.GoalItemRequest{
			{PlayerID: playerID1, TeamID: homeTeamID, Minute: 10},
		},
	})

	_, err := svc.ReportResult(context.Background(), finishedMatch.ID, matches.ReportResultRequest{
		HomeScore: 1,
		AwayScore: 0,
		Goals: []matches.GoalItemRequest{
			{PlayerID: playerID1, TeamID: homeTeamID, Minute: 10},
		},
	})
	if !errors.Is(err, matches.ErrMatchAlreadyFinished) {
		t.Fatalf("ReportResult on finished match = %v, want ErrMatchAlreadyFinished", err)
	}
}
