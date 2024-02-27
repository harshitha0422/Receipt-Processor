package receipts

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/backend/processortest/models"
	"github.com/backend/processortest/utils"
	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
)

// ProcessReceipts handles the processing of receipts.
func ProcessReceipts(w http.ResponseWriter, r *http.Request, c *cache.Cache, processedReceiptsCache *cache.Cache) {
	var receipt models.Receipt
	err := json.NewDecoder(r.Body).Decode(&receipt)
	if err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Validate receipt against the schema
	if err := utils.ValidateReceipt(receipt); err != nil {
		http.Error(w, fmt.Sprintf("Invalid receipt: %s", err), http.StatusBadRequest)
		return
	}

	receiptID := utils.GenerateReceiptID(c)

	// Validate the generated ID format
	if err := utils.ValidateID(receiptID); err != nil {
		http.Error(w, fmt.Sprintf("Invalid generated ID: %s", err), http.StatusInternalServerError)
		return
	}

	// Mark the receipt ID as used
	c.Set(receiptID, receipt, cache.DefaultExpiration)

	if receiptID == "" {
		// Receipt ID generation failed, return an error
		http.Error(w, "Failed to generrate unique receipt ID", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"id": receiptID})
}

// GetPoints handles the retrieval of points for a given receipt ID.
func GetPoints(w http.ResponseWriter, r *http.Request, c *cache.Cache, processedReceiptsCache *cache.Cache) {
	params := mux.Vars(r)
	receiptID := params["id"]
	fmt.Println(receiptID)

	// Validate the receipt ID format
	if err := utils.ValidateID(receiptID); err != nil {
		http.Error(w, fmt.Sprintf("Invalid receipt ID format: %s", err), http.StatusBadRequest)
		return
	}

	// Check if the receipt ID has been processed for points retrieval
	if _, exists := processedReceiptsCache.Get(receiptID); exists {
		// Receipt has been processed before, return a response indicating it
		json.NewEncoder(w).Encode(models.PointsResponse{Points: 0})
		return
	}

	// Get the receipt from the cache
	receipt, exists := c.Get(receiptID)
	if !exists {
		// Receipt not found, return an error
		http.Error(w, "No Receipt found for that id", http.StatusNotFound)
		return
	}

	// Calculate points
	points, err := utils.CalculatePoints(receipt.(models.Receipt))
	if err != nil {
		http.Error(w, "Error calculating points", http.StatusInternalServerError)
		return
	}

	// Mark the receipt ID as processed for points retrieval
	processedReceiptsCache.Set(receiptID, true, cache.DefaultExpiration)

	// Return the points
	json.NewEncoder(w).Encode(models.PointsResponse{Points: points})
}
