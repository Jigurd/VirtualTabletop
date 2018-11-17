package web

import (
	//jwt "github.com/dgrijalva/jwt-go"
	"time"
)

// User structs
type User struct {
	UserName string `json:"username"`
	jwt.StandardClaims
}

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

// CreateToken creates a token so the user dont have to logginn severaltimes
func CreateToken(InnuserName string) string {

	// 1. Create Time Stamp
	// 2. Create Signature using HS256 or RSA
	// 3. Store username, Sig + time stamp in db
	// 4. Send sig to user

	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), &User{
		UserName: InnuserName,
	})

	// token -> string. Only server knows this secret (sword).
	tokenstring, err := token.SignedString([]byte("sword"))
	if err != nil {
		log.Fatalln(err)
	}
	return tokenstring

}
