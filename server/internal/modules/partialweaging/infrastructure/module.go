package infrastructure

import (
	"net/http"
	"time"

	partialweagingapplication "server/internal/modules/partialweaging/application"
	"server/internal/modules/partialweaging/ports"
)

type Module struct {
	CreatePartialWeaging *partialweagingapplication.CreatePartialWeagingService
	ListPartialWeagings  *partialweagingapplication.ListPartialWeagingsService
	UpdatePartialWeaging *partialweagingapplication.UpdatePartialWeagingService
	DeletePartialWeaging *partialweagingapplication.DeletePartialWeagingService
	Handler              *PartialWeagingHandler
}

// NewModule assembles the partial weaging use cases and HTTP handler from its dependencies.
// It returns a module exposing the use cases and handler for wiring.
func NewModule(
	repository ports.PartialWeagingRepository,
	clock func() time.Time,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *Module {
	createPartialWeaging := partialweagingapplication.NewCreatePartialWeagingService(repository, clock)
	listPartialWeagings := partialweagingapplication.NewListPartialWeagingsService(repository)
	updatePartialWeaging := partialweagingapplication.NewUpdatePartialWeagingService(repository, clock)
	deletePartialWeaging := partialweagingapplication.NewDeletePartialWeagingService(repository, clock)

	return &Module{
		CreatePartialWeaging: createPartialWeaging,
		ListPartialWeagings:  listPartialWeagings,
		UpdatePartialWeaging: updatePartialWeaging,
		DeletePartialWeaging: deletePartialWeaging,
		Handler:              NewPartialWeagingHandler(createPartialWeaging, listPartialWeagings, deletePartialWeaging, updatePartialWeaging, authMiddleware, adminMiddleware),
	}
}
