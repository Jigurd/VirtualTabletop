package main

import (
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
