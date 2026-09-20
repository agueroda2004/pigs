package infrastructure

import "golang.org/x/crypto/bcrypt"

type BcryptPasswordHasher struct {
	cost int
}

// NewBcryptPasswordHasher creates a hasher using the given bcrypt cost.
// When cost is zero it falls back to bcrypt.DefaultCost.
func NewBcryptPasswordHasher(cost int) *BcryptPasswordHasher {
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}
	return &BcryptPasswordHasher{cost: cost}
}

// Hash generates a bcrypt hash for the given plain-text password.
// It returns an error when bcrypt cannot process the input.
func (h *BcryptPasswordHasher) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// Verify compares a bcrypt hash against a plain-text password.
// It returns nil on match and an error otherwise.
func (h *BcryptPasswordHasher) Verify(hashedPassword string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
