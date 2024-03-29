package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// MessageData holds the extracted price and description.
type MessageData struct {
	Price       float64
	Description string // Description is now optional
}

// ExtractPriceAndDescription attempts to extract a price and description from a message.
func ExtractPriceAndDescription(message string) (MessageData, error) {
	var data MessageData
	parts := strings.Fields(message) // Split the message into parts
	priceFound := false

	for _, part := range parts {
		// Try to parse the part as a float, considering both '.' and ',' as decimal separators.
		price, err := strconv.ParseFloat(strings.Replace(part, ",", ".", 1), 64)
		if err == nil {
			data.Price = price
			priceFound = true
		} else {
			// Part of the description
			if len(data.Description) > 0 {
				data.Description += " "
			}
			data.Description += part
		}
	}

	if !priceFound {
		return data, fmt.Errorf("no price found in message")
	}

	// No error is returned if a description is not found
	return data, nil
}

func (md MessageData) String() string {
	return fmt.Sprintf("Added %.2f (%s)", md.Price, md.Description)
}

func (md MessageData) DescriptionWithFallback() string {
	if len(md.Description) > 0 {
		return md.Description
	}
	return "?"
}

func GetNumberFromString(s string) string {
	re := regexp.MustCompile(`\d{1,3}(?:,\d{3})*(?:\.\d+)?|\d+(?:,\d+)?`)
	price := re.FindString(s)
	return price
}

// removeFirstSeparator removes the first comma or dot from the left in a string
// that contains both a comma and a dot, and returns the modified string.
func removeFirstSeparator(s string) string {
	commaIndex := strings.Index(s, ",")
	dotIndex := strings.Index(s, ".")

	// Check if both comma and dot are present
	if commaIndex != -1 && dotIndex != -1 {
		// Determine which separator comes first and remove it
		if commaIndex < dotIndex {
			// Remove the first comma
			return s[:commaIndex] + s[commaIndex+1:]
		} else {
			// Remove the first dot
			return s[:dotIndex] + s[dotIndex+1:]
		}
	}

	// Return the original string if both a comma and dot are not present
	return s
}

func GetPriceFromString(s string) float64 {
	price := GetNumberFromString(s)
	if price == "" {
		return -1
	}
	price = removeFirstSeparator(price)
	price = strings.Replace(price, ",", ".", 1)
	p, err := strconv.ParseFloat(price, 64)
	if err != nil {
		return -1
	}
	return p
}
