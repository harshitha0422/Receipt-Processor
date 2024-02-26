package routes

import (
	"net/http"

	controllers "github.com/backend/processortest/controllers/receipts"
	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
)

var RegisterRoutes = func(router *mux.Router, c *cache.Cache, processedReceiptsCache *cache.Cache) {
	router.HandleFunc("/receipts/process", func(w http.ResponseWriter, r *http.Request) {
		controllers.ProcessReceipts(w, r, c, processedReceiptsCache)
	}).Methods("POST")
	router.HandleFunc("/receipts/{id}/points", func(w http.ResponseWriter, r *http.Request) {
		controllers.GetPoints(w, r, c, processedReceiptsCache)
	}).Methods("GET")
}
