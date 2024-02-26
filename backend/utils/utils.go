package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"

	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/backend/processortest/models"
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

func GenerateReceiptID(receipt models.Receipt, c *cache.Cache) (string, error) {
	// Sort item descriptions for consistency
	sort.Slice(receipt.Items, func(i, j int) bool {
		return receipt.Items[i].ShortDescription < receipt.Items[j].ShortDescription
	})

	// Create a deterministic hash of relevant information
	hasher := sha256.New()
	_, err := fmt.Fprintf(hasher, "%s%s%s", receipt.Retailer, receipt.PurchaseDate, receipt.PurchaseTime)
	if err != nil {
		return "", fmt.Errorf("failed to write retailer, date, and time to hasher: %w", err)
	}

	for _, item := range receipt.Items {
		_, err := fmt.Fprintf(hasher, "%s%s", item.ShortDescription, item.Price)
		if err != nil {
			return "", fmt.Errorf("failed to write item description and price to hasher: %w", err)
		}
	}

	_, err = fmt.Fprintf(hasher, "%s", receipt.Total)
	if err != nil {
		return "", fmt.Errorf("failed to write total to hasher: %w", err)
	}

	receiptID := hex.EncodeToString(hasher.Sum(nil))

	// Check if the receipt ID has been used before
	if _, exists := c.Get(receiptID); exists {
		// Receipt ID has been used before, return an error
		return receiptID, fmt.Errorf("The same receipt is passed")
	}

	// Mark the receipt ID as used
	c.Set(receiptID, receipt, cache.DefaultExpiration)

	return receiptID, nil
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

	// Validate purchase date format
	if _, err := time.Parse("2006-01-02", receipt.PurchaseDate); err != nil {
		return errors.New("Invalid PurchaseDate format")
	}

	// Validate purchase time format
	if _, err := time.Parse("15:04", receipt.PurchaseTime); err != nil {
		return errors.New("Invalid PurchaseTime format")
	}

	// Validate total amount format
	if matched, _ := regexp.MatchString("^\\d+\\.\\d{2}$", receipt.Total); !matched {
		return errors.New("Invalid Total format")
	}

	// Validate each item
	for _, item := range receipt.Items {
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
