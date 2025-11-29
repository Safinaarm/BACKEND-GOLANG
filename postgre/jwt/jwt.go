// postgre/jwt/jwt.go
package jwt

import (
	"time"

	jwtpkg "github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	UserID string `json:"userId"`
	RoleID string `json:"roleId"`
	Role   string `json:"role"`
	jwtpkg.RegisteredClaims `json:",inline"`
}

type JWTService interface {
	GenerateToken(userID, roleID, role string) (string, error)
	ValidateToken(tokenStr string) (*jwtpkg.Token, error)
	CheckPasswordHash(password, hash string) bool
	HashPassword(password string) (string, error)
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
		RegisteredClaims: jwtpkg.RegisteredClaims{
			ExpiresAt: jwtpkg.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24h for demo; make refresh longer in prod
			IssuedAt:  jwtpkg.NewNumericDate(time.Now()),
		},
	}
	token := jwtpkg.NewWithClaims(jwtpkg.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *jwtService) ValidateToken(tokenStr string) (*jwtpkg.Token, error) {
	return jwtpkg.ParseWithClaims(tokenStr, &Claims{}, func(token *jwtpkg.Token) (interface{}, error) {
		return s.secret, nil
	})
}

func (s *jwtService) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (s *jwtService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}