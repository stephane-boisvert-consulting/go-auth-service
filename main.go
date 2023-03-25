package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"go-auth-service/auth"
	"go-auth-service/db"
	"go-auth-service/health"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

// Define the interval at which to check the health status
const checkInterval = 10 * time.Second

func main() {

	var datab *db.Database
	var err error

	for {
		datab, err = db.Create()
		if err != nil {
			log.Println(err)
			time.Sleep(5 * time.Second) // Attendre 5 secondes avant de réessayer
			continue
		}
		break
	}

	defer datab.Close()

	//err := healthHandler.Check()

	// Create a new HealthHandler
	healthHandler := health.NewHealthHandler(datab, 3)
	authHandler := auth.NewAuthHandler(datab)

	// Set up a Gin router and register the health check endpoint
	router := gin.Default()

	router.GET("/health", healthHandler.Handle)
	router.POST("/auth", authHandler.Handle)

	router.Static("/static", "./static")

	// Définir une route pour servir la page de login
	router.GET("/login", func(c *gin.Context) {
		receivedCookie, err := c.Cookie("mytoken")

		if err != nil {
			// Le cookie n'existe pas
			c.File("./static/login.html")
		} else {
			// Le cookie existe, afficher sa valeur
			fmt.Println("Valeur du cookie:", receivedCookie)
			if validateJWTToken(receivedCookie) {
			} else {
				c.File("./static/login.html")
			}
		}

	})

	router.GET("/logout", func(c *gin.Context) {
		// Supprimer le cookie en définissant une date d'expiration passée
		cookie := http.Cookie{
			Name:   "mytoken",
			Value:  "",
			MaxAge: -1,
		}

		// Définir le cookie dans la réponse
		http.SetCookie(c.Writer, &cookie)

		// Rediriger l'utilisateur vers la page de connexion
		c.Redirect(http.StatusFound, "/login")
	})

	router.Run(":8080")
}

type Claims struct {
	Username string

	jwt.StandardClaims
}

func validateJWTToken(tokenString string) bool {
	// Parse le jeton JWT en une structure de jeton
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("my-secret-key"), nil
	})
	if err != nil {
		log.Printf("Erreur lors de l'analyse du jeton JWT: %v, token %v", err, token)
		return false
	}

	// Vérifie si le jeton est valide
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		fmt.Printf("%+v\n", token.Claims)
		log.Printf("Jeton JWT valide pour l'utilisateur: %v", claims.Username)
		return true
	} else {
		log.Println("Jeton JWT invalide")
		return false
	}
}
