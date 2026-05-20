package models

import "time"

type User struct {
	ID           int    `db:"id"`
	Username     string `db:"username"`
	PasswordHash string `db:"password_hash"`
	Role         string `db:"role"` // "GM" or "Player"
}

type Game struct {
	ID          int       `db:"id"`
	Title       string    `db:"title"`
	System      string    `db:"system"`
	GMID        int       `db:"gm_id"`
	StartTime   time.Time `db:"start_time"`
	Price       float64   `db:"price"`
	MaxPlayers  int       `db:"max_players"`
	Description string    `db:"description"`
}

type Booking struct {
	ID       int    `db:"id"`
	GameID   int    `db:"game_id"`
	PlayerID int    `db:"player_id"`
	Status   string `db:"status"`
}
