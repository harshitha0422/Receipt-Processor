package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/backend/processortest/routes"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
)

func main() {

	c := cache.New(720*time.Minute, 720*time.Minute)
	processedReceiptsCache := cache.New(cache.NoExpiration, cache.NoExpiration)
	router := mux.NewRouter()

	routes.RegisterRoutes(router, c, processedReceiptsCache)

	http.Handle("/", router)

	err := http.ListenAndServe(":9018", router)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}

}
