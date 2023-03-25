package db

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Database représente une connexion à la base de données.
type Database struct {
	db *sql.DB
}

type HealthHandler struct {
	db          *Database
	maxRetries  int
	currentTry  int
	healthState int
}

const (
	HEALTH_STATE_UNKNOWN = iota
	HEALTH_STATE_OK
	HEALTH_STATE_ERROR
)

// Create a database connection
func Create() (*Database, error) {
	db, err := sql.Open("mysql", "user:password@tcp(db:3306)/dbname")
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	database := NewDatabase(db)
	return database, nil
}

// NewDatabase crée une nouvelle instance de la structure Database.
func NewDatabase(db *sql.DB) *Database {
	return &Database{db: db}
}

// Close ferme la connexion à la base de données.
func (d *Database) Close() error {
	return d.db.Close()
}

// QueryRow exécute une requête SQL qui renvoie une seule ligne.
func (d *Database) QueryRow(query string, args ...interface{}) *sql.Row {
	return d.db.QueryRow(query, args...)
}

// Query exécute une requête SQL qui renvoie plusieurs lignes.
func (d *Database) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return d.db.Query(query, args...)
}

// NewHealthHandler crée une nouvelle instance de la structure HealthHandler.
func NewHealthHandler(db *sql.DB, maxRetries int) *HealthHandler {
	return &HealthHandler{
		db:          NewDatabase(db),
		maxRetries:  maxRetries,
		currentTry:  0,
		healthState: HEALTH_STATE_UNKNOWN,
	}
}

// Handle gère les requêtes de test de santé.
func (h *HealthHandler) Handle(c *gin.Context) {
	if h.healthState == HEALTH_STATE_OK {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	} else if h.healthState == HEALTH_STATE_ERROR {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERROR"})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "UNKNOWN"})
	}
}

func (h *HealthHandler) Check() error {
	// Vérification de l'état de santé actuel
	if h.healthState == HEALTH_STATE_OK {
		return nil // La base de données est en bonne santé
	}

	// Vérification du nombre maximal de tentatives
	if h.currentTry >= h.maxRetries {
		h.healthState = HEALTH_STATE_ERROR
		return errors.New("Nombre maximal de tentatives atteint")
	}

	// Vérification de la connexion à la base de données
	err := h.db.Ping()
	if err != nil {
		// Incrémentation du nombre de tentatives actuelles
		h.currentTry++

		// Mise à jour de l'état de santé
		h.healthState = HEALTH_STATE_ERROR

		return err
	}

	// Réinitialisation du nombre de tentatives actuelles
	h.currentTry = 0

	// Mise à jour de l'état de santé
	h.healthState = HEALTH_STATE_OK

	return nil
}

// Ping vérifie si la base de données est disponible.
func (d *Database) Ping() error {
	return d.db.Ping()
}
