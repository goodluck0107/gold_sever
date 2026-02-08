package client

import "github.com/dgrijalva/jwt-go"

type JwtClaims struct {
	jwt.StandardClaims
}

type JwtHeader struct {
	Kid string `json:"kid"`
	Alg string `json:"alg"`
}

type JwtKeys struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}
