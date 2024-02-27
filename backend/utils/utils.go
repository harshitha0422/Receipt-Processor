package utils

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/backend/processortest/models"
	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
)

var receiptsMutex sync.Mutex

// CalculatePoints calculates the points for a given receipt based on the specified rules.
func CalculatePoints(receipt models.Receipt) (int, error) {
	receiptsMutex.Lock()
	defer receiptsMutex.Unlock()

	points := 0

	// Rule: One point for every alphanumeric character in the retailer name.
	points += len(strings.ReplaceAll(receipt.Retailer, " ", ""))

	// Rule: 50 points if the total is a round dollar amount with no cents.
	if isRoundDollarAmount(receipt.Total) {
		points += 50
	}

	// Rule: 25 points if the total is a multiple of 0.25.
	if isMultipleOfQuarter(receipt.Total) {
		points += 25
	}

	// Rule: 5 points for every two items on the receipt.
	points += (len(receipt.Items) / 2) * 5

	// Rule: If the trimmed length of the item description is a multiple of 3,
	// multiply the price by 0.2 and round up to the nearest integer.
	for _, item := range receipt.Items {
		trimmedLength := len(strings.TrimSpace(item.ShortDescription))
		if trimmedLength%3 == 0 {
			price, err := strconv.ParseFloat(item.Price, 64)
			if err != nil {
				// Handle parsing error
				return 0, fmt.Errorf("failed to parse price: %w", err)
				// continue
			}
			points += int(math.Ceil(price * 0.2))
		}
	}

	// Rule: 6 points if the day in the purchase date is odd.
	purchaseDate, err := time.Parse("2006-01-02", receipt.PurchaseDate)
	if err != nil {
		// Handle date parsing error
		return 0, fmt.Errorf("failed to parse purchase date: %w", err)
	}
	if purchaseDate.Day()%2 == 1 {
		points += 6
	}

	// Rule: 10 points if the time of purchase is after 2:00pm and before 4:00pm.
	purchaseTime, err := time.Parse("15:04", receipt.PurchaseTime)
	if err != nil {
		// Handle time parsing error
		return 0, fmt.Errorf("failed to parse purchase time: %w", err)
	}
	if purchaseTime.After(time.Date(0, 1, 1, 14, 0, 0, 0, time.UTC)) &&
		purchaseTime.Before(time.Date(0, 1, 1, 16, 0, 0, 0, time.UTC)) {
		points += 10
	}

	return points, nil
}

// Helper function to check if the total is a round dollar amount with no cents.
func isRoundDollarAmount(total string) bool {
	// Assuming total is in the format "xx.xx"
	return strings.HasSuffix(total, ".00")
}

// Helper function to check if the total is a multiple of 0.25.
func isMultipleOfQuarter(total string) bool {
	// Assuming total is in the format "xx.xx"
	value := parseTotal(total)
	return math.Mod(value, 0.25) == 0
}

// Helper function to parse the total value as a float64.
func parseTotal(total string) float64 {
	value, err := strconv.ParseFloat(total, 64)
	if err != nil {
		return 0.0
	}
	return value
}

func GenerateReceiptID(c *cache.Cache) string {
	var GenID string

	for {
		// Generate a new UUID
		GenID = uuid.New().String()

		// Check if the generated UUID exists in the cache
		if _, exists := c.Get(GenID); !exists {
			break // Break the loop if the UUID is not in the cache
		}
	}

	return GenID
}

// ValidateReceipt checks if the provided receipt is valid according to the OpenAPI schema.
func ValidateReceipt(receipt models.Receipt) error {
	// Validate required fields
	if receipt.Retailer == "" {
		return errors.New("Retailer is required")
	}
	if receipt.PurchaseDate == "" {
		return errors.New("PurchaseDate is required")
	}
	if receipt.PurchaseTime == "" {
		return errors.New("PurchaseTime is required")
	}
	if receipt.Total == "" {
		return errors.New("Total is required")
	}
	if len(receipt.Items) == 0 {
		return errors.New("At least one item is required")
	}

	// Validate retailer name pattern
	if matched, _ := regexp.MatchString("^[\\w\\s\\-]+$", receipt.Retailer); !matched {
		return errors.New("Retailer name has invalid characters")
	}

	// Validate purchase date format  //priority 1
	if _, err := time.Parse("2006-01-02", receipt.PurchaseDate); err != nil {
		return errors.New("Invalid PurchaseDate format")
	}

	// Validate purchase time format //priority 2
	if _, err := time.Parse("15:04", receipt.PurchaseTime); err != nil {
		return errors.New("Invalid PurchaseTime format")
	}

	// Validate total amount format
	if matched, _ := regexp.MatchString("^\\d+\\.\\d{2}$", receipt.Total); !matched {
		return errors.New("Invalid Total format")
	}

	// Calculate total from item prices
	var totalFromItems float64
	for _, item := range receipt.Items {
		// Validate each item
		if item.ShortDescription == "" {
			return errors.New("Item ShortDescription is required")
		}
		if item.Price == "" {
			return errors.New("Item Price is required")
		}

		// Validate item short description pattern
		if matched, _ := regexp.MatchString("^[\\w\\s\\-]+$", item.ShortDescription); !matched {
			return errors.New("Item ShortDescription has invalid characters")
		}

		// Validate item price format
		if matched, _ := regexp.MatchString("^\\d+\\.\\d{2}$", item.Price); !matched {
			return errors.New("Item Price has an invalid format")
		}

		// Parse item price to float64 and add it to total
		price, err := strconv.ParseFloat(item.Price, 64)
		if err != nil {
			return errors.New("Failed to parse item price")
		}
		totalFromItems += price
	}

	// Parse total from string to float64
	total, err := strconv.ParseFloat(receipt.Total, 64)
	if err != nil {
		return errors.New("Failed to parse total amount")
	}

	// Compare total from items with provided total
	epsilon := 0.0001 // Adjust based on your precision requirements
	if math.Abs(totalFromItems-total) > epsilon {
		return errors.New("Total amount does not match the sum of item prices")
	}

	return nil
}

// ValidateID checks if the provided ID matches the expected pattern.
func ValidateID(id string) error {
	// Validate ID pattern
	if matched, _ := regexp.MatchString("^\\S+$", id); !matched {
		return errors.New("Invalid ID format")
	}
	return nil
}
