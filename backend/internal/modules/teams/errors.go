package teams

import "errors"

var (
	ErrTeamNotFound       = errors.New("team not found")
	ErrTeamNameConflict   = errors.New("team name already exists")
	ErrInvalidFoundedYear = errors.New("founded year cannot be in the future")
)
