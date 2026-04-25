package apierrors

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"slices"

	todoapi "github.com/C5383717/my-todo/internal/api/todo"
	logger "github.com/C5383717/my-todo/internal/log"
	"github.com/C5383717/my-todo/internal/ptr"
	slogctx "github.com/veqryn/slog-context"
)

// APIErrorMapper is the central error-to-HTTP mapping registry.
// It contains two lists:
//   - APIErrors:      checked by best-match (most specific mapping wins)
//   - PriorityErrors: checked first regardless of specificity
type APIErrorMapper struct {
	APIErrors      []APIErrors
	PriorityErrors []APIErrors
}

// APIErrors pairs one or more sentinel errors with the HTTP response to return
// when any of them appear in the error chain.
type APIErrors struct {
	Errors        []error
	ExposedError  todoapi.DetailedError
	ContextGetter func(error) map[string]any
}

// apiErrorMapper is the singleton registry wired at package init.
var apiErrorMapper = APIErrorMapper{
	APIErrors: slices.Concat(
		todoMapper,  // domain-specific errors (highest priority in normal list)
		defaultMapper,
	),
	PriorityErrors: highPrio,
}

// TransformToAPIError is the single entry point for error → HTTP response mapping.
// Called by ResponseErrorHandlerFunc whenever a handler returns an error.
//
// Algorithm:
//  1. If the error matches a priority mapping → return it immediately
//  2. Walk all APIErrors, count how many sentinel errors appear in the chain
//  3. Return the mapping with the highest match count
//  4. If nothing matches → log and return 500
func TransformToAPIError(ctx context.Context, err error) *todoapi.ErrorMessage {
	e := apiErrorMapper.transform(ctx, err)
	if e == nil {
		logger.Info(
			ctx, "No appropriate error mapping. Defaulting to generic 500",
			slog.String(slogctx.ErrKey, err.Error()),
		)

		e = ptr.PointTo(InternalServerErrorMessage())
	}

	return e
}

func (m *APIErrorMapper) transform(ctx context.Context, err error) *todoapi.ErrorMessage {
	e, ok := m.containsAsPriority(err)
	if ok {
		return e
	}

	result := m.getBestMatches(err)

	debugMappingCandidates(ctx, err, result)

	if len(result) == 0 {
		return nil
	}

	selected := result[0]

	detail := selected.ExposedError
	if selected.ContextGetter != nil {
		detail.Context = ptr.PointTo(selected.ContextGetter(err))
	}

	return &todoapi.ErrorMessage{Error: detail}
}

func (m *APIErrorMapper) containsAsPriority(err error) (*todoapi.ErrorMessage, bool) {
	for _, priorityErrors := range m.PriorityErrors {
		if countMatchingErrors(err, priorityErrors.Errors) > 0 {
			return &todoapi.ErrorMessage{Error: priorityErrors.ExposedError}, true
		}
	}

	return nil, false
}

func (m *APIErrorMapper) getBestMatches(err error) []APIErrors {
	minCount := 1

	var result []APIErrors

	for _, apiErr := range m.APIErrors {
		count := countMatchingErrors(err, apiErr.Errors)

		if len(apiErr.Errors) > count {
			continue
		}

		if count == minCount {
			result = append(result, apiErr)
		} else if count > minCount {
			minCount = count
			result = []APIErrors{apiErr}
		}
	}

	return result
}

func countMatchingErrors(err error, candidates []error) int {
	matchCount := 0

	for _, candidateErr := range candidates {
		if errors.Is(err, candidateErr) {
			matchCount++
		}
	}

	return matchCount
}

func debugMappingCandidates(ctx context.Context, err error, mappingCandidates []APIErrors) {
	if len(mappingCandidates) > 1 {
		logger.Debug(
			ctx, "Mapping more than one error; selecting candidates",
			slog.String(slogctx.ErrKey, err.Error()),
		)

		for position, me := range mappingCandidates {
			logger.Debug(
				ctx, "Matched candidate",
				slog.Int("position", position),
				slog.String("code", me.ExposedError.Code),
				slog.String("status", http.StatusText(me.ExposedError.Status)),
				slog.Int("matchedLength", len(me.Errors)),
			)
		}
	}
}
