package models

import "time"

type Room struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Capacity    int    `json:"capacity"`
}

type Slot struct {
	ID        string    `json:"id"`
	RoomID    string    `json:"roomId"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	IsBooked  bool      `json:"isBooked"`
}

type Booking struct {
	ID     string `json:"id"`
	SlotID string `json:"slotId"`
	UserID string `json:"userId"`
	Status string `json:"status"`
}
