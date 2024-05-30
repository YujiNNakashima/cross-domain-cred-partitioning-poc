package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/rs/cors"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

type TokenRequest struct {
	Key   string `json:"key"`
	Token string `json:"token"`
}

type GetTokenRequest struct {
	Racf string `json:"racf"`
}

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	http.HandleFunc("/store-token", func(w http.ResponseWriter, r *http.Request) {

		fmt.Println(r)

		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var req TokenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err := rdb.Set(ctx, req.Key, req.Token, 0).Err()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Token stored successfully"))
	})

	http.HandleFunc("/get-token", func(w http.ResponseWriter, r *http.Request) {

		log.Println("Received request for /get-token")

		racf := r.URL.Query().Get("racf")

		token, err := rdb.Get(ctx, racf).Result()

		if err == redis.Nil {
			http.Error(w, "Token not found", http.StatusNotFound)
			return
		} else if err != nil {

			log.Printf("Error getting token from Redis: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Println("Token retrieved successfully")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"token": token,
		})
	})
	handler := cors.Default().Handler(http.DefaultServeMux)

	log.Println("API is running on port 8083")
	log.Fatal(http.ListenAndServe(":8083", handler))
}
