package auth

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/revel/revel"
)

var secret = []byte(os.Getenv("JWT_SECRET"))

// Claims contains standard fields of claims and contains
// username to identify the user on request
type Claims struct {
	jwt.StandardClaims
	UserID   int
	Username string
}

// NewClaims creates custom claims given standard claim and username
func NewClaims(claims jwt.StandardClaims, id int, username string) *Claims {
	return &Claims{StandardClaims: claims, UserID: id, Username: username}
}

// Tokener is how the handlers will interface with tokens
type Tokener interface {
	NewToken(id int, username string) string
	CheckRequest(*revel.Request) (*Claims, error)
	GetClaims(token string) (*Claims, error)
	GetToken(r *revel.Request) (string, error)
}

// JWT holds the standard claims and has method
// that follow the Authoizor interface
type JWT struct {
	Claims jwt.StandardClaims
}

// NewJWT creates a new manager holding the standard claims for all tokens
func NewJWT() *JWT {
	return &JWT{
		Claims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			Issuer:    "Conduit",
		},
	}
}

// NewToken creates a new JWT with a 24 hour expire date
// and with the user's username in the claims
func (j *JWT) NewToken(id int, username string) string {
	claims := NewClaims(j.Claims, id, username)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, _ := token.SignedString([]byte(secret))
	return ss
}

// validateToken ensures that the tokenString provided is valid
// then returns the claims
func validateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("Token not valid")
	}

	return claims, nil
}

// CheckRequest ensures that the JWT provided in the header of
// the request is valid, and then returns claims
func (j JWT) CheckRequest(r *revel.Request) (*Claims, error) {
	token, err := j.GetToken(r)
	if err != nil {
		return nil, err
	}
	return j.GetClaims(token)
}
func (JWT) GetClaims(token string) (*Claims, error) {
	claims, err := validateToken(token)
	if err != nil {
		return nil, err
	}
	return claims, nil
}
func (JWT) GetToken(r *revel.Request) (string, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", fmt.Errorf("Authorization header is empty")
	}

	token := strings.TrimPrefix(auth, "Token ")
	return token, nil
}
