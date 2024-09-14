package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

func main() {
	privateKeyPath := flag.String("key", "key.rsa", "Filename for the private key")
	issuer := flag.String("issuer", "localhost:8080", "Issuer")
	expiresIn := flag.String("expires-in", "never", "Expires in duration (e.g. 1h)")
	flag.Parse()

	// Read private key from file.
	keyPEM, err := os.ReadFile(*privateKeyPath)
	if err != nil {
		log.Fatalf("failed to read private key from file: %v", err)
	}

	// Parse private key.
	privateKey, err := jwk.ParseKey(keyPEM, jwk.WithPEM(true))
	if err != nil {
		panic(err)
	}

	// Build token.
	builder := jwt.NewBuilder().
		Issuer(*issuer).
		IssuedAt(time.Now()).
		Claim("scopes", []string{"extension:read"})

	// Set expiration if not "never".
	if *expiresIn != "never" {
		expires, err := time.ParseDuration(*expiresIn)
		if err != nil {
			log.Fatalf("failed to parse expires in duration: %v", err)
		}
		builder = builder.Expiration(time.Now().Add(expires))
	}

	// Create token.
	token, err := builder.Build()
	if err != nil {
		log.Fatalf("failed to build token: %v", err)
	}

	// Sign token.
	signed, err := jwt.Sign(token, jwt.WithKey(jwa.RS256, privateKey))
	if err != nil {
		log.Fatalf("failed to sign token: %v", err)
	}

	// Print signed token.
	fmt.Println(string(signed))
}
