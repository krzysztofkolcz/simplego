package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/krzysztofkolcz/mymigrations/internal/application/command"
	"github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db"
	commanddbpub "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/command/public"
	commanddbtenant "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/command/tenant"
	infraevent "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/event"
	tenantrepo "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/repository/tenant"
	"github.com/krzysztofkolcz/mymigrations/internal/infrastructure/worker"
)

// truncateOutbox usuwa wszystkie wiersze z outbox_events danego tenanta.
func truncateOutbox(ctx context.Context, t *testing.T) {
	t.Helper()
	_, err := ConnectionPool.Exec(ctx,
		fmt.Sprintf(`TRUNCATE TABLE "%s".outbox_events RESTART IDENTITY`, TenantSchema),
	)
	require.NoError(t, err)
}

// unpublishedEvents zwraca niepublikowane eventy z tabeli outbox_events tenanta.
func unpublishedEvents(ctx context.Context, t *testing.T) []commanddbtenant.SelectUnpublishedOutboxEventsRow {
	t.Helper()
	var events []commanddbtenant.SelectUnpublishedOutboxEventsRow
	txManager := db.NewTxManager(ConnectionPool)
	err := txManager.WithinTransaction(ctx, TenantSchema, func(q *commanddbtenant.Queries) error {
		var err error
		events, err = q.SelectUnpublishedOutboxEvents(ctx)
		return err
	})
	require.NoError(t, err)
	return events
}

// allOutboxEvents zwraca wszystkie wiersze z outbox_events (w tym już opublikowane).
func allOutboxEvents(ctx context.Context, t *testing.T) []struct {
	ID          int64
	EventName   string
	PublishedAt *time.Time
} {
	t.Helper()
	rows, err := ConnectionPool.Query(ctx,
		fmt.Sprintf(
			`SELECT id, event_name, published_at FROM "%s".outbox_events ORDER BY id`,
			TenantSchema,
		),
	)
	require.NoError(t, err)
	defer rows.Close()

	var result []struct {
		ID          int64
		EventName   string
		PublishedAt *time.Time
	}
	for rows.Next() {
		var r struct {
			ID          int64
			EventName   string
			PublishedAt *time.Time
		}
		require.NoError(t, rows.Scan(&r.ID, &r.EventName, &r.PublishedAt))
		result = append(result, r)
	}
	require.NoError(t, rows.Err())
	return result
}

// payloadField dekoduje payload JSONB i zwraca wartość pola.
func payloadField(t *testing.T, payload []byte, field string) string {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal(payload, &m))
	v, ok := m[field]
	require.True(t, ok, "field %q not found in payload %s", field, payload)
	return v.(string)
}

// ── testy OutboxPublisher ─────────────────────────────────────────────────────

func TestOutboxPublisher_CreateTodo_WritesEventToOutbox(t *testing.T) {
	ctx := context.Background()
	truncateOutbox(ctx, t)

	uow := tenantrepo.NewTodoUnitOfWork(db.NewTxManager(ConnectionPool), TenantSchema)
	id, err := command.NewCreateTodoHandler(uow, infraevent.NewOutboxPublisher()).
		Handle(ctx, command.CreateTodoCommand{Title: "outbox create test"})
	require.NoError(t, err)

	events := unpublishedEvents(ctx, t)
	require.Len(t, events, 1)
	assert.Equal(t, "todo.created", events[0].EventName)
	assert.Equal(t, id.String(), payloadField(t, events[0].Payload, "TodoID"))
	assert.Equal(t, "outbox create test", payloadField(t, events[0].Payload, "Title"))
}

func TestOutboxPublisher_CompleteTodo_WritesEventToOutbox(t *testing.T) {
	ctx := context.Background()

	uow := tenantrepo.NewTodoUnitOfWork(db.NewTxManager(ConnectionPool), TenantSchema)
	id, err := command.NewCreateTodoHandler(uow, infraevent.NewLogPublisher()).
		Handle(ctx, command.CreateTodoCommand{Title: "outbox complete test"})
	require.NoError(t, err)

	truncateOutbox(ctx, t) // wyczyść event z Create

	err = command.NewCompleteTodoHandler(uow, infraevent.NewOutboxPublisher()).
		Handle(ctx, command.CompleteTodoCommand{ID: id})
	require.NoError(t, err)

	events := unpublishedEvents(ctx, t)
	require.Len(t, events, 1)
	assert.Equal(t, "todo.completed", events[0].EventName)
	assert.Equal(t, id.String(), payloadField(t, events[0].Payload, "TodoID"))
}

func TestOutboxPublisher_DeleteTodo_WritesEventToOutbox(t *testing.T) {
	ctx := context.Background()

	uow := tenantrepo.NewTodoUnitOfWork(db.NewTxManager(ConnectionPool), TenantSchema)
	id, err := command.NewCreateTodoHandler(uow, infraevent.NewLogPublisher()).
		Handle(ctx, command.CreateTodoCommand{Title: "outbox delete test"})
	require.NoError(t, err)

	truncateOutbox(ctx, t)

	err = command.NewDeleteTodoHandler(uow, infraevent.NewOutboxPublisher()).
		Handle(ctx, command.DeleteTodoCommand{ID: id})
	require.NoError(t, err)

	events := unpublishedEvents(ctx, t)
	require.Len(t, events, 1)
	assert.Equal(t, "todo.deleted", events[0].EventName)
	assert.Equal(t, id.String(), payloadField(t, events[0].Payload, "TodoID"))
}

func TestOutboxPublisher_LogPublisher_DoesNotWriteToOutbox(t *testing.T) {
	ctx := context.Background()
	truncateOutbox(ctx, t)

	uow := tenantrepo.NewTodoUnitOfWork(db.NewTxManager(ConnectionPool), TenantSchema)
	_, err := command.NewCreateTodoHandler(uow, infraevent.NewLogPublisher()).
		Handle(ctx, command.CreateTodoCommand{Title: "log publisher test"})
	require.NoError(t, err)

	events := unpublishedEvents(ctx, t)
	assert.Empty(t, events)
}

// ── testy OutboxWorker ────────────────────────────────────────────────────────

func TestOutboxWorker_ProcessOnce_MarksEventsPublished(t *testing.T) {
	ctx := context.Background()
	truncateOutbox(ctx, t)

	// Zapisz event do outbox przez OutboxPublisher
	uow := tenantrepo.NewTodoUnitOfWork(db.NewTxManager(ConnectionPool), TenantSchema)
	_, err := command.NewCreateTodoHandler(uow, infraevent.NewOutboxPublisher()).
		Handle(ctx, command.CreateTodoCommand{Title: "worker test"})
	require.NoError(t, err)

	// Przed processOnce: event jest niepublikowany
	before := unpublishedEvents(ctx, t)
	require.Len(t, before, 1)

	// Uruchom worker
	txManager := db.NewTxManager(ConnectionPool)
	publicQ := commanddbpub.New(ConnectionPool)
	w := worker.NewOutboxWorker(txManager, publicQ, time.Second)
	require.NoError(t, w.ProcessOnce(ctx))

	// Po processOnce: brak niepublikowanych eventów
	after := unpublishedEvents(ctx, t)
	assert.Empty(t, after)

	// Sprawdź że event ma published_at ustawiony
	all := allOutboxEvents(ctx, t)
	require.Len(t, all, 1)
	assert.Equal(t, "todo.created", all[0].EventName)
	assert.NotNil(t, all[0].PublishedAt)
}
