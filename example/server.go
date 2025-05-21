package testyexample

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Server struct {
	Router http.Handler
}

func NewServer(connStr string) *Server {
	router := httprouter.New()
	router.POST("/users/add", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		defer func() {
			if r := recover(); r != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				fmt.Println(r)
			}
		}()

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

	return &Server{
		Router: router,
	}
}
