package infrastructure

import (
	"net/http"
	"time"

	weagingapplication "server/internal/modules/weaging/application"
	"server/internal/modules/weaging/ports"
)

type Module struct {
	CreateWeaging *weagingapplication.CreateWeagingService
	ListWeagings  *weagingapplication.ListWeagingsService
	UpdateWeaging *weagingapplication.UpdateWeagingService
	DeleteWeaging *weagingapplication.DeleteWeagingService
	Handler       *WeagingHandler
}

// NewModule assembles the weaging use cases and HTTP handler from its dependencies.
// It returns a module exposing the use cases and handler for wiring.
func NewModule(
	repository ports.WeagingRepository,
	clock func() time.Time,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *Module {
	createWeaging := weagingapplication.NewCreateWeagingService(repository, clock)
	listWeagings := weagingapplication.NewListWeagingsService(repository)
	updateWeaging := weagingapplication.NewUpdateWeagingService(repository, clock)
	deleteWeaging := weagingapplication.NewDeleteWeagingService(repository, clock)

	return &Module{
		CreateWeaging: createWeaging,
		ListWeagings:  listWeagings,
		UpdateWeaging: updateWeaging,
		DeleteWeaging: deleteWeaging,
		Handler: NewWeagingHandler(
			createWeaging,
			listWeagings,
			updateWeaging,
			deleteWeaging,
			authMiddleware,
			adminMiddleware,
		),
	}
}
