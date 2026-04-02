package server

import (
	"booking-api/internal/models"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Server struct {
	DB *sql.DB
}

// GET /_info
func (s *Server) InfoHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// POST /admin/rooms
func (s *Server) CreateRoomHandler(w http.ResponseWriter, r *http.Request) {
	var rm struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Capacity    int    `json:"capacity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&rm); err != nil {
		http.Error(w, "invalid body", 400)
		return
	}
	id := uuid.New().String()
	_, err := s.DB.Exec(`INSERT INTO rooms (id, name, description, capacity) VALUES ($1, $2, $3, $4)`,
		id, rm.Name, rm.Description, rm.Capacity)
	if err != nil {
		http.Error(w, "db error", 500)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

// POST /admin/rooms/{id}/schedule
func (s *Server) CreateScheduleHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("id")
	var req struct {
		DaysOfWeek []int  `json:"daysOfWeek"`
		StartTime  string `json:"startTime"`
		EndTime    string `json:"endTime"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", 400)
		return
	}

	startT, _ := time.Parse("15:04", req.StartTime)
	endT, _ := time.Parse("15:04", req.EndTime)

	tx, _ := s.DB.Begin()
	for i := 0; i < 7; i++ {
		date := time.Now().AddDate(0, 0, i)
		isTargetDay := false
		for _, d := range req.DaysOfWeek {
			if int(date.Weekday()) == d {
				isTargetDay = true
				break
			}
		}
		if !isTargetDay {
			continue
		}

		curr := time.Date(date.Year(), date.Month(), date.Day(), startT.Hour(), startT.Minute(), 0, 0, time.UTC)
		limit := time.Date(date.Year(), date.Month(), date.Day(), endT.Hour(), endT.Minute(), 0, 0, time.UTC)

		for curr.Before(limit) {
			slotEnd := curr.Add(30 * time.Minute)
			tx.Exec(`INSERT INTO slots (id, room_id, start_time, end_time) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING`,
				uuid.New().String(), roomID, curr, slotEnd)
			curr = slotEnd
		}
	}
	tx.Commit()
	w.WriteHeader(http.StatusCreated)
}

// GET /rooms/list
func (s *Server) ListRoomsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.Query("SELECT id, name, description, capacity FROM rooms")
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	defer rows.Close()

	var rooms []models.Room
	for rows.Next() {
		var rm models.Room
		rows.Scan(&rm.ID, &rm.Name, &rm.Description, &rm.Capacity)
		rooms = append(rooms, rm)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rooms)
}

// GET /rooms/{id}/slots
func (s *Server) ListSlotsHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("id")
	dateStr := r.URL.Query().Get("date")

	if dateStr == "" {
		http.Error(w, "Date required", 400)
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
		http.Error(w, "DB error", 500)
		return
	}
	defer rows.Close()

	slots := []models.Slot{} // Инициализируем пустым слайсом вместо nil
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
	json.NewDecoder(r.Body).Decode(&req)

	bookingID := uuid.New().String()
	userID := "00000000-0000-0000-0000-000000000001" // В идеале из JWT

	query := `INSERT INTO bookings (id, slot_id, user_id, status) VALUES ($1, $2, $3, 'active')`
	_, err := s.DB.Exec(query, bookingID, req.SlotID, userID)

	if err != nil {
		http.Error(w, "Already booked", 409)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": bookingID})
}

// POST /bookings/{id}/cancel
func (s *Server) CancelBookingHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, err := s.DB.Exec(`UPDATE bookings SET status = 'cancelled' WHERE id = $1`, id)
	if err != nil {
		http.Error(w, "DB error", 500)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GET /bookings/my
func (s *Server) ListMyBookingsHandler(w http.ResponseWriter, r *http.Request) {
	userID := "00000000-0000-0000-0000-000000000001"
	rows, err := s.DB.Query(`SELECT id, slot_id, status FROM bookings WHERE user_id = $1 AND status = 'active'`, userID)
	if err != nil {
		http.Error(w, "DB error", 500)
		return
	}
	defer rows.Close()

	var bookings []map[string]string
	for rows.Next() {
		var id, sid, stat string
		rows.Scan(&id, &sid, &stat)
		bookings = append(bookings, map[string]string{"id": id, "slotId": sid, "status": stat})
	}
	json.NewEncoder(w).Encode(bookings)
}

// GET /admin/bookings
func (s *Server) ListAllBookingsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.Query(`SELECT id, slot_id, user_id, status FROM bookings`)
	if err != nil {
		http.Error(w, "DB error", 500)
		return
	}
	defer rows.Close()

	bookings := []map[string]string{}
	for rows.Next() {
		var id, sid, uid, stat string
		rows.Scan(&id, &sid, &uid, &stat)
		bookings = append(bookings, map[string]string{"id": id, "slotId": sid, "userId": uid, "status": stat})
	}
	json.NewEncoder(w).Encode(bookings)
}

func (s *Server) DummyLoginHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"token": "fixed-test-token"})
}
