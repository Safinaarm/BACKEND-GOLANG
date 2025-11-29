// postgre/jwt/jwt.go
package jwt

import (
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	UserID string `json:"userId"`
	RoleID string `json:"roleId"`
	Role   string `json:"role"`
	jwt.RegisteredClaims `json:",inline"`
}

type JWTService interface {
	GenerateToken(userID, roleID, role string) (string, error)
	ValidateToken(tokenStr string) (*jwt.Token, error)
	CheckPasswordHash(password, hash string) bool
}

type jwtService struct {
	secret []byte
}

func NewJWTService(secret string) JWTService {
	return &jwtService{secret: []byte(secret)}
}

func (s *jwtService) GenerateToken(userID, roleID, role string) (string, error) {
	claims := &Claims{
		UserID: userID,
		RoleID: roleID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24h for demo; make refresh longer in prod
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *jwtService) ValidateToken(tokenStr string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return s.secret, nil
	})
}

func (s *jwtService) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}