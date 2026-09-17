package ports

type TokenHasher interface {
	Hash(token string) (string, error)
}
