package matches

import "errors"

var (
	ErrMatchNotFound        = errors.New("match not found")
	ErrSelfPlayNotAllowed   = errors.New("team cannot play against itself")
	ErrTeamDeactivated      = errors.New("team is deactivated")
	ErrMatchAlreadyFinished = errors.New("match result already reported, use PUT to update")
	ErrMatchNotScheduled    = errors.New("match result can only be reported for a scheduled match")
	ErrScoreMismatch        = errors.New("score does not match number of goals recorded")
	ErrPlayerNotOnTeam      = errors.New("player does not belong to specified team in this match")
	ErrPlayerDeactivated    = errors.New("player is deactivated")
)
