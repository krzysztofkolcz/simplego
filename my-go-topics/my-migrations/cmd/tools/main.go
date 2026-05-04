package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/krzysztofkolcz/mymigrations/db"
)

/*
go run ./cmd/tools migrate:new add_orders
make migrate-new name=add_orders
*/
func main() {
	if len(os.Args) < 3 {
		fmt.Println("usage: migrate:new <name>")
		os.Exit(1)
	}

	command := os.Args[1]
	name := os.Args[2]

	switch command {
	case "migrate:new":
		createMigration(name)
	case "migrate:validate":
		err := db.ValidateMigrations(db.MigrationsFS, "migrations/tenant")
		if err != nil {
			panic(err)
		}
		fmt.Println("OK")
	default:
		fmt.Println("unknown command")
	}
}

func createMigration(name string) {
	timestamp := time.Now().Format("20060102150405")

	basePath := "db/migrations/tenant"

	upFile := filepath.Join(basePath, fmt.Sprintf("%s_%s.up.sql", timestamp, name))
	downFile := filepath.Join(basePath, fmt.Sprintf("%s_%s.down.sql", timestamp, name))

	writeFile(upFile, "-- Write your UP migration here\n")
	writeFile(downFile, "-- Write your DOWN migration here\n")

	fmt.Println("Created:")
	fmt.Println(upFile)
	fmt.Println(downFile)
}

func writeFile(path, content string) {
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		panic(err)
	}
}