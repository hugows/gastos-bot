package main

import (
	"fmt"
	"time"
)

type Stats struct {
	Total      float64
	TotalToday float64
	Count      int
}

var CURRENCY = "R$ "

func (app *App) GetSpendingStats(userName string) (*Stats, error) {
	sid := ""
	if sid = app.FindUserSheet(userName); sid == "" {
		return nil, fmt.Errorf("Spreadsheet not found for user " + userName)
	}

	readRange := "Sheet1!A2:E9999" // skip the header
	resp, err := app.Srv.Spreadsheets.Values.Get(sid, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}

	stats := &Stats{Total: 0, TotalToday: 0, Count: len(resp.Values)}

	if len(resp.Values) > 0 {
		for _, row := range resp.Values {
			spent := GetPriceFromString(row[1].(string))
			if spent == -1 {
				fmt.Println("Error parsing price:", err, row[1].(string))
			} else {
				stats.Total += spent
				parsedTime, err := time.Parse("2006-01-02 15:04:05", row[0].(string))
				if err == nil {
					if parsedTime.Day() == time.Now().Day() && parsedTime.Month() == time.Now().Month() && parsedTime.Year() == time.Now().Year() {
						spent := GetPriceFromString(row[1].(string))
						if spent == -1 {
							fmt.Println("Error parsing price:", err, row[1].(string))
						} else {
							stats.TotalToday += spent
						}
					}
				}
			}
		}
	} else {
		return nil, fmt.Errorf("no data found")
	}
	return stats, nil
}

func (stats *Stats) ToString() string {
	return fmt.Sprintf("Total: %s%.2f (%d compras, média %s%.2f)\nHoje: %s%.2f",
		CURRENCY, stats.Total, stats.Count,
		CURRENCY, stats.Total/float64(stats.Count),
		CURRENCY, stats.TotalToday)
}
