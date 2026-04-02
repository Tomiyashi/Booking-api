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

	// Public & Health
	mux.HandleFunc("GET /_info", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	mux.HandleFunc("POST /dummyLogin", srv.DummyLoginHandler)

	// Admin (нужна проверка роли "admin" в токене)
	mux.HandleFunc("POST /admin/rooms", srv.CreateRoomHandler)
	mux.HandleFunc("POST /admin/rooms/{id}/schedule", srv.CreateScheduleHandler)
	mux.HandleFunc("GET /admin/bookings", srv.ListAllBookingsHandler)

	// User (нужна проверка роли "user" в токене)
	mux.HandleFunc("GET /rooms/list", srv.ListRoomsHandler)
	mux.HandleFunc("GET /rooms/{id}/slots", srv.ListSlotsHandler)
	mux.HandleFunc("POST /bookings/create", srv.CreateBookingHandler)
	mux.HandleFunc("POST /bookings/{id}/cancel", srv.CancelBookingHandler)
	mux.HandleFunc("GET /bookings/my", srv.ListMyBookingsHandler)
	log.Println("Server started on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
