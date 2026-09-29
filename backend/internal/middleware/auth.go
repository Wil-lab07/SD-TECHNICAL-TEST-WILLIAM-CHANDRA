package middleware

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	authmod "football-api/internal/modules/auth"
	"football-api/pkg/respond"
)

const claimsKey = "claims"

func Auth(jwtSecret string, authRepo authmod.Repository, log *slog.Logger) gin.HandlerFunc {
	secretBytes := []byte(jwtSecret)

	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		
		if !strings.HasPrefix(header, "Bearer ") {
			respond.Unauthorized(c, authmod.ErrTokenMissing.Error())
			c.Abort()
			return
		}
		
		raw := strings.TrimPrefix(header, "Bearer ")

		claims := &authmod.Claims{}
		
		token, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, authmod.ErrTokenInvalid
			}
			return secretBytes, nil
		})
		
		if err != nil || !token.Valid {
			if errors.Is(err, jwt.ErrTokenExpired) {
				respond.Unauthorized(c, "token has expired")
			} else {
				respond.Unauthorized(c, authmod.ErrTokenInvalid.Error())
			}
			c.Abort()
			return
		}

		if claims.ID == "" {
			respond.Unauthorized(c, authmod.ErrTokenInvalid.Error())
			c.Abort()
			return
		}

		revoked, err := authRepo.IsRevoked(c.Request.Context(), claims.ID)
		
		if err != nil {
			log.Error("middleware.Auth IsRevoked", slog.String("error", err.Error()))
		
			respond.InternalError(c, "internal server error")
		
			c.Abort()
		
			return
		}
		
		if revoked {
			respond.Unauthorized(c, authmod.ErrTokenRevoked.Error())
			c.Abort()
			return
		}

		c.Set(claimsKey, claims)
		
		c.Next()
	}
}
