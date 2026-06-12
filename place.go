package main

// Place holds scraped Google Maps business data.
type Place struct {
	Query       string
	Name        string
	Address     string
	Phone       string
	Website     string
	Email       string
	Rating      string
	Reviews     string
	Category    string
	Status      string
	Hours       string
	PlusCode    string
	Latitude    string
	Longitude   string
	PlaceID     string
	GoogleID    string
	PriceLevel  string
	Description string
	Amenities   string
	MenuURL     string
	BookingURL  string
	OrderURL    string
	DistanceKm  string
	MapsURL     string
}
