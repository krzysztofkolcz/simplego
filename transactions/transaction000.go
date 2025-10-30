package transactions

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"simpleGo/model"
)

func Transaction000() {

	dsn := "host=localhost user=postgres password=secret dbname=cmk port=5433 sslmode=disable"
	db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	db.AutoMigrate(&model.User{})

	// Session A: trzyma transakcję
	tx := db.Begin()
	defer tx.Rollback() // rollback na koniec

	fmt.Println("BEGIN - brak locków dopóki nie zrobimy SELECT FOR UPDATE")
	time.Sleep(5 * time.Second)

	var u model.User
	tx.First(&u, 1) // zwykły SELECT – brak blokady
	fmt.Println("SELECT zwykły - nadal brak locków")

	tx.Raw("SELECT * FROM users WHERE id = ? FOR UPDATE", 1).Scan(&u)
	fmt.Println("SELECT FOR UPDATE - mamy row lock na wierszu ID=1")

	time.Sleep(30 * time.Second) // trzymamy lock
	tx.Commit()

}
