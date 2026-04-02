package main

import (
	postgres "booking-api/internal/database"
	api "booking-api/internal/server"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	db, err := postgres.NewPostgresDB(dsn)
	if err != nil {
		log.Fatal("Could not connect to DB: ", err)
	}
	defer db.Close()

	srv := &api.Server{DB: db}

	mux := http.NewServeMux()

	mux.HandleFunc("/_info", srv.InfoHandler)
	mux.HandleFunc("/dummyLogin", srv.DummyLoginHandler)
	mux.HandleFunc("/rooms/list", srv.GetAllRoomHandler)
	mux.HandleFunc("/bookings/create", srv.CreateBookingHandler)

	mux.HandleFunc("/rooms/", srv.RoutesHandler)
	mux.HandleFunc("/bookings/", srv.RoutesHandler)

	log.Println("Server started on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
