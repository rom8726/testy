package testyexample

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

type Server struct {
	Router http.Handler
}

func NewServer(connStr string) *Server {
	router := httprouter.New()

	authKeys := map[string]struct{}{}

	router.POST("/auth", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		type authRequest struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		var req authRequest
		if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			fmt.Println(err)

			return
		}

		if req.Username != "user" || req.Password != "password" {
			writer.WriteHeader(http.StatusUnauthorized)

			return
		}

		authKey := uuid.NewString()
		authKeys[authKey] = struct{}{}

		type response struct {
			Token string `json:"token"`
		}
		resp := response{Token: authKey}
		_ = json.NewEncoder(writer).Encode(resp)

		return
	})

	router.POST("/users/add", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		defer func() {
			if r := recover(); r != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				fmt.Println(r)
			}
		}()

		token := request.Header.Get("Authorization")
		if _, ok := authKeys[token]; !ok {
			writer.WriteHeader(http.StatusUnauthorized)

			return
		}

		db, err := sql.Open("postgres", connStr)
		if err != nil {
			panic(err)
		}
		defer db.Close()

		data, err := io.ReadAll(request.Body)
		if err != nil {
			panic(err)
		}
		defer request.Body.Close()

		type userRequest struct {
			Name string `json:"name"`
		}
		var req userRequest
		if err := json.Unmarshal(data, &req); err != nil {
			panic(err)
		}

		const query = `INSERT INTO users (name, created_at) VALUES ($1, NOW())`
		_, err = db.ExecContext(request.Context(), query, req.Name)
		if err != nil {
			panic(err)
		}

		if err := sendAddUserNotification(req.Name); err != nil {
			panic(err)
		}

		writer.WriteHeader(http.StatusOK)

		return
	})

	router.GET("/users", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		defer func() {
			if r := recover(); r != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				fmt.Println(r)
			}
		}()

		token := request.Header.Get("Authorization")
		if _, ok := authKeys[token]; !ok {
			writer.WriteHeader(http.StatusUnauthorized)

			return
		}

		db, err := sql.Open("postgres", connStr)
		if err != nil {
			panic(err)
		}
		defer db.Close()

		const query = `SELECT id, name FROM users ORDER BY id ASC`
		rows, err := db.QueryContext(request.Context(), query)
		if err != nil {
			panic(err)
		}
		defer rows.Close()

		type userResponse struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}
		var users []userResponse
		for rows.Next() {
			var user userResponse
			if err := rows.Scan(&user.ID, &user.Name); err != nil {
				panic(err)
			}

			users = append(users, user)
		}

		type targetResponse struct {
			Users []userResponse `json:"users"`
		}
		targetResp := targetResponse{Users: users}

		writer.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(writer).Encode(targetResp)

		return
	})

	router.GET("/user/:user_id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		defer func() {
			if r := recover(); r != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				fmt.Println(r)
			}
		}()

		token := request.Header.Get("Authorization")
		if _, ok := authKeys[token]; !ok {
			writer.WriteHeader(http.StatusUnauthorized)

			return
		}

		db, err := sql.Open("postgres", connStr)
		if err != nil {
			panic(err)
		}
		defer db.Close()

		const query = `SELECT id, name FROM users WHERE id = $1 LIMIT 1`
		var userID int
		var userName string
		err = db.QueryRowContext(request.Context(), query, params.ByName("user_id")).Scan(&userID, &userName)
		if err != nil {
			panic(err)
		}

		type userResponse struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}
		resp := userResponse{ID: userID, Name: userName}

		writer.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(writer).Encode(resp)

		return
	})

	return &Server{
		Router: router,
	}
}

func sendAddUserNotification(name string) error {
	baseURL := os.Getenv("NOTIFICATION_BASE_URL")
	if baseURL == "" {
		return fmt.Errorf("NOTIFICATION_BASE_URL is not set")
	}

	type notificationRequest struct {
		Name string `json:"name"`
	}
	req := notificationRequest{Name: name}
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(baseURL+"/send", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
