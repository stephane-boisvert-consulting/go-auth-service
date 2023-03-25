# Image de base
FROM golang:1.20.2-alpine3.17

# Configuration de l'environnement
ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# Copie des fichiers sources dans le conteneur
COPY . /app
WORKDIR /app

# Installation des dépendances et compilation du binaire
RUN apk add --no-cache git \
    && go mod download \
    && go build -o go-auth-service .

# Image finale
FROM alpine:3.17
RUN apk add --no-cache ca-certificates
WORKDIR /app

# Copie du binaire depuis la première image
COPY --from=0 /app /app

# Exposition du port
EXPOSE 8080

# Commande par défaut
CMD ["./go-auth-service"]
