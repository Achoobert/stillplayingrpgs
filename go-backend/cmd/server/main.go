package main

import (
	"database/sql"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"startplaying-clone/internal/models"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

// TemplateRenderer is a custom html/template renderer for Echo
type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	// Initialize Database
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := initDB(db); err != nil {
		log.Fatal(err)
	}
	// Seed Prototype Users if not exists
	var count int
	db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if count == 0 {
		passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
		db.Exec("INSERT INTO users (username, password_hash, role) VALUES (?, ?, ?)", "gm", string(passwordHash), "GM")
		db.Exec("INSERT INTO users (username, password_hash, role) VALUES (?, ?, ?)", "player", string(passwordHash), "Player")
		// Seed prototype game listing
		db.Exec("INSERT INTO games (title, system, gm_id, start_time, price, max_players, description) VALUES (?, ?, ?, ?, ?, ?, ?)",
			"Dragon of Icespire Peak", "D&D 5e", 1, time.Now().Add(24*time.Hour), 15.0, 5, "An introductory adventure for the world's greatest roleplaying game.")
		log.Println("Seeded prototype users and game listing")
	}

	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Static("/static", "static")

	// Template Renderer
	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}
	e.Renderer = renderer

	// Routes
	e.GET("/", func(c echo.Context) error {
		games, _ := getGames(db)
		return c.Render(http.StatusOK, "index.html", map[string]interface{}{
			"games": games,
		})
	})

	e.GET("/login", func(c echo.Context) error {
		return c.Render(http.StatusOK, "login.html", nil)
	})

	e.GET("/register", func(c echo.Context) error {
		return c.Render(http.StatusOK, "register.html", nil)
	})

	e.POST("/register", func(c echo.Context) error {
		username := c.FormValue("username")
		password := c.FormValue("password")
		role := c.FormValue("role")

		passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		_, err := db.Exec("INSERT INTO users (username, password_hash, role) VALUES (?, ?, ?)", username, string(passwordHash), role)
		if err != nil {
			return c.Render(http.StatusOK, "register.html", map[string]interface{}{
				"error": "Username already exists",
			})
		}

		return c.Redirect(http.StatusSeeOther, "/login")
	})

	e.POST("/login", func(c echo.Context) error {
		username := c.FormValue("username")
		password := c.FormValue("password")

		user, err := getUserByUsername(db, username)
		if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
			return c.Redirect(http.StatusSeeOther, "/login")
		}

		// Set simple secure cookie for proto-auth (can upgrade to JWT later)
		cookie := new(http.Cookie)
		cookie.Name = "session_token"
		cookie.Value = user.Username
		cookie.HttpOnly = true
		cookie.Path = "/"
		c.SetCookie(cookie)

		if user.Role == "GM" {
			return c.Redirect(http.StatusSeeOther, "/gm_dashboard")
		}
		return c.Redirect(http.StatusSeeOther, "/player_dashboard")
	})

	e.POST("/logout", func(c echo.Context) error {
		cookie := new(http.Cookie)
		cookie.Name = "session_token"
		cookie.Value = ""
		cookie.MaxAge = -1
		cookie.Path = "/"
		c.SetCookie(cookie)
		return c.Redirect(http.StatusSeeOther, "/")
	})
	e.GET("/gm_dashboard", func(c echo.Context) error {
		user, err := getCurrentUser(c, db)
		if err != nil || user.Role != "GM" {
			return c.Redirect(http.StatusSeeOther, "/login")
		}
		games, _ := getGMGames(db, user.ID)
		return c.Render(http.StatusOK, "gm_dashboard.html", map[string]interface{}{
			"current_user": user,
			"games":        games,
		})
	})

	e.POST("/gm_dashboard/create_game", func(c echo.Context) error {
		user, err := getCurrentUser(c, db)
		if err != nil || user.Role != "GM" {
			return c.Redirect(http.StatusSeeOther, "/login")
		}

		startTime, _ := time.Parse("2006-01-02T15:04", c.FormValue("start_time"))
		price, _ := strconv.ParseFloat(c.FormValue("price"), 64)
		maxPlayers, _ := strconv.Atoi(c.FormValue("max_players"))

		game := models.Game{
			Title:       c.FormValue("title"),
			System:      c.FormValue("system"),
			GMID:        user.ID,
			StartTime:   startTime,
			Price:       price,
			MaxPlayers:  maxPlayers,
			Description: c.FormValue("description"),
		}

		if err := createGame(db, game); err != nil {
			return c.String(http.StatusInternalServerError, "Failed to create game")
		}
		return c.Redirect(http.StatusSeeOther, "/gm_dashboard")
	})

	e.GET("/player_dashboard", func(c echo.Context) error {
		user, err := getCurrentUser(c, db)
		if err != nil || user.Role != "Player" {
			return c.Redirect(http.StatusSeeOther, "/login")
		}
		bookings, _ := getPlayerBookings(db, user.ID)
		return c.Render(http.StatusOK, "player_dashboard.html", map[string]interface{}{
			"current_user": user,
			"bookings":     bookings,
		})
	})

	e.POST("/game/:id/join", func(c echo.Context) error {
		user, err := getCurrentUser(c, db)
		if err != nil || user.Role != "Player" {
			return c.Redirect(http.StatusSeeOther, "/login")
		}
		gameID, _ := strconv.Atoi(c.Param("id"))
		if err := createBooking(db, gameID, user.ID); err != nil {
			return c.String(http.StatusInternalServerError, "Failed to join game")
		}
		return c.Redirect(http.StatusSeeOther, "/player_dashboard")
	})

	// Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "30011"
	}
	e.Logger.Fatal(e.Start(":" + port))
}

func initDB(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE,
		password_hash TEXT,
		role TEXT
	);
	CREATE TABLE IF NOT EXISTS games (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT,
		system TEXT,
		gm_id INTEGER,
		start_time DATETIME,
		price REAL,
		max_players INTEGER,
		description TEXT,
		FOREIGN KEY(gm_id) REFERENCES users(id)
	);
	CREATE TABLE IF NOT EXISTS bookings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		game_id INTEGER,
		player_id INTEGER,
		status TEXT,
		FOREIGN KEY(game_id) REFERENCES games(id),
		FOREIGN KEY(player_id) REFERENCES users(id)
	);`
	_, err := db.Exec(schema)
	return err
}

func getGames(db *sql.DB) ([]models.Game, error) {
	rows, err := db.Query("SELECT id, title, system, gm_id, start_time, price, max_players, description FROM games")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []models.Game
	for rows.Next() {
		var g models.Game
		err := rows.Scan(&g.ID, &g.Title, &g.System, &g.GMID, &g.StartTime, &g.Price, &g.MaxPlayers, &g.Description)
		if err != nil {
			return nil, err
		}
		games = append(games, g)
	}
	return games, nil
}

func getUserByUsername(db *sql.DB, username string) (*models.User, error) {
	var u models.User
	err := db.QueryRow("SELECT id, username, password_hash, role FROM users WHERE username = ?", username).
		Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
func getCurrentUser(c echo.Context, db *sql.DB) (*models.User, error) {
	cookie, err := c.Cookie("session_token")
	if err != nil {
		return nil, err
	}
	return getUserByUsername(db, cookie.Value)
}

func getGMGames(db *sql.DB, gmID int) ([]models.Game, error) {
	rows, err := db.Query("SELECT id, title, system, gm_id, start_time, price, max_players, description FROM games WHERE gm_id = ?", gmID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []models.Game
	for rows.Next() {
		var g models.Game
		err := rows.Scan(&g.ID, &g.Title, &g.System, &g.GMID, &g.StartTime, &g.Price, &g.MaxPlayers, &g.Description)
		if err != nil {
			return nil, err
		}
		games = append(games, g)
	}
	return games, nil
}

func createGame(db *sql.DB, g models.Game) error {
	_, err := db.Exec("INSERT INTO games (title, system, gm_id, start_time, price, max_players, description) VALUES (?, ?, ?, ?, ?, ?, ?)",
		g.Title, g.System, g.GMID, g.StartTime, g.Price, g.MaxPlayers, g.Description)
	return err
}

func getPlayerBookings(db *sql.DB, playerID int) ([]models.Booking, error) {
	rows, err := db.Query("SELECT id, game_id, player_id, status FROM bookings WHERE player_id = ?", playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []models.Booking
	for rows.Next() {
		var b models.Booking
		err := rows.Scan(&b.ID, &b.GameID, &b.PlayerID, &b.Status)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, b)
	}
	return bookings, nil
}

func createBooking(db *sql.DB, gameID, playerID int) error {
	_, err := db.Exec("INSERT INTO bookings (game_id, player_id, status) VALUES (?, ?, ?)",
		gameID, playerID, "confirmed")
	return err
}
