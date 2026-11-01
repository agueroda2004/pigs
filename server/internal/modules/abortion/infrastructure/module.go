package infrastructure

import (
	"net/http"
	"time"

	abortionapplication "server/internal/modules/abortion/application"
	"server/internal/modules/abortion/ports"
)

type Module struct {
	CreateAbortion    *abortionapplication.CreateAbortionService
	ListAbortions     *abortionapplication.ListAbortionsService
	UpdateAbortion    *abortionapplication.UpdateAbortionService
	DeleteAbortion    *abortionapplication.DeleteAbortionService
	AbortHistoryBySow *abortionapplication.AbortHistoryBySowService
	Handler           *AbortionHandler
}

// NewModule assembles the abortion use cases and HTTP handler from its dependencies.
// It returns a module exposing the use cases and handler for wiring.
func NewModule(
	repository ports.AbortionRepository,
	clock func() time.Time,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *Module {
	createAbortion := abortionapplication.NewCreateAbortionService(repository, clock)
	listAbortions := abortionapplication.NewListAbortionsService(repository)
	updateAbortion := abortionapplication.NewUpdateAbortionService(repository, clock)
	deleteAbortion := abortionapplication.NewDeleteAbortionService(repository, clock)
	abortHistoryBySow := abortionapplication.NewAbortHistoryBySowService(repository)

	return &Module{
		CreateAbortion:    createAbortion,
		ListAbortions:     listAbortions,
		UpdateAbortion:    updateAbortion,
		DeleteAbortion:    deleteAbortion,
		AbortHistoryBySow: abortHistoryBySow,
		Handler:           NewAbortionHandler(createAbortion, listAbortions, updateAbortion, deleteAbortion, authMiddleware, adminMiddleware),
	}
}
