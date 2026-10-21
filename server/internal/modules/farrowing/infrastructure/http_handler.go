package infrastructure

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	authdomain "server/internal/modules/auth/domain"
	farrowingapplication "server/internal/modules/farrowing/application"
	farrowingdomain "server/internal/modules/farrowing/domain"
	"server/internal/modules/farrowing/ports"
	platformhttp "server/internal/platform/http"
)

const dateLayout = "2006-01-02"

type CreateFarrowingUseCase interface {
	Execute(context.Context, farrowingapplication.CreateFarrowingCommand) (*farrowingdomain.Farrowing, error)
}

type ListFarrowingsUseCase interface {
	Execute(context.Context, ports.FarrowingFilter, int) (farrowingapplication.FarrowingPage, error)
}

type UpdateFarrowingUseCase interface {
	Execute(context.Context, uuid.UUID, farrowingapplication.UpdateFarrowingCommand) error
}

type DeleteFarrowingUseCase interface {
	Execute(context.Context, farrowingapplication.DeleteFarrowingCommand) error
}

type FarrowingHandler struct {
	createFarrowing CreateFarrowingUseCase
	listFarrowings  ListFarrowingsUseCase
	updateFarrowing UpdateFarrowingUseCase
	deleteFarrowing DeleteFarrowingUseCase
	authMiddleware  func(http.Handler) http.Handler
	adminMiddleware func(http.Handler) http.Handler
}

// NewFarrowingHandler wires the farrowing use cases and middlewares into a handler.
// It returns a handler ready to register its routes.
func NewFarrowingHandler(
	createFarrowing CreateFarrowingUseCase,
	listFarrowings ListFarrowingsUseCase,
	updateFarrowing UpdateFarrowingUseCase,
	deleteFarrowing DeleteFarrowingUseCase,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) *FarrowingHandler {
	return &FarrowingHandler{
		createFarrowing: createFarrowing,
		listFarrowings:  listFarrowings,
		updateFarrowing: updateFarrowing,
		deleteFarrowing: deleteFarrowing,
		authMiddleware:  authMiddleware,
		adminMiddleware: adminMiddleware,
	}
}

// RegisterRoutes registers the farrowing create, list, update and delete endpoints.
// Write routes are admin-only while the list route only requires authentication.
func (h *FarrowingHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/farrowings", h.adminMiddleware(http.HandlerFunc(h.create)))
	mux.Handle("GET /api/v1/farrowings", h.authMiddleware(http.HandlerFunc(h.list)))
	mux.Handle("PATCH /api/v1/farrowings/{id}", h.adminMiddleware(http.HandlerFunc(h.update)))
	mux.Handle("DELETE /api/v1/farrowings/{id}", h.adminMiddleware(http.HandlerFunc(h.delete)))
}

type farrowingOperatorRequest struct {
	OperatorID string `json:"operator_id"`
}

type farrowingMedicationRequest struct {
	MedicationID string  `json:"medication_id"`
	Dose         float64 `json:"dose"`
	AppliedBy    string  `json:"applied_by"`
}

type createFarrowingRequest struct {
	SowID           string                       `json:"sow_id"`
	FarrowDate      string                       `json:"farrow_date"`
	StartTime       *string                      `json:"start_time"`
	EndTime         *string                      `json:"end_time"`
	Location        *string                      `json:"location"`
	LiveBorn        *int                         `json:"live_born"`
	Stillborn       *int                         `json:"stillborn"`
	Mummified       *int                         `json:"mummified"`
	LitterWeight    *float64                     `json:"litter_weight"`
	StillbornWeight *float64                     `json:"stillborn_weight"`
	IsManipulated   bool                         `json:"is_manipulated"`
	Note            *string                      `json:"note"`
	Operators       []farrowingOperatorRequest   `json:"operators"`
	Medications     []farrowingMedicationRequest `json:"medications"`
}

type updateFarrowingRequest struct {
	FarrowDate           *string                       `json:"farrow_date"`
	StartTime            *string                       `json:"start_time"`
	EndTime              *string                       `json:"end_time"`
	Location             *string                       `json:"location"`
	LiveBorn             *int                          `json:"live_born"`
	Stillborn            *int                          `json:"stillborn"`
	Mummified            *int                          `json:"mummified"`
	LitterWeight         *float64                      `json:"litter_weight"`
	ClearLitterWeight    *bool                         `json:"clear_litter_weight"`
	StillbornWeight      *float64                      `json:"stillborn_weight"`
	ClearStillbornWeight *bool                         `json:"clear_stillborn_weight"`
	IsManipulated        *bool                         `json:"is_manipulated"`
	Note                 *string                       `json:"note"`
	Operators            *[]farrowingOperatorRequest   `json:"operators"`
	Medications          *[]farrowingMedicationRequest `json:"medications"`
}

type farrowingOperatorResponse struct {
	ID         uuid.UUID `json:"id"`
	OperatorID uuid.UUID `json:"operator_id"`
}

type farrowingMedicationResponse struct {
	ID           uuid.UUID `json:"id"`
	MedicationID uuid.UUID `json:"medication_id"`
	Dose         float64   `json:"dose"`
	AppliedBy    uuid.UUID `json:"applied_by"`
}

type farrowingResponse struct {
	ID              uuid.UUID                     `json:"id"`
	SowID           uuid.UUID                     `json:"sow_id"`
	SowCode         string                        `json:"sow_code"`
	ServiceID       uuid.UUID                     `json:"service_id"`
	FarrowDate      string                        `json:"farrow_date"`
	StartTime       *string                       `json:"start_time"`
	EndTime         *string                       `json:"end_time"`
	Location        *string                       `json:"location"`
	LiveBorn        int                           `json:"live_born"`
	Stillborn       int                           `json:"stillborn"`
	Mummified       int                           `json:"mummified"`
	CurrentPiglets  int                           `json:"current_piglets"`
	LitterWeight    *float64                      `json:"litter_weight"`
	StillbornWeight *float64                      `json:"stillborn_weight"`
	IsManipulated   bool                          `json:"is_manipulated"`
	IsNurse         bool                          `json:"is_nurse"`
	NurseStartDate  *string                       `json:"nurse_start_date"`
	Note            *string                       `json:"note"`
	Operators       []farrowingOperatorResponse   `json:"operators"`
	Medications     []farrowingMedicationResponse `json:"medications"`
}

// farrowingPageResponse is one page of farrowings together with its pagination metadata.
type farrowingPageResponse struct {
	Items      []farrowingResponse `json:"items"`
	Total      int                 `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	TotalPages int                 `json:"total_pages"`
}

// create handles POST /api/v1/farrowings and registers a farrowing.
// It parses the sow, date, times, required piglet counts, operators and medications,
// reads the actor from the context and lets the use case derive the service and
// apply the state changes.
func (h *FarrowingHandler) create(w http.ResponseWriter, r *http.Request) {
	var request createFarrowingRequest
	if err := platformhttp.DecodeJSON(w, r, &request); err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	sowID, err := uuid.Parse(request.SowID)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("La cerda no es válida"))
		return
	}

	farrowDate, err := time.Parse(dateLayout, request.FarrowDate)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("La fecha del parto no es válida"))
		return
	}

	liveBorn, stillborn, mummified, err := parsePigletCounts(request)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	operators, err := parseOperators(request.Operators)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	medications, err := parseMedications(request.Medications)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}

	if _, err := h.createFarrowing.Execute(r.Context(), farrowingapplication.CreateFarrowingCommand{
		SowID:           sowID,
		FarrowDate:      farrowDate,
		StartTime:       request.StartTime,
		EndTime:         request.EndTime,
		Location:        request.Location,
		LiveBorn:        liveBorn,
		Stillborn:       stillborn,
		Mummified:       mummified,
		LitterWeight:    request.LitterWeight,
		StillbornWeight: request.StillbornWeight,
		IsManipulated:   request.IsManipulated,
		Note:            request.Note,
		Operators:       operators,
		Medications:     medications,
		CreatedBy:       actor.UserID,
	}); err != nil {
		writeFarrowingError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// list handles GET /api/v1/farrowings and returns one page of farrowings matching the filter.
// It parses the optional sow_code, from and to query parameters plus the page parameter.
func (h *FarrowingHandler) list(w http.ResponseWriter, r *http.Request) {
	filter, err := parseFarrowingFilter(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	page, err := parseFarrowingPage(r)
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	farrowings, err := h.listFarrowings.Execute(r.Context(), filter, page)
	if err != nil {
		writeFarrowingError(w, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, toFarrowingPageResponse(farrowings))
}

// update handles PATCH /api/v1/farrowings/{id} and applies the provided fields.
// It parses the id path value, decodes the partial body, reads the actor and
// returns 200 without a body.
func (h *FarrowingHandler) update(w http.ResponseWriter, r *http.Request) {
	farrowingID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El identificador del parto no es válido"))
		return
	}

	var request updateFarrowingRequest
	if err := platformhttp.DecodeJSON(w, r, &request); err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, err)
		return
	}

	var farrowDate *time.Time
	if request.FarrowDate != nil {
		parsed, err := time.Parse(dateLayout, *request.FarrowDate)
		if err != nil {
			platformhttp.WriteError(w, http.StatusBadRequest, errors.New("La fecha del parto no es válida"))
			return
		}
		farrowDate = &parsed
	}

	command := farrowingapplication.UpdateFarrowingCommand{
		FarrowDate:      farrowDate,
		StartTime:       request.StartTime,
		EndTime:         request.EndTime,
		Location:        request.Location,
		LiveBorn:        request.LiveBorn,
		Stillborn:       request.Stillborn,
		Mummified:       request.Mummified,
		LitterWeight:    request.LitterWeight,
		StillbornWeight: request.StillbornWeight,
		IsManipulated:   request.IsManipulated,
		Note:            request.Note,
	}
	if request.ClearLitterWeight != nil {
		command.ClearLitterWeight = *request.ClearLitterWeight
	}
	if request.ClearStillbornWeight != nil {
		command.ClearStillbornWeight = *request.ClearStillbornWeight
	}

	if request.Operators != nil {
		operators, err := parseOperators(*request.Operators)
		if err != nil {
			platformhttp.WriteError(w, http.StatusBadRequest, err)
			return
		}
		command.Operators = operators
		command.UpdateOperators = true
	}
	if request.Medications != nil {
		medications, err := parseMedications(*request.Medications)
		if err != nil {
			platformhttp.WriteError(w, http.StatusBadRequest, err)
			return
		}
		command.Medications = medications
		command.UpdateMedications = true
	}

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}
	command.UpdatedBy = actor.UserID

	if err := h.updateFarrowing.Execute(r.Context(), farrowingID, command); err != nil {
		writeFarrowingError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// delete handles DELETE /api/v1/farrowings/{id} and removes a farrowing.
// It parses the id path value, reads the actor and returns 204 without a body.
func (h *FarrowingHandler) delete(w http.ResponseWriter, r *http.Request) {
	farrowingID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, errors.New("El identificador del parto no es válido"))
		return
	}

	actor, ok := authdomain.AuthenticatedUserFromContext(r.Context())
	if !ok {
		platformhttp.WriteError(w, http.StatusUnauthorized, errors.New("Usuario no autenticado"))
		return
	}

	if err := h.deleteFarrowing.Execute(r.Context(), farrowingapplication.DeleteFarrowingCommand{
		ID:        farrowingID,
		DeletedBy: actor.UserID,
	}); err != nil {
		writeFarrowingError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// parsePigletCounts reads the required live born, stillborn and mummified counts.
// It returns the matching domain error when any of them is missing or null.
func parsePigletCounts(request createFarrowingRequest) (int, int, int, error) {
	if request.LiveBorn == nil {
		return 0, 0, 0, farrowingdomain.ErrLiveBornRequired
	}
	if request.Stillborn == nil {
		return 0, 0, 0, farrowingdomain.ErrStillbornRequired
	}
	if request.Mummified == nil {
		return 0, 0, 0, farrowingdomain.ErrMummifiedRequired
	}
	return *request.LiveBorn, *request.Stillborn, *request.Mummified, nil
}

// parseOperators parses the operator identifiers of a create request.
// It returns an error when any operator id is not a valid UUID.
func parseOperators(requests []farrowingOperatorRequest) ([]farrowingapplication.CreateFarrowingOperatorInput, error) {
	operators := make([]farrowingapplication.CreateFarrowingOperatorInput, 0, len(requests))
	for _, request := range requests {
		operatorID, err := uuid.Parse(request.OperatorID)
		if err != nil {
			return nil, errors.New("El operador del parto no es válido")
		}
		operators = append(operators, farrowingapplication.CreateFarrowingOperatorInput{OperatorID: operatorID})
	}
	return operators, nil
}

// parseMedications parses the medications of a create request.
// It returns an error when any medication or applying operator id is not a valid UUID.
func parseMedications(requests []farrowingMedicationRequest) ([]farrowingapplication.CreateFarrowingMedicationInput, error) {
	medications := make([]farrowingapplication.CreateFarrowingMedicationInput, 0, len(requests))
	for _, request := range requests {
		medicationID, err := uuid.Parse(request.MedicationID)
		if err != nil {
			return nil, errors.New("El medicamento del parto no es válido")
		}
		appliedBy, err := uuid.Parse(request.AppliedBy)
		if err != nil {
			return nil, errors.New("El operador que aplica el medicamento no es válido")
		}
		medications = append(medications, farrowingapplication.CreateFarrowingMedicationInput{
			MedicationID: medicationID,
			Dose:         request.Dose,
			AppliedBy:    appliedBy,
		})
	}
	return medications, nil
}

// parseFarrowingFilter reads the optional list filters from the query string.
// It returns an error for a malformed from or to value and rejects a date range
// whose start is later than its end; sow_code narrows by the related sow code.
func parseFarrowingFilter(r *http.Request) (ports.FarrowingFilter, error) {
	query := r.URL.Query()
	var filter ports.FarrowingFilter

	if sowCode := strings.TrimSpace(query.Get("sow_code")); sowCode != "" {
		filter.SowCode = &sowCode
	}

	if from := strings.TrimSpace(query.Get("from")); from != "" {
		fromDate, err := time.Parse(dateLayout, from)
		if err != nil {
			return ports.FarrowingFilter{}, errors.New("La fecha inicial no es válida")
		}
		filter.FromDate = &fromDate
	}

	if to := strings.TrimSpace(query.Get("to")); to != "" {
		toDate, err := time.Parse(dateLayout, to)
		if err != nil {
			return ports.FarrowingFilter{}, errors.New("La fecha final no es válida")
		}
		filter.ToDate = &toDate
	}

	if filter.FromDate != nil && filter.ToDate != nil && filter.FromDate.After(*filter.ToDate) {
		return ports.FarrowingFilter{}, errors.New("La fecha inicial no puede ser posterior a la fecha final")
	}

	return filter, nil
}

// parseFarrowingPage reads the optional page query parameter.
// It defaults to page one and rejects non-positive or malformed values.
func parseFarrowingPage(r *http.Request) (int, error) {
	raw := strings.TrimSpace(r.URL.Query().Get("page"))
	if raw == "" {
		return 1, nil
	}
	page, err := strconv.Atoi(raw)
	if err != nil || page < 1 {
		return 0, errors.New("La página no es válida")
	}
	return page, nil
}

// toFarrowingResponse maps a domain farrowing to the HTTP response shape.
// It formats the date as YYYY-MM-DD and omits the audit fields.
func toFarrowingResponse(farrowing *farrowingdomain.Farrowing) farrowingResponse {
	return farrowingResponse{
		ID:              farrowing.ID,
		SowID:           farrowing.SowID,
		SowCode:         farrowing.SowCode,
		ServiceID:       farrowing.ServiceID,
		FarrowDate:      farrowing.FarrowDate.UTC().Format(dateLayout),
		StartTime:       farrowing.StartTime,
		EndTime:         farrowing.EndTime,
		Location:        farrowing.Location,
		LiveBorn:        farrowing.LiveBorn,
		Stillborn:       farrowing.Stillborn,
		Mummified:       farrowing.Mummified,
		CurrentPiglets:  farrowing.CurrentPiglets,
		LitterWeight:    farrowing.LitterWeight,
		StillbornWeight: farrowing.StillbornWeight,
		IsManipulated:   farrowing.IsManipulated,
		IsNurse:         farrowing.IsNurse,
		NurseStartDate:  toDatePointer(farrowing.NurseStartDate),
		Note:            farrowing.Note,
		Operators:       toOperatorResponses(farrowing.Operators),
		Medications:     toMedicationResponses(farrowing.Medications),
	}
}

// toFarrowingPageResponse maps one page of domain farrowings to the paginated shape.
// It returns an empty items slice instead of null when there are no farrowings.
func toFarrowingPageResponse(page farrowingapplication.FarrowingPage) farrowingPageResponse {
	return farrowingPageResponse{
		Items:      toFarrowingResponses(page.Items),
		Total:      page.Total,
		Page:       page.Page,
		PageSize:   page.PageSize,
		TotalPages: page.TotalPages,
	}
}

// toDatePointer formats an optional date as YYYY-MM-DD in UTC.
// It returns nil when the given date is nil so the response stays nullable.
func toDatePointer(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.UTC().Format(dateLayout)
	return &formatted
}

// toFarrowingResponses maps a list of domain farrowings to HTTP response shapes.
// It returns an empty slice instead of null when there are no farrowings.
func toFarrowingResponses(farrowings []*farrowingdomain.Farrowing) []farrowingResponse {
	responses := make([]farrowingResponse, 0, len(farrowings))
	for _, farrowing := range farrowings {
		responses = append(responses, toFarrowingResponse(farrowing))
	}
	return responses
}

// toOperatorResponses maps the operator links of a farrowing to the HTTP shape.
// It returns an empty slice instead of null when there are no operators.
func toOperatorResponses(operators []*farrowingdomain.FarrowingOperator) []farrowingOperatorResponse {
	responses := make([]farrowingOperatorResponse, 0, len(operators))
	for _, operator := range operators {
		responses = append(responses, farrowingOperatorResponse{
			ID:         operator.ID,
			OperatorID: operator.OperatorID,
		})
	}
	return responses
}

// toMedicationResponses maps the medication links of a farrowing to the HTTP shape.
// It returns an empty slice instead of null when there are no medications.
func toMedicationResponses(medications []*farrowingdomain.FarrowingMedication) []farrowingMedicationResponse {
	responses := make([]farrowingMedicationResponse, 0, len(medications))
	for _, medication := range medications {
		responses = append(responses, farrowingMedicationResponse{
			ID:           medication.ID,
			MedicationID: medication.MedicationID,
			Dose:         medication.Dose,
			AppliedBy:    medication.AppliedBy,
		})
	}
	return responses
}

// writeFarrowingError maps farrowing domain, application and port errors to status codes.
// It falls back to 500 for unrecognized errors.
func writeFarrowingError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, ports.ErrFarrowingNotFound),
		errors.Is(err, ports.ErrSowNotFound),
		errors.Is(err, ports.ErrServiceNotFound),
		errors.Is(err, ports.ErrOperatorNotFound),
		errors.Is(err, ports.ErrMedicationNotFound):
		status = http.StatusNotFound
	case errors.Is(err, farrowingapplication.ErrSowNotGestating),
		errors.Is(err, farrowingapplication.ErrServiceNotConfirmed),
		errors.Is(err, ports.ErrFarrowingAlreadyExists),
		errors.Is(err, ports.ErrFarrowingHasRelatedRecords):
		status = http.StatusConflict
	case errors.Is(err, farrowingdomain.ErrInvalidID),
		errors.Is(err, farrowingdomain.ErrInvalidService),
		errors.Is(err, farrowingdomain.ErrInvalidSow),
		errors.Is(err, farrowingdomain.ErrInvalidFarrowDate),
		errors.Is(err, farrowingdomain.ErrInvalidLastMount),
		errors.Is(err, farrowingdomain.ErrInvalidStartTime),
		errors.Is(err, farrowingdomain.ErrInvalidEndTime),
		errors.Is(err, farrowingdomain.ErrInvalidLocation),
		errors.Is(err, farrowingdomain.ErrInvalidLiveBorn),
		errors.Is(err, farrowingdomain.ErrInvalidStillborn),
		errors.Is(err, farrowingdomain.ErrInvalidMummified),
		errors.Is(err, farrowingdomain.ErrLiveBornTooHigh),
		errors.Is(err, farrowingdomain.ErrStillbornTooHigh),
		errors.Is(err, farrowingdomain.ErrMummifiedTooHigh),
		errors.Is(err, farrowingdomain.ErrLiveBornRequired),
		errors.Is(err, farrowingdomain.ErrStillbornRequired),
		errors.Is(err, farrowingdomain.ErrMummifiedRequired),
		errors.Is(err, farrowingdomain.ErrLitterWeightRequired),
		errors.Is(err, farrowingdomain.ErrStillbornWeightRequired),
		errors.Is(err, farrowingdomain.ErrLitterWeightTooHigh),
		errors.Is(err, farrowingdomain.ErrStillbornWeightTooHigh),
		errors.Is(err, farrowingdomain.ErrMedicationDoseTooHigh),
		errors.Is(err, farrowingdomain.ErrInvalidLitterWeight),
		errors.Is(err, farrowingdomain.ErrInvalidStillbornWeight),
		errors.Is(err, farrowingdomain.ErrInvalidNote),
		errors.Is(err, farrowingdomain.ErrInvalidNurseStartDate),
		errors.Is(err, farrowingdomain.ErrInvalidCreatedBy),
		errors.Is(err, farrowingdomain.ErrInvalidUpdatedBy),
		errors.Is(err, farrowingdomain.ErrFarrowDateInFuture),
		errors.Is(err, farrowingdomain.ErrFarrowDateBeforeMount),
		errors.Is(err, farrowingdomain.ErrInvalidUpdate),
		errors.Is(err, farrowingdomain.ErrFarrowDateAfterRelatedRecords),
		errors.Is(err, farrowingdomain.ErrLiveBornLocked),
		errors.Is(err, farrowingdomain.ErrDuplicateOperator),
		errors.Is(err, farrowingdomain.ErrDuplicateMedication),
		errors.Is(err, farrowingdomain.ErrInvalidFarrowingOperatorID),
		errors.Is(err, farrowingdomain.ErrInvalidFarrowingOperatorFarrowing),
		errors.Is(err, farrowingdomain.ErrInvalidFarrowingOperatorOperator),
		errors.Is(err, farrowingdomain.ErrInvalidFarrowingMedicationID),
		errors.Is(err, farrowingdomain.ErrInvalidFarrowingMedicationFarrowing),
		errors.Is(err, farrowingdomain.ErrInvalidFarrowingMedicationMedication),
		errors.Is(err, farrowingdomain.ErrInvalidFarrowingMedicationDose),
		errors.Is(err, farrowingdomain.ErrInvalidFarrowingMedicationAppliedBy):
		status = http.StatusBadRequest
	}
	platformhttp.WriteError(w, status, err)
}
