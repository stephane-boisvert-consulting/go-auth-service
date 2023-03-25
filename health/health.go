package health

import (
	"errors"
	"go-auth-service/db"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	db          *db.Database
	maxRetries  int
	currentTry  int
	healthState int
}

const (
	HEALTH_STATE_UNKNOWN = iota
	HEALTH_STATE_OK
	HEALTH_STATE_ERROR
)

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

// Start a background routine to check the health status periodically
func loop(h *HealthHandler, d *db.Database) {
	maxAttempts := 10
	attempt := 0
	delay := 2 * time.Second

	for {
		attempt++
		err := d.Ping()
		if err == nil {
			// La vérification de la santé est réussie, réinitialiser le délai de backoff et le nombre de tentatives
			attempt = 0
			h.healthState = HEALTH_STATE_OK
		} else {
			// Une erreur s'est produite lors de la vérification de la santé, enregistrer le message d'erreur et attendre avant de réessayer
			h.healthState = HEALTH_STATE_ERROR
			log.Printf("Error checking health (attempt %d/%d, will retry in %v): %v\n", attempt, maxAttempts, delay, err)
			if attempt == maxAttempts {
				log.Fatalf("Failed to check health after %d attempts: %v", maxAttempts, err)
				os.Exit(1)
			}
		}

		time.Sleep(delay)
	}
}

// NewHealthHandler crée une nouvelle instance de la structure HealthHandler.
func NewHealthHandler(db *db.Database, maxRetries int) *HealthHandler {
	h := &HealthHandler{
		db:          db,
		maxRetries:  maxRetries,
		currentTry:  0,
		healthState: HEALTH_STATE_UNKNOWN,
	}
	log.Printf("start go looo")
	go loop(h, db)
	return h
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
