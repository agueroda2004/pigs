package ports

type TokenGenerator interface {
	Generate() (string, error)
}
