package server

import (
	"booking-api/internal/models"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type Server struct {
	DB *sql.DB
}

// GET /_info
func (s *Server) InfoHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// POST /dummyLogin
func (s *Server) DummyLoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": "fixed-test-token"})
}

// GET /rooms/list
func (s *Server) GetAllRoomHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.Query("SELECT id, name, description, capacity FROM rooms")
	if err != nil {
		log.Printf("Ошибка получения списка комнат: %v", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	defer rows.Close()

	var rooms []models.Room
	for rows.Next() {
		var rm models.Room
		if err := rows.Scan(&rm.ID, &rm.Name, &rm.Description, &rm.Capacity); err != nil {
			continue
		}
		rooms = append(rooms, rm)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rooms)
}

// GET /rooms/{id}/slots?date=YYYY-MM-DD
func (s *Server) GetSlotsHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		http.Error(w, "Room ID required", 400)
		return
	}
	roomID := parts[1]
	dateStr := r.URL.Query().Get("date")

	if dateStr == "" {
		http.Error(w, "Date parameter is required", 400)
		return
	}

	query := `
		SELECT s.id, s.room_id, s.start_time, s.end_time 
		FROM slots s
		LEFT JOIN bookings b ON s.id = b.slot_id AND b.status = 'active'
		WHERE s.room_id = $1 AND DATE(s.start_time) = $2 AND b.id IS NULL
		ORDER BY s.start_time ASC`

	rows, err := s.DB.Query(query, roomID, dateStr)
	if err != nil {
		http.Error(w, "Database error", 500)
		return
	}
	defer rows.Close()

	var slots []models.Slot
	for rows.Next() {
		var sl models.Slot
		rows.Scan(&sl.ID, &sl.RoomID, &sl.StartTime, &sl.EndTime)
		slots = append(slots, sl)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(slots)
}

// POST /bookings/create
func (s *Server) CreateBookingHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SlotID string `json:"slotId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}

	// Заглушка ID пользователя для тестов
	userID := "00000000-0000-0000-0000-000000000001"

	var bookingID string
	query := `INSERT INTO bookings (slot_id, user_id, status) VALUES ($1, $2, 'active') RETURNING id`
	err := s.DB.QueryRow(query, req.SlotID, userID).Scan(&bookingID)

	if err != nil {
		http.Error(w, "Slot already booked or invalid", http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": bookingID})
}

// POST /bookings/{id}/cancel
func (s *Server) CancelBookingHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		http.Error(w, "Booking ID required", 400)
		return
	}
	bookingID := parts[1]

	query := `UPDATE bookings SET status = 'cancelled' WHERE id = $1`
	res, err := s.DB.Exec(query, bookingID)
	if err != nil {
		http.Error(w, "DB Error", 500)
		return
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		http.Error(w, "Booking not found", 404)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Cancelled"))
}

// Общий обработчик для сложных путей
func (s *Server) RoutesHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if strings.HasPrefix(path, "/rooms/") && strings.HasSuffix(path, "/slots") {
		if r.Method == http.MethodGet {
			s.GetSlotsHandler(w, r)
			return
		}
	}

	if strings.HasPrefix(path, "/bookings/") && strings.HasSuffix(path, "/cancel") {
		if r.Method == http.MethodPost {
			s.CancelBookingHandler(w, r)
			return
		}
	}

	http.NotFound(w, r)
}
