# Google Maps Scraper (Go)

A Go CLI tool that scrapes business listings from **Google Maps** and exports them to **CSV**. Built for local lead generation searches like `carpenter in daman` or `steel industry in vapi`.

Uses [chromedp](https://github.com/chromedp/chromedp) (headless Chrome) to load Google Maps, scroll through all results, open each listing, and save available business data.

## Features

- Search by keyword + location (natural language queries)
- **Unlimited results** by default (scrolls until Google Maps has no more listings)
- **15 km radius** filter from the query location (configurable)
- Multiple queries in one run (flags, comma-separated, file, or positional args)
- Exports 24 fields including name, address, phone, website, rating, coordinates, and more
- Headless or visible browser mode

## Requirements

- [Go](https://go.dev/dl/) 1.23+
- [Google Chrome](https://www.google.com/chrome/) (used automatically by chromedp)

## Installation

### Clone the repository

```bash
git clone https://github.com/ppritesh/ggl-scrap-script.git
cd ggl-scrap-script
```

### Build

```bash
go build -o google-maps-scraper .
```

On Windows:

```powershell
go build -o google-maps-scraper.exe .
```

### Or run without building

```bash
go run . -q "carpenter in daman"
```

## Quick start

```bash
./google-maps-scraper -q "carpenter in daman" -output results.csv
```

```bash
./google-maps-scraper -q "steel industry in vapi" -radius 15 -output results.csv
```

## Usage

### Single query

```bash
google-maps-scraper -q "carpenter in daman"
```

### Multiple queries

Repeat the flag:

```bash
google-maps-scraper -q "carpenter in daman" -q "steel industry in vapi"
```

Comma-separated in one flag:

```bash
google-maps-scraper -q "carpenter in daman, steel industry in vapi"
```

Positional arguments (no flag):

```bash
google-maps-scraper "carpenter in daman" "steel industry in vapi"
```

### Queries from a file

Create a text file with one query per line (see `queries.example.txt`):

```
carpenter in daman
steel industry in vapi
```

Run:

```bash
google-maps-scraper -queries queries.example.txt -output results.csv
```

Lines starting with `#` and blank lines are ignored.

## Options

| Flag | Default | Description |
|------|---------|-------------|
| `-q`, `-query` | — | Search query (repeat or comma-separate for multiple) |
| `-queries` | — | Path to file with one query per line |
| `-output` | `results.csv` | Output CSV file path |
| `-limit` | `0` | Max results per query (`0` = unlimited) |
| `-radius` | `15` | Search radius in km from the query location |
| `-headless` | `true` | Run browser in headless mode |
| `-timeout` | `600` | Timeout in seconds per query |
| `-delay` | `2` | Delay in seconds between detail page loads |

### Examples with options

```bash
# All results within 15 km (default)
google-maps-scraper -q "plumber in surat" -output surat-plumbers.csv

# Wider area: 25 km radius
google-maps-scraper -q "hotels in goa" -radius 25

# Cap at 50 results per query
google-maps-scraper -q "restaurants in mumbai" -limit 50

# Debug with visible browser
google-maps-scraper -q "carpenter in daman" -headless=false

# Large scrape: longer timeout (20 minutes)
google-maps-scraper -queries queries.example.txt -timeout 1200
```

## CSV output

Each row is one business. Columns:

| Column | Description |
|--------|-------------|
| `query` | Search query used |
| `name` | Business name |
| `address` | Full address |
| `phone` | Phone number |
| `website` | Website or social link |
| `email` | Email (if listed) |
| `rating` | Star rating |
| `reviews` | Review count |
| `category` | Business type |
| `status` | Open / Closed |
| `hours` | Opening hours |
| `plus_code` | Google Plus Code |
| `latitude` | GPS latitude |
| `longitude` | GPS longitude |
| `place_id` | Google place ID |
| `google_id` | Google entity ID |
| `price_level` | Price range |
| `description` | About / overview text |
| `amenities` | Services and features |
| `menu_url` | Menu link |
| `booking_url` | Reservation link |
| `order_url` | Online order link |
| `distance_km` | Distance from search center (km) |
| `maps_url` | Google Maps link |

Empty cells mean that field was not available on the listing page.

## How it works

1. Opens Google Maps with your search query
2. Detects the map center for the location in the query (e.g. **daman** in `carpenter in daman`)
3. Re-centers the search with a zoom level suited to the `-radius` setting
4. Scrolls the results panel until all listings are loaded (unless `-limit` is set)
5. Opens each listing and extracts available fields
6. Keeps only businesses within the configured radius
7. Writes everything to CSV

## Project structure

```
.
├── main.go          # CLI entry point
├── scraper.go       # Google Maps scraping logic
├── place.go         # Place data model
├── csv.go           # CSV export
├── geo.go           # Radius / distance helpers
├── queries.example.txt
├── go.mod
└── README.md
```

## Troubleshooting

| Issue | Suggestion |
|-------|------------|
| No results | Try `-headless=false` to see what Maps shows; check query spelling |
| Timeout | Increase `-timeout` for large areas or many listings |
| Missing phone/website | Not all listings expose every field on Maps |
| Few results | Increase `-radius` or use a broader location in the query |
| Chrome not found | Install Google Chrome; chromedp needs a Chromium-based browser |

## Notes

- Scraping may be slower for unlimited runs because each listing detail page is opened individually.
- Google Maps HTML can change over time; if fields stop populating, selectors in `scraper.go` may need updates.
- Use responsibly and respect applicable terms of service and local laws regarding data collection.

## License

This project is provided as-is for educational and personal use.
