package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	filename := flag.String("filename", "key", "Filename for the private and public key")
	bitSize := flag.Int("bit-size", 4096, "Bit size for the RSA key")
	force := flag.Bool("force", false, "Force generation of new key pair")
	flag.Parse()

	privateKeyFilename := *filename + ".rsa"
	publicKeyFilename := *filename + ".rsa.pub"

	// If private key file exists, do nothing.
	if _, err := os.Stat(privateKeyFilename); err == nil && !*force {
		return
	}

	key, err := rsa.GenerateKey(rand.Reader, *bitSize)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to generate RSA key: %w", err))
	}

	pub := key.Public()

	// Encode private key to PKCS#1 ASN.1 PEM.
	keyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)

	// Encode public key to PKCS#1 ASN.1 PEM.
	pubPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(pub.(*rsa.PublicKey)),
		},
	)

	// Write private key to file.
	if err := os.WriteFile(privateKeyFilename, keyPEM, 0700); err != nil {
		log.Fatal(fmt.Errorf("failed to write private key to file: %w", err))
	}

	// Write public key to file.
	if err := os.WriteFile(publicKeyFilename, pubPEM, 0755); err != nil {
		log.Fatal(fmt.Errorf("failed to write public key to file: %w", err))
	}
}
