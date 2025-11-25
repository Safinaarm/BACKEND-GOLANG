// postgre/jwt.go
package jwt

import (
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

type JWTService interface {
	GenerateToken(userID, roleID, roleName string) (string, error)
	ValidateToken(tokenString string) (*jwtlib.Token, error)
}

type jwtService struct {
	secretKey string
}

func NewJWTService(secret string) JWTService {
	return &jwtService{
		secretKey: secret,
	}
}

// Claims JWT
type JWTCustomClaim struct {
	UserID  string `json:"userId"`
	RoleID  string `json:"roleId"`
	Role    string `json:"role"`
	jwtlib.RegisteredClaims
}

func (j *jwtService) GenerateToken(userID, roleID, roleName string) (string, error) {
	claims := &JWTCustomClaim{
		UserID:  userID,
		RoleID:  roleID,
		Role:    roleName,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour * 24)),
			IssuedAt:  jwtlib.NewNumericDate(time.Now()),
		},
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

func (j *jwtService) ValidateToken(tokenString string) (*jwtlib.Token, error) {
	return jwtlib.ParseWithClaims(tokenString, &JWTCustomClaim{}, func(token *jwtlib.Token) (interface{}, error) {
		return []byte(j.secretKey), nil
	})
}