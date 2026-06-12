package main

import (
	"encoding/csv"
	"fmt"
	"os"
)

var csvHeaders = []string{
	"query",
	"name",
	"address",
	"phone",
	"website",
	"email",
	"rating",
	"reviews",
	"category",
	"status",
	"hours",
	"plus_code",
	"latitude",
	"longitude",
	"place_id",
	"google_id",
	"price_level",
	"description",
	"amenities",
	"menu_url",
	"booking_url",
	"order_url",
	"maps_url",
}

func writeCSV(path string, places []Place) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create csv: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write(csvHeaders); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	for _, p := range places {
		row := []string{
			p.Query,
			p.Name,
			p.Address,
			p.Phone,
			p.Website,
			p.Email,
			p.Rating,
			p.Reviews,
			p.Category,
			p.Status,
			p.Hours,
			p.PlusCode,
			p.Latitude,
			p.Longitude,
			p.PlaceID,
			p.GoogleID,
			p.PriceLevel,
			p.Description,
			p.Amenities,
			p.MenuURL,
			p.BookingURL,
			p.OrderURL,
			p.MapsURL,
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("write row: %w", err)
		}
	}

	return writer.Error()
}
