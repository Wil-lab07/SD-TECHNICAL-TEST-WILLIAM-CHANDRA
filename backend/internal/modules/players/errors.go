package players

import "errors"

var (
	ErrPlayerNotFound    = errors.New("player not found")
	ErrJerseyNumberTaken = errors.New("jersey number already taken in this team")
	ErrTeamNotFound      = errors.New("team not found or has been deleted")
)
