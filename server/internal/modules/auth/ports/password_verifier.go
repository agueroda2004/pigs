package ports

type PasswordVerifier interface {
	Verify(hashedPassword string, password string) error
}
