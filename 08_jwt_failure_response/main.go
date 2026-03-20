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

type apiErr struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func main() {
	var addr = flag.String("addr", "127.0.0.1:8080", "listen addr")
	var tokenTTL = flag.Duration("ttl", 10*time.Minute, "token ttl")
	flag.Parse()

	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if secret == "" {
		secret = "secret"
	}

	users := map[string]string{
		"admin": "123456",
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, apiErr{Code: 10001, Message: "method not allowed"})
			return
		}
		var req loginReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, apiErr{Code: 10002, Message: "bad json: " + err.Error()})
			return
		}
		pass, ok := users[req.Username]
		if !ok || pass != req.Password {
			writeJSON(w, http.StatusUnauthorized, apiErr{Code: 10003, Message: "invalid username or password"})
			return
		}

		c := claims{
			User: req.Username,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "go-zero-learning",
				Subject:   req.Username,
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(*tokenTTL)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
		signed, err := token.SignedString([]byte(secret))
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, apiErr{Code: 10004, Message: "sign token failed: " + err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, loginResp{Token: signed})
	})

	mux.HandleFunc("/api/secure/hello", func(w http.ResponseWriter, r *http.Request) {
		user, apiError := requireJWT(r, secret)
		if apiError != nil {
			writeJSON(w, http.StatusUnauthorized, *apiError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"message": "hello " + user,
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	log.Println("listening on", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

// returns: (user, error). error is already mapped to apiErr for consistent front-end handling.
func requireJWT(r *http.Request, secret string) (string, *apiErr) {
	auth := r.Header.Get("Authorization")
	if strings.TrimSpace(auth) == "" {
		return "", &apiErr{Code: 20001, Message: "missing Authorization header"}
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", &apiErr{Code: 20002, Message: "invalid Authorization header (expect Bearer <token>)"}
	}
	tokenStr := strings.TrimSpace(parts[1])
	if tokenStr == "" {
		return "", &apiErr{Code: 20003, Message: "empty bearer token"}
	}

	t, err := jwt.ParseWithClaims(tokenStr, &claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		// Distinguish common JWT failures for clearer UI behavior.
		var ve *jwt.ValidationError
		if errors.As(err, &ve) {
			switch {
			case ve.Errors&jwt.ValidationErrorMalformed != 0:
				return "", &apiErr{Code: 20004, Message: "token malformed"}
			case ve.Errors&jwt.ValidationErrorExpired != 0:
				return "", &apiErr{Code: 20005, Message: "token expired"}
			case ve.Errors&jwt.ValidationErrorSignatureInvalid != 0:
				return "", &apiErr{Code: 20006, Message: "token signature invalid"}
			default:
				return "", &apiErr{Code: 20007, Message: "token invalid: " + err.Error()}
			}
		}
		return "", &apiErr{Code: 20007, Message: "token invalid: " + err.Error()}
	}

	c, ok := t.Claims.(*claims)
	if !ok || !t.Valid {
		return "", &apiErr{Code: 20007, Message: "token invalid"}
	}
	if strings.TrimSpace(c.User) == "" {
		return "", &apiErr{Code: 20008, Message: "missing user in token claims"}
	}
	return c.User, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

