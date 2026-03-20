package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResp struct {
	Token string `json:"token"`
}

type claims struct {
	User string `json:"user"`
	jwt.RegisteredClaims
}

func main() {
	var addr = flag.String("addr", "127.0.0.1:8070", "listen addr")
	var tokenTTL = flag.Duration("ttl", 30*time.Minute, "token ttl")
	flag.Parse()

	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if secret == "" {
		secret = "secret"
	}

	users := map[string]string{
		"admin": "123456",
		"test":  "test123",
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeErr(w, http.StatusMethodNotAllowed, 10001, "method not allowed")
			return
		}
		var req loginReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErr(w, http.StatusBadRequest, 10002, "bad json: "+err.Error())
			return
		}
		pass, ok := users[req.Username]
		if !ok || pass != req.Password {
			writeErr(w, http.StatusUnauthorized, 10003, "invalid username or password")
			return
		}

		claims := claims{
			User: req.Username,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "go-zero-learning",
				Subject:   req.Username,
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(*tokenTTL)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := token.SignedString([]byte(secret))
		if err != nil {
			writeErr(w, http.StatusInternalServerError, 10004, "sign token failed: "+err.Error())
			return
		}
		writeJSON(w, http.StatusOK, loginResp{Token: signed})
	})

	// A protected endpoint.
	mux.HandleFunc("/api/secure/hello", func(w http.ResponseWriter, r *http.Request) {
		user, err := requireJWT(r, secret)
		if err != nil {
			writeErr(w, http.StatusUnauthorized, 20001, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"message": "hello " + user,
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	log.Println("listening on", *addr)
	log.Println("login:  POST /api/login {\"username\":\"admin\",\"password\":\"123456\"}")
	log.Println("secure: GET  /api/secure/hello  with header Authorization: Bearer <token>")
	log.Fatal(http.ListenAndServe(*addr, mux))
}

func requireJWT(r *http.Request, secret string) (string, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", errors.New("missing Authorization header")
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("invalid Authorization header (expect Bearer)")
	}
	tokenStr := strings.TrimSpace(parts[1])
	if tokenStr == "" {
		return "", errors.New("empty bearer token")
	}

	t, err := jwt.ParseWithClaims(tokenStr, &claims{}, func(token *jwt.Token) (any, error) {
		// Ensure expected signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return "", err
	}
	c, ok := t.Claims.(*claims)
	if !ok || !t.Valid {
		return "", errors.New("invalid token")
	}
	if strings.TrimSpace(c.User) == "" {
		return "", errors.New("missing user in token claims")
	}
	return c.User, nil
}

func writeErr(w http.ResponseWriter, status int, code int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"code":    code,
		"message": msg,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

