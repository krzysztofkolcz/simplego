package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/krzysztofkolcz/mymigrations/internal/application/command"
	"github.com/krzysztofkolcz/mymigrations/internal/application/query"
	"github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db"
	infraevent "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/event"
	tenantrepo "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/repository/tenant"
)

func TestListTodosHandler_Handle(t *testing.T) {
	ctx := context.Background()
	txManager := db.NewTxManager(ConnectionPool)
	uow := tenantrepo.NewTodoUnitOfWork(txManager, TenantSchema)

	id1, err := command.NewCreateTodoHandler(uow, infraevent.NewLogPublisher()).Handle(ctx, command.CreateTodoCommand{Title: "list test todo 1"})
	require.NoError(t, err)

	id2, err := command.NewCreateTodoHandler(uow, infraevent.NewLogPublisher()).Handle(ctx, command.CreateTodoCommand{Title: "list test todo 2"})
	require.NoError(t, err)

	repo := tenantrepo.NewTodoReadRepository(txManager, TenantSchema)
	list, err := query.NewListTodosHandler(repo).Handle(ctx, query.ListTodosQuery{})
	require.NoError(t, err)

	ids := make(map[interface{}]string, len(list))
	for _, item := range list {
		ids[item.ID] = item.Title
	}

	assert.Equal(t, "list test todo 1", ids[id1])
	assert.Equal(t, "list test todo 2", ids[id2])
}
