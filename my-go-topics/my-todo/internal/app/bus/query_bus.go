// Package bus also implements the CQRS query bus.
//
// A Query is a request for data. Queries NEVER mutate state — they are
// read-only operations that return DTOs (read models), not domain aggregates.
//
// The naming convention distinguishes the two buses intentionally:
//   - CommandBus.Dispatch  (an imperative: "do this")
//   - QueryBus.Ask         (a question:   "what is this?")
//
// QueryHandler.Handle returns (any, error). The concrete type is determined
// by each query handler and must be type-asserted by the caller.
package bus

import (
	"context"
	"fmt"
	"reflect"
)

// Query is the marker interface for all queries.
type Query interface{}

// QueryHandler processes a single query type and returns a read model (DTO).
// The return type is `any` because different queries return different DTOs.
// The HTTP handler type-asserts the result to the expected DTO type.
type QueryHandler interface {
	Handle(ctx context.Context, query Query) (any, error)
}

// QueryBus routes queries to their registered handlers.
type QueryBus interface {
	Ask(ctx context.Context, query Query) (any, error)
}

// InMemoryQueryBus routes queries using a reflect.Type → handler map.
type InMemoryQueryBus struct {
	handlers map[reflect.Type]QueryHandler
}

func NewInMemoryQueryBus() *InMemoryQueryBus {
	return &InMemoryQueryBus{
		handlers: make(map[reflect.Type]QueryHandler),
	}
}

// Register associates a query type with a handler.
// Pass a zero-value query struct as the first argument (used only for its type).
func (b *InMemoryQueryBus) Register(query Query, handler QueryHandler) {
	b.handlers[reflect.TypeOf(query)] = handler
}

// Ask finds the handler for query, calls Handle, and returns the result DTO.
func (b *InMemoryQueryBus) Ask(ctx context.Context, query Query) (any, error) {
	handler, ok := b.handlers[reflect.TypeOf(query)]
	if !ok {
		return nil, fmt.Errorf("no handler registered for query %T", query)
	}

	return handler.Handle(ctx, query)
}
