package mysqlc

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/krzysztofkolcz/my-sqlc/internal/db"
)

func main() {
    ctx := context.Background()

    pool, err := pgxpool.New(ctx, "postgres://mysqlcuser:mypass@localhost:5432/mysqlcdb")
    if err != nil {
        panic(err)
    }

    queries := db.New(pool)

    user, err := queries.GetUserByEmail(ctx, "test@example.com")
    if err != nil {
        panic(err)
    }

    fmt.Println(user)
}