package utils

import (
	"testing"

	"github.com/backend/processortest/models"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
)

func TestCalculatePoints_FailureCase(t *testing.T) {
	// Prepare a receipt that should result in a failure
	receipt := models.Receipt{
		Retailer: "Invalid Retailer %$@", // Retailer with non-alphanumeric characters
		Total:    "15.75",                // Total not a round dollar amount
		Items: []models.Item{
			{ShortDescription: "Item 1", Price: "9.99"},
			{ShortDescription: "Item 2", Price: "5.76"},
		},
		PurchaseDate: "2022-02-23",
		PurchaseTime: "14:30",
	}

	// Call CalculatePoints function
	points, _ := CalculatePoints(receipt)

	// Assert that there is an error and points are 0

	assert.NotEqual(t, 0, points)

}

func TestCalculatePoints_SuccessCase(t *testing.T) {
	// Prepare a receipt that should result in success
	receipt := models.Receipt{
		Retailer: "Valid Retailer", // Retailer with alphanumeric characters
		Total:    "20.00",          // Total is a round dollar amount
		Items: []models.Item{
			{ShortDescription: "Item 1", Price: "10.00"},
			{ShortDescription: "Item 2", Price: "10.00"},
		},
		PurchaseDate: "2022-02-23",
		PurchaseTime: "15:30",
	}

	// Call CalculatePoints function
	points, _ := CalculatePoints(receipt)

	// Assert that there is no error and points are calculated as expected

	assert.Equal(t, 113, points)

}

func TestCalculatePoints(t *testing.T) {
	// Test case for a valid receipt
	receipt := models.Receipt{
		Retailer:     "Test Retailer",
		PurchaseDate: "2022-02-28",
		PurchaseTime: "15:30",
		Total:        "100.00",
		Items: []models.Item{
			{ShortDescription: "Item 1", Price: "50.00"},
			{ShortDescription: "Item 2", Price: "25.00"},
		},
	}
	points, err := CalculatePoints(receipt)
	assert.NoErrorf(t, err, "Unexpected error calculating points: %v", err)
	assert.Equalf(t, 63, points, "Unexpected points value: expected %d, got %d", 63, points)

	// Test case for an invalid receipt (missing required field)
	invalidReceipt := models.Receipt{}
	_, err = CalculatePoints(invalidReceipt)
	assert.Errorf(t, err, "Expected error for invalid receipt")
	assert.Contains(t, err.Error(), "Retailer is required", "Unexpected error message for missing retailer")
}

func TestGenerateReceiptID(t *testing.T) {
	// Test case for a valid receipt ID generation
	receipt := models.Receipt{
		Retailer:     "Test Retailer",
		PurchaseDate: "2022-02-28",
		PurchaseTime: "15:30",
		Total:        "100.00",
		Items: []models.Item{
			{ShortDescription: "Item 1", Price: "50.00"},
			{ShortDescription: "Item 2", Price: "25.00"},
		},
	}
	cache := cache.New(cache.NoExpiration, cache.NoExpiration)
	receiptID, err := GenerateReceiptID(receipt, cache)
	assert.NoErrorf(t, err, "Unexpected error generating receipt ID: %v", err)
	assert.NotEmpty(t, receiptID, "Generated receipt ID is empty")

	// Test case for an invalid receipt ID generation (duplicate receipt)
	_, err = GenerateReceiptID(receipt, cache)
	assert.Errorf(t, err, "Expected error for duplicate receipt")
	assert.Contains(t, err.Error(), "The same receipt is passed", "Unexpected error message for duplicate receipt")
}

func TestValidateReceipt(t *testing.T) {
	// Test case for a valid receipt
	receipt := models.Receipt{
		Retailer:     "Test Retailer",
		PurchaseDate: "2022-02-28",
		PurchaseTime: "15:30",
		Total:        "100.00",
		Items: []models.Item{
			{ShortDescription: "Item 1", Price: "50.00"},
			{ShortDescription: "Item 2", Price: "25.00"},
		},
	}
	err := ValidateReceipt(receipt)
	assert.NoErrorf(t, err, "Unexpected error validating receipt: %v", err)

	// Test case for an invalid receipt (missing required field)
	invalidReceipt := models.Receipt{}
	err = ValidateReceipt(invalidReceipt)
	assert.Errorf(t, err, "Expected error for invalid receipt")
	assert.Contains(t, err.Error(), "Retailer is required", "Unexpected error message for missing retailer")
}

func TestValidateID(t *testing.T) {
	// Test case for a valid ID
	validID := "abc123"
	err := ValidateID(validID)
	assert.NoErrorf(t, err, "Unexpected error validating ID: %v", err)

	// Test case for an invalid ID (contains space)
	invalidID := "abc 123"
	err = ValidateID(invalidID)
	assert.Errorf(t, err, "Expected error for invalid ID")
	assert.Contains(t, err.Error(), "Invalid ID format", "Unexpected error message for invalid ID format")
}
