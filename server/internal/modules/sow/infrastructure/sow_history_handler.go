package infrastructure

import (
	"context"
	"errors"
	"net/http"
	"sort"

	abortiondomain "server/internal/modules/abortion/domain"
	farrowingdomain "server/internal/modules/farrowing/domain"
	servicedomain "server/internal/modules/service/domain"
	sowapplication "server/internal/modules/sow/application"
	sowdomain "server/internal/modules/sow/domain"
	"server/internal/modules/sow/ports"
	platformhttp "server/internal/platform/http"
)

type SowHistoryBySowUseCase interface {
	Execute(context.Context, string) (*sowdomain.SowHistoryBySow, error)
}

type ServicesHistoryBySowUseCase interface {
	Execute(context.Context, string) ([]*servicedomain.ServiceHistoryBySow, error)
}

type AbortionHistoryBySowUseCase interface {
	Execute(context.Context, string) ([]abortiondomain.AbortionHistoryBySow, error)
}

type FarrowingHistoryBySowUseCase interface {
	Execute(context.Context, string) ([]*farrowingdomain.FarrowingHistoryBySow, error)
}

// SowHistoryHandler serves the authenticated sow-history read endpoint.
type SowHistoryHandler struct {
	getSowHistory       SowHistoryBySowUseCase
	getServicesHistory  ServicesHistoryBySowUseCase
	getAbortionHistory  AbortionHistoryBySowUseCase
	getFarrowingHistory FarrowingHistoryBySowUseCase
	authMiddleware      func(http.Handler) http.Handler
}

// NewSowHistoryHandler wires the use cases and authentication middleware.
func NewSowHistoryHandler(
	getSowHistory SowHistoryBySowUseCase,
	getServicesHistory ServicesHistoryBySowUseCase,
	getAbortionHistory AbortionHistoryBySowUseCase,
	getFarrowingHistory FarrowingHistoryBySowUseCase,
	authMiddleware func(http.Handler) http.Handler,
) *SowHistoryHandler {
	return &SowHistoryHandler{
		getSowHistory:       getSowHistory,
		getServicesHistory:  getServicesHistory,
		getAbortionHistory:  getAbortionHistory,
		getFarrowingHistory: getFarrowingHistory,
		authMiddleware:      authMiddleware,
	}
}

// RegisterRoutes registers the authenticated sow-history route.
func (h *SowHistoryHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/v1/sows/{code}/history", h.authMiddleware(http.HandlerFunc(h.get)))
}

type sowHistoryBySowResponse struct {
	Code      string  `json:"code"`
	EntryDate string  `json:"entry_date"`
	BirthDate *string `json:"birth_date"`
	State     string  `json:"state"`
	Parity    int     `json:"parity"`
	BreedName string  `json:"breed_name"`
	Age       string  `json:"age"`
}

type serviceHistoryMountBySowResponse struct {
	BoarCode     string `json:"boar_code"`
	OperatorName string `json:"operator_name"`
	MountDate    string `json:"mount_date"`
	Type         string `json:"type"`
}

type serviceHistoryBySowResponse struct {
	SowCode  string                             `json:"sow_code"`
	Note     *string                            `json:"note"`
	State    string                             `json:"state"`
	Location *string                            `json:"location"`
	Mounts   []serviceHistoryMountBySowResponse `json:"mounts"`
}

type abortionHistoryBySowResponse struct {
	SowCode      string  `json:"sow_code"`
	AbortionDate string  `json:"abortion_date"`
	Cause        string  `json:"cause"`
	Note         *string `json:"note"`
}

type farrowingHistoryOperatorResponse struct {
	OperatorName string `json:"operator_name"`
}

type farrowingHistoryMedicationResponse struct {
	MedicationName string  `json:"medication_name"`
	Dose           float64 `json:"dose"`
	AppliedByName  string  `json:"applied_by_name"`
}

type farrowingHistoryBySowResponse struct {
	FarrowDate      string                               `json:"farrow_date"`
	StartTime       *string                              `json:"start_time"`
	EndTime         *string                              `json:"end_time"`
	Duration        *string                              `json:"duration"`
	Location        *string                              `json:"location"`
	LiveBorn        int                                  `json:"live_born"`
	Mummified       int                                  `json:"mummified"`
	Stillborn       int                                  `json:"stillborn"`
	LitterWeight    *float64                             `json:"litter_weight"`
	StillbornWeight *float64                             `json:"stillborn_weight"`
	IsManipulated   bool                                 `json:"is_manipulated"`
	Note            *string                              `json:"note"`
	IsNurse         bool                                 `json:"is_nurse"`
	Operators       []farrowingHistoryOperatorResponse   `json:"operators"`
	Medications     []farrowingHistoryMedicationResponse `json:"medications"`
}

type sowHistoryEventResponse struct {
	Type string `json:"type"`
	Date string `json:"date"`
	Data any    `json:"data"`
}

type sowHistoryResponse struct {
	Sow    sowHistoryBySowResponse   `json:"sow"`
	Events []sowHistoryEventResponse `json:"events"`
}

// get returns the unified sow history timeline for the requested code.
// It loads the sow, services, abortions and farrowings, then merges them into
// newest-first events while preserving the existing not-found behavior.
func (h *SowHistoryHandler) get(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	sow, err := h.getSowHistory.Execute(r.Context(), code)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, ports.ErrSowNotFound) || errors.Is(err, sowapplication.ErrSowInactive) {
			status = http.StatusNotFound
		}
		platformhttp.WriteError(w, status, err)
		return
	}

	services, err := h.getServicesHistory.Execute(r.Context(), code)
	if err != nil {
		platformhttp.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	abortions, err := h.getAbortionHistory.Execute(r.Context(), code)
	if err != nil {
		platformhttp.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	farrowings, err := h.getFarrowingHistory.Execute(r.Context(), code)
	if err != nil {
		platformhttp.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	events := make([]sowHistoryEventResponse, 0, 1+len(services)+len(abortions)+len(farrowings))
	events = append(events, sowHistoryEventResponse{
		Type: "sow_entry",
		Date: sow.EntryDate,
		Data: toSowHistoryBySowResponse(sow),
	})
	for _, service := range services {
		events = append(events, sowHistoryEventResponse{
			Type: "service",
			Date: serviceEventDate(service),
			Data: toServiceHistoryResponse(service),
		})
	}
	for _, abortion := range abortions {
		events = append(events, sowHistoryEventResponse{
			Type: "abortion",
			Date: abortion.AbortionDate,
			Data: abortionHistoryBySowResponse{
				SowCode:      abortion.SowCode,
				AbortionDate: abortion.AbortionDate,
				Cause:        abortion.Cause,
				Note:         abortion.Note,
			},
		})
	}
	for _, farrowing := range farrowings {
		events = append(events, sowHistoryEventResponse{
			Type: "farrowing",
			Date: farrowing.FarrowDate,
			Data: toFarrowingHistoryResponse(farrowing),
		})
	}

	sort.SliceStable(events, func(i, j int) bool {
		return events[i].Date > events[j].Date
	})

	platformhttp.WriteJSON(w, http.StatusOK, sowHistoryResponse{
		Sow:    toSowHistoryBySowResponse(sow),
		Events: events,
	})
}

// toSowHistoryBySowResponse maps the sow read model to the HTTP response shape.
// It keeps nullable dates as null and formats the precomputed age unchanged.
func toSowHistoryBySowResponse(sow *sowdomain.SowHistoryBySow) sowHistoryBySowResponse {
	return sowHistoryBySowResponse{
		Code:      sow.Code,
		EntryDate: sow.EntryDate,
		BirthDate: sow.BirthDate,
		State:     string(sow.State),
		Parity:    sow.Parity,
		BreedName: sow.BreedName,
		Age:       sow.Age,
	}
}

// toServiceHistoryResponse maps a service history item and its mounts to the HTTP shape.
// It returns an empty mounts slice instead of null when the service has no mounts.
func toServiceHistoryResponse(service *servicedomain.ServiceHistoryBySow) serviceHistoryBySowResponse {
	mounts := make([]serviceHistoryMountBySowResponse, 0, len(service.Mounts))
	for _, mount := range service.Mounts {
		mounts = append(mounts, serviceHistoryMountBySowResponse{
			BoarCode:     mount.BoarCode,
			OperatorName: mount.OperatorName,
			MountDate:    mount.MountDate,
			Type:         string(mount.Type),
		})
	}
	return serviceHistoryBySowResponse{
		SowCode:  service.SowCode,
		Note:     service.Note,
		State:    string(service.State),
		Location: service.Location,
		Mounts:   mounts,
	}
}

// serviceEventDate returns the earliest mount date of a service, used as its event date.
// It returns an empty string when the service has no mount with a date.
func serviceEventDate(service *servicedomain.ServiceHistoryBySow) string {
	date := ""
	for _, mount := range service.Mounts {
		if mount.MountDate == "" {
			continue
		}
		if date == "" || mount.MountDate < date {
			date = mount.MountDate
		}
	}
	return date
}

// toFarrowingHistoryResponse maps a farrowing history item to the HTTP shape.
// It returns empty operator and medication slices instead of null when there are none.
func toFarrowingHistoryResponse(farrowing *farrowingdomain.FarrowingHistoryBySow) farrowingHistoryBySowResponse {
	operators := make([]farrowingHistoryOperatorResponse, 0, len(farrowing.Operators))
	for _, operator := range farrowing.Operators {
		operators = append(operators, farrowingHistoryOperatorResponse{OperatorName: operator.OperatorName})
	}
	medications := make([]farrowingHistoryMedicationResponse, 0, len(farrowing.Medications))
	for _, medication := range farrowing.Medications {
		medications = append(medications, farrowingHistoryMedicationResponse{
			MedicationName: medication.MedicationName,
			Dose:           medication.Dose,
			AppliedByName:  medication.AppliedByName,
		})
	}
	return farrowingHistoryBySowResponse{
		FarrowDate:      farrowing.FarrowDate,
		StartTime:       farrowing.StartTime,
		EndTime:         farrowing.EndTime,
		Duration:        farrowing.Duration,
		Location:        farrowing.Location,
		LiveBorn:        farrowing.LiveBorn,
		Mummified:       farrowing.Mummified,
		Stillborn:       farrowing.Stillborn,
		LitterWeight:    farrowing.LitterWeight,
		StillbornWeight: farrowing.StillbornWeight,
		IsManipulated:   farrowing.IsManipulated,
		Note:            farrowing.Note,
		IsNurse:         farrowing.IsNurse,
		Operators:       operators,
		Medications:     medications,
	}
}
