package tests

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/krzysztofkolcz/my-sqlc/internal/db"
)

// func TestCreateUser(t *testing.T) {
// 	cleanDB(t)

// 	q := db.New(testDB)

// 	ctx := context.Background()

// 	id := uuid.New()

// 	err := q.CreateUser(ctx, db.CreateUserParams{
// 		ID:       id,
// 		Email:    "test@example.com",
// 		Password: "secret",
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	user, err := q.GetUserByEmail(ctx, "test@example.com")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if user.ID != id {
// 		t.Fatalf("expected %v, got %v", id, user.ID)
// 	}
// }

func withTx(t *testing.T) (*db.Queries, func()) {
	t.Helper()

	ctx := context.Background()

	tx, err := testDB.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}

	q := db.New(testDB).WithTx(tx)

	cleanup := func() {
		tx.Rollback(ctx)
	}

	return q, cleanup
}

func TestCreateUser(t *testing.T) {
	q, cleanup := withTx(t)
	defer cleanup()

	ctx := context.Background()

	id := uuid.New()

	_, err := q.CreateUser(ctx, db.CreateUserParams{
		ID:       id,
		Email:    "test@example.com",
		Password: "secret",
	})
	if err != nil {
		t.Fatal(err)
	}
}