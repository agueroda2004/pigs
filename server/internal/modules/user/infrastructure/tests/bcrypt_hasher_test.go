package tests

import (
	"testing"

	userinfra "server/internal/modules/user/infrastructure"
)

func TestBcryptPasswordHasherVerify(t *testing.T) {
	hasher := userinfra.NewBcryptPasswordHasher(4)

	hash, err := hasher.Hash("secret")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if err := hasher.Verify(hash, "secret"); err != nil {
		t.Fatalf("Verify() correct password error = %v", err)
	}

	if err := hasher.Verify(hash, "wrong"); err == nil {
		t.Fatal("Verify() wrong password should return an error")
	}
}
