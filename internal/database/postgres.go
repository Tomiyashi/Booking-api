package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

func NewPostgresDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)

	if err != nil {
		return nil, err
	}

	for i := 0; i <= 10; i++ {
		if err = db.Ping(); err == nil {
			return db, err
		}
		log.Println("База спит")
		time.Sleep(2 * time.Second)
	}
	return nil, fmt.Errorf("Не удалось подключится к базе после 10 попыток")
}
