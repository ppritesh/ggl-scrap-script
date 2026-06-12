package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// multiString accepts repeated flags and comma-separated values.
type multiString []string

func (m *multiString) String() string {
	return strings.Join(*m, ", ")
}

func (m *multiString) Set(value string) error {
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			*m = append(*m, part)
		}
	}
	return nil
}

func main() {
	var (
		queryParams multiString
		qParams     multiString
		queriesFile string
		output      string
		limit       int
		radiusKm    float64
		headless    bool
		timeoutSec  int
		delaySec    int
	)

	flag.Var(&queryParams, "query", "Search query (repeat flag or comma-separate for multiple)")
	flag.Var(&qParams, "q", "Shorthand for -query")
	flag.StringVar(&queriesFile, "queries", "", "File with one query per line")
	flag.StringVar(&output, "output", "results.csv", "Output CSV file path")
	flag.IntVar(&limit, "limit", 0, "Max results per query (0 = unlimited, scroll all)")
	flag.Float64Var(&radiusKm, "radius", defaultRadiusKm, "Search radius in km from query location")
	flag.BoolVar(&headless, "headless", true, "Run browser in headless mode")
	flag.IntVar(&timeoutSec, "timeout", 600, "Timeout in seconds per query")
	flag.IntVar(&delaySec, "delay", 2, "Delay in seconds between place detail requests")
	flag.Parse()

	queries, err := loadQueries(queryParams, qParams, queriesFile, flag.Args())
	if err != nil {
		log.Fatal(err)
	}
	if len(queries) == 0 {
		fmt.Println("Google Maps scraper — saves business listings to CSV")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println(`  google-maps-scraper.exe -q "carpenter in daman"`)
		fmt.Println(`  google-maps-scraper.exe -q "carpenter in daman" -radius 15`)
		fmt.Println(`  google-maps-scraper.exe -q "carpenter in daman" -q "steel industry in vapi"`)
		fmt.Println(`  google-maps-scraper.exe -q "carpenter in daman, steel industry in vapi"`)
		fmt.Println(`  google-maps-scraper.exe "carpenter in daman" "steel industry in vapi"`)
		fmt.Println(`  google-maps-scraper.exe -queries queries.txt -output results.csv`)
		fmt.Println()
		fmt.Println("Flags:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	scraper := &Scraper{
		Headless: headless,
		Timeout:  time.Duration(timeoutSec) * time.Second,
		Delay:    time.Duration(delaySec) * time.Second,
		RadiusKm: radiusKm,
	}

	ctx := context.Background()
	allPlaces := make([]Place, 0)

	for i, query := range queries {
		log.Printf("=== query %d/%d: %q ===", i+1, len(queries), query)

		places, err := scraper.ScrapeQuery(ctx, query, limit)
		if err != nil {
			log.Printf("error scraping %q: %v", query, err)
			continue
		}

		allPlaces = append(allPlaces, places...)
		log.Printf("collected %d places for %q", len(places), query)
	}

	if len(allPlaces) == 0 {
		log.Fatal("no results scraped")
	}

	if err := writeCSV(output, allPlaces); err != nil {
		log.Fatalf("write csv: %v", err)
	}

	log.Printf("saved %d rows to %s", len(allPlaces), output)
}

func loadQueries(queryParams, qParams multiString, file string, args []string) ([]string, error) {
	seen := make(map[string]struct{})
	var queries []string

	add := func(q string) {
		q = strings.TrimSpace(q)
		if q == "" {
			return
		}
		if _, ok := seen[q]; ok {
			return
		}
		seen[q] = struct{}{}
		queries = append(queries, q)
	}

	for _, q := range queryParams {
		add(q)
	}
	for _, q := range qParams {
		add(q)
	}

	for _, arg := range args {
		for _, part := range strings.Split(arg, ",") {
			add(part)
		}
	}

	if file != "" {
		f, err := os.Open(file)
		if err != nil {
			return nil, fmt.Errorf("open queries file: %w", err)
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			add(line)
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("read queries file: %w", err)
		}
	}

	return queries, nil
}
