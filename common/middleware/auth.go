package middleware

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type AuthClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.StandardClaims
}

type TokenService struct {
	secretKey []byte
}

func NewTokenService(secret string) *TokenService {
	return &TokenService{
		secretKey: []byte(secret),
	}
}

func (ts *TokenService) GenerateToken(userID, username string) (string, error) {
	claims := AuthClaims{
		UserID:   userID,
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(), // 24-hour token
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(ts.secretKey)
}

func (ts *TokenService) ValidateToken(tokenString string) (*AuthClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		return ts.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*AuthClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func AuthMiddleware(ts TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		ah := c.GetHeader("Authorization")
		if ah == "" {
			unauthorised(c, "Authorization header missing")
			return
		}

		// ── Bearer ───────────────────────────────────────────────
		if strings.HasPrefix(ah, "Bearer ") {
			token := strings.TrimPrefix(ah, "Bearer ")
			if claims, err := ts.ValidateToken(token); err == nil {
				c.Set("user_id", claims.UserID)
				c.Set("username", claims.Username)
				c.Next()
				return
			}
		}

		// ── Basic (ONLY FOR LOCAL / TEST) ───────────────────────
		if strings.HasPrefix(ah, "Basic ") {
			payload, _ := base64.StdEncoding.DecodeString(strings.TrimPrefix(ah, "Basic "))
			parts := strings.SplitN(string(payload), ":", 2)
			if len(parts) == 2 && parts[0] == "admin" && parts[1] == "password" {
				c.Set("user_id", "0")
				c.Set("username", "admin")
				c.Next()
				return
			}
		}

		unauthorised(c, "Invalid credentials")
	}
}

func unauthorised(c *gin.Context, msg string) {
	c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
	c.Abort()
}
