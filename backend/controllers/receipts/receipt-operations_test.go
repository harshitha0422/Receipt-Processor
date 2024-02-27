package receipts

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/backend/processortest/models"

	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
)

func TestProcessReceiptsAndRetrievePoints(t *testing.T) {
	// Initialize caches
	c := cache.New(cache.NoExpiration, cache.NoExpiration)
	processedReceiptsCache := cache.New(cache.NoExpiration, cache.NoExpiration)

	// Create a new router
	r := mux.NewRouter()

	// Attach your handlers to the router
	r.HandleFunc("/process", func(w http.ResponseWriter, r *http.Request) {
		ProcessReceipts(w, r, c, processedReceiptsCache)
	}).Methods("POST")

	r.HandleFunc("/{id}/points", func(w http.ResponseWriter, r *http.Request) {
		GetPoints(w, r, c, processedReceiptsCache)
	}).Methods("GET")

	// Test Case 1: Test successful processing and points retrieval
	t.Run("TestSuccessfulProcessingAndPointsRetrieval", func(t *testing.T) {
		// Prepare a valid receipt payload
		validReceipt := models.Receipt{
			Retailer:     "Test Retailer",
			PurchaseDate: "2022-02-28",
			PurchaseTime: "12:34",
			Total:        "25.99",
			Items: []models.Item{
				{ShortDescription: "Item 1", Price: "10.99"},
				{ShortDescription: "Item 2", Price: "15.00"},
			},
		}

		// Convert the receipt to JSON
		receiptJSON, err := json.Marshal(validReceipt)
		if err != nil {
			t.Fatal(err)
		}

		// Simulate a POST request to process the receipt
		reqProcess := httptest.NewRequest("POST", "/process", bytes.NewBuffer(receiptJSON))
		reqProcess.Header.Set("Content-Type", "application/json")
		recProcess := httptest.NewRecorder()

		r.ServeHTTP(recProcess, reqProcess)

		// Check if the processing was successful
		if recProcess.Code != http.StatusOK {
			t.Fatalf("Expected status %d; got %d", http.StatusOK, recProcess.Code)
		}

		// Decode the response to get the generated ID
		var processResponse map[string]string
		if err := json.Unmarshal(recProcess.Body.Bytes(), &processResponse); err != nil {
			t.Fatal(err)
		}

		// Extract the generated receipt ID
		receiptID := processResponse["id"]

		// Simulate a GET request to retrieve points using the generated ID
		reqPoints := httptest.NewRequest("GET", "/"+receiptID+"/points", nil)
		recPoints := httptest.NewRecorder()

		r.ServeHTTP(recPoints, reqPoints)

		// Check if points retrieval was successful
		if recPoints.Code != http.StatusOK {
			t.Fatalf("Expected status %d; got %d", http.StatusOK, recPoints.Code)
		}

		// Decode the response to get the points
		var pointsResponse models.PointsResponse
		if err := json.Unmarshal(recPoints.Body.Bytes(), &pointsResponse); err != nil {
			t.Fatal(err)
		}

		// Check if the points match the expected value
		expectedPoints := 23
		if pointsResponse.Points != expectedPoints {
			t.Fatalf("Expected points %d; got %d", expectedPoints, pointsResponse.Points)
		}
	})

}

func TestGetPoints(t *testing.T) {
	// Initialize caches
	c := cache.New(cache.NoExpiration, cache.NoExpiration)
	processedReceiptsCache := cache.New(cache.NoExpiration, cache.NoExpiration)

	// Create a new router
	r := mux.NewRouter()
	r.HandleFunc("/{id}/points", func(w http.ResponseWriter, r *http.Request) {
		GetPoints(w, r, c, processedReceiptsCache)
	}).Methods("GET")

	// Test Case 1: Test points retrieval for a processed receipt
	t.Run("TestPointsRetrievalForProcessedReceipt", func(t *testing.T) {
		// Add a processed receipt to the cache
		receiptID := "abc123"
		processedReceiptsCache.Set(receiptID, true, cache.DefaultExpiration)

		// Simulate a GET request to retrieve points using the processed ID
		req := httptest.NewRequest("GET", "/"+receiptID+"/points", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		// Check if points retrieval was successful
		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status %d; got %d", http.StatusOK, rec.Code)
		}

		// Decode the response to get the points
		var pointsResponse models.PointsResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &pointsResponse); err != nil {
			t.Fatal(err)
		}

		// Check if the points match the expected value for a processed receipt
		expectedPoints := 0
		if pointsResponse.Points != expectedPoints {
			t.Fatalf("Expected points %d; got %d", expectedPoints, pointsResponse.Points)
		}
	})

	// Test Case 2: Test points retrieval for an unprocessed receipt
	t.Run("TestPointsRetrievalForUnprocessedReceipt", func(t *testing.T) {
		// Simulate a GET request to retrieve points using an unprocessed ID
		req := httptest.NewRequest("GET", "/unknownID/points", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		// Check if the status is Not Found for an unknown receipt ID
		if rec.Code != http.StatusNotFound {
			t.Fatalf("Expected status %d; got %d", http.StatusNotFound, rec.Code)
		}
	})
}
