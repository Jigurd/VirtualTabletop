package web

import (
	//jwt "github.com/dgrijalva/jwt-go"
	"time"
)

// json web token structure  Header.Payload.Signature

// Header identifies what algorithm is used to generate the signature
// HS256 indicatess that this algorithm has used HMAC with SHA-256
// HMAC is a symetric algorithm
type Header struct {
	Algorithm string `json:"alg"`
	Type      string `json:"HS256"`
}

// Payload that we will use when making claim
type Payload struct {
	LoggedInAs string    `json:"user"`
	Exp        time.Time `json:"exp"` // Expiration time
}

func CreateToken(userName string) {

	// 1. Create Time Stamp
	// 2. Create Signature using HS256 or RSA
	// 3. Store username, Sig + time stamp in db
	// 4. Send sig to user

}