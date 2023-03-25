package auth

import (
	"database/sql"
	"go-auth-service/db"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// AuthInfo contient les informations d'authentification d'un utilisateur.
type AuthInfo struct {
	Username string
	Password string
	isAdmin  bool
}

// AuthHandler gère les requêtes d'authentification.
type AuthHandler struct {
	db   *db.Database
	conf *oauth2.Config
}

// isAuthInfoValid vérifie si les informations d'authentification sont valides.
func (h *AuthHandler) isAuthInfoValid(authInfo *AuthInfo, providedAuthInfo AuthInfo) bool {
	log.Printf("Auth info validation started...")                  // Ajouter un message de debug
	log.Printf("Provided username: %s", providedAuthInfo.Username) // Afficher le nom d'utilisateur fourni
	log.Printf("Provided password: %s", providedAuthInfo.Password) // Afficher le mot de passe fourni
	log.Printf("Provided token: %s", providedAuthInfo.Token)       // Afficher le mot de passe fourni
	log.Printf("Retreived username: %s", authInfo.Username)        // Afficher le nom d'utilisateur fourni
	log.Printf("Retreived password: %s", authInfo.Password)        // Afficher le mot de passe fourni
	if authInfo.Password == providedAuthInfo.Password {
		log.Printf("Auth info validation succeeded.") // Ajouter un message de debug
		return true
	} else {
		log.Printf("Auth info validation failed.") // Ajouter un message de debug
		return false
	}
}

// NewAuthHandler crée une nouvelle instance de la structure AuthHandler.
func NewAuthHandler(db *db.Database) *AuthHandler {
	// Configuration du client Google OAuth
	conf := &oauth2.Config{
		ClientID:     "client_id",
		ClientSecret: "client_secret",
		RedirectURL:  "http://localhost:8080/oauth2/callback",
		Scopes:       []string{"email", "profile"},
		Endpoint:     google.Endpoint,
	}

	return &AuthHandler{
		db:   db,
		conf: conf,
	}
}

// Handle gère la requête d'authentification.
func (h *AuthHandler) Handle(c *gin.Context) {
	log.Printf("handle")
	// Récupération du nom d'utilisateur et du token depuis la requête

	providedAuthInfo := AuthInfo{
		Username: c.PostForm("username"),
		Password: c.PostForm("password"),
		Token:    "mytoken123",
	}

	// Vérification des informations d'authentification sur le backend de la base de données
	authInfoFromDB, err := h.getAuthInfoFromDB(providedAuthInfo.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			c.AbortWithStatus(http.StatusForbidden) // utilisateur inexistant
			return
		}
		log.Printf("Erreur lors de la vérification des informations d'authentification de la base de données: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Vérification de la validité des informations d'authentification
	if !h.isAuthInfoValid(authInfoFromDB, providedAuthInfo) {
		c.Status(http.StatusForbidden)
		return
	}

	// Si les informations d'authentification sont valides, générez un token JWT
	tokenString, err := h.generateJWTToken(authInfoFromDB.Username)
	log.Printf("tokenstring: %v", err)
	if err != nil {
		log.Printf("Erreur lors de la génération du token JWT: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Créer le cookie
	cookie := http.Cookie{
		Name:  "mytoken",
		Value: tokenString,
	}

	// Définir le cookie
	http.SetCookie(c.Writer, &cookie)

}

// getAuthInfoFromDB récupère les informations d'authentification de la base de données.
func (h *AuthHandler) getAuthInfoFromDB(username string) (*AuthInfo, error) {
	// Requête pour récupérer les informations d'authentification
	row := h.db.QueryRow("SELECT username, password FROM users WHERE username = ?", username)

	// Récupération des informations d'authentification

	authInfo := &AuthInfo{}

	log.Printf("authinfo %v", authInfo)
	err := row.Scan(&authInfo.Username, &authInfo.Password, &authInfo.isAdmin)
	if err != nil {
		return nil, err
	}

	return authInfo, nil
}

// Structure de la charge utile (payload) du jeton JWT
type Claims struct {
	Username string
	jwt.StandardClaims
}

// generateJWTToken génère un nouveau token JWT pour le nom d'utilisateur donné.
func (h *AuthHandler) generateJWTToken(username string) (string, error) {
	// Création d'un nouveau token JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 1).Unix(),
			Issuer:    "mon_issuer",
		},
	})
	// Signature et récupération du token complet encodé en tant que chaîne de caractères
	tokenString, err := token.SignedString([]byte("my-secret-key"))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
