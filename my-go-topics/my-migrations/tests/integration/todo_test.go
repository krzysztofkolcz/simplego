package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)


func TestFirst(t *testing.T){

	assert.True(t, true)
}

// func TestCreateTodo(t *testing.T) {
// 	ctx := context.Background()

// 	tenantID := "test123"

// 	// tworzymy tenant + migracje
// 	err := db.CreateTenant(ctx, testDB, testDSN, tenantID)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	schema := "tenant_" + tenantID

// 	q, cleanup, err := database.WithTenant(ctx, testDB, schema)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer cleanup()

// 	// test
// 	err = q.CreateTodo(ctx, db.CreateTodoParams{
// 		ID:    uuid.New(),
// 		Title: "test",
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// }