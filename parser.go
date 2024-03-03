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
	re := regexp.MustCompile(`(\d+[,.]?\d*)`)
	price := re.FindString(s)
	// if price == "" {
	// 	return ""
	// }
	// split := strings.Split(s, price)
	// other := ""
	// for _, s := range split {
	// 	other += strings.TrimSpace(s)
	// 	other += " "
	// }
	return price
}

func GetPriceFromString(s string) float64 {
	price := GetNumberFromString(s)
	if price == "" {
		return -1
	}
	price = strings.Replace(price, ",", ".", 1)
	p, err := strconv.ParseFloat(price, 64)
	if err != nil {
		return -1
	}
	return p
}
