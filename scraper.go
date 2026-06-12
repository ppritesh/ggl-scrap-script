package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

type Scraper struct {
	Headless bool
	Timeout  time.Duration
	Delay    time.Duration
}

func (s *Scraper) ScrapeQuery(ctx context.Context, query string, limit int) ([]Place, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", s.Headless),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	browserCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	if s.Timeout > 0 {
		var timeoutCancel context.CancelFunc
		browserCtx, timeoutCancel = context.WithTimeout(browserCtx, s.Timeout)
		defer timeoutCancel()
	}

	searchURL := "https://www.google.com/maps/search/" + strings.ReplaceAll(query, " ", "+")
	log.Printf("searching: %q", query)

	if err := chromedp.Run(browserCtx,
		chromedp.Navigate(searchURL),
		chromedp.Sleep(3*time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			return dismissConsent(ctx)
		}),
		chromedp.Sleep(2*time.Second),
	); err != nil {
		return nil, fmt.Errorf("open search page: %w", err)
	}

	if err := scrollResultsFeed(browserCtx, limit); err != nil {
		log.Printf("warning: scroll feed: %v", err)
	}

	links, err := collectResultLinks(browserCtx, limit)
	if err != nil {
		return nil, fmt.Errorf("collect results: %w", err)
	}

	log.Printf("found %d listings for %q", len(links), query)

	places := make([]Place, 0, len(links))
	for i, link := range links {
		log.Printf("  [%d/%d] %s", i+1, len(links), link.Name)

		place, err := scrapePlaceDetail(browserCtx, query, link)
		if err != nil {
			log.Printf("warning: skip %q: %v", link.Name, err)
			continue
		}

		places = append(places, place)

		if s.Delay > 0 && i < len(links)-1 {
			time.Sleep(s.Delay)
		}
	}

	return places, nil
}

type resultLink struct {
	Name string
	URL  string
}

func dismissConsent(ctx context.Context) error {
	selectors := []string{
		`//button[contains(., "Accept all")]`,
		`//button[contains(., "I agree")]`,
		`//button[contains(., "Accept")]`,
	}

	for _, sel := range selectors {
		var visible bool
		if err := chromedp.Run(ctx,
			chromedp.EvaluateAsDevTools(fmt.Sprintf(`document.evaluate(%q, document, null, XPathResult.FIRST_ORDERED_NODE_TYPE, null).singleNodeValue !== null`, sel), &visible),
		); err != nil || !visible {
			continue
		}

		_ = chromedp.Run(ctx, chromedp.Click(sel, chromedp.BySearch))
		time.Sleep(1 * time.Second)
		return nil
	}

	return nil
}

func scrollResultsFeed(ctx context.Context, limit int) error {
	scrolls := limit/5 + 3
	if scrolls > 15 {
		scrolls = 15
	}

	for i := 0; i < scrolls; i++ {
		var atEnd bool
		err := chromedp.Run(ctx, chromedp.EvaluateAsDevTools(`
			(() => {
				const feed = document.querySelector('div[role="feed"]');
				if (!feed) return true;
				const before = feed.scrollTop;
				feed.scrollTop = feed.scrollHeight;
				return feed.scrollTop === before;
			})()
		`, &atEnd))
		if err != nil {
			return err
		}

		time.Sleep(1500 * time.Millisecond)

		if atEnd {
			break
		}
	}

	return nil
}

func collectResultLinks(ctx context.Context, limit int) ([]resultLink, error) {
	var raw []map[string]string
	err := chromedp.Run(ctx, chromedp.EvaluateAsDevTools(fmt.Sprintf(`
		(() => {
			const cards = Array.from(document.querySelectorAll('div[role="feed"] a[href*="/maps/place/"]'));
			const seen = new Set();
			const out = [];

			for (const a of cards) {
				const href = a.href || '';
				if (!href || seen.has(href)) continue;

				let name = (a.getAttribute('aria-label') || '').trim();
				if (!name) {
					const heading = a.querySelector('.fontHeadlineSmall, .qBF1Pd, [class*="fontHeadline"]');
					name = heading ? heading.textContent.trim() : '';
				}
				if (!name) continue;

				seen.add(href);
				out.push({ name, url: href });
				if (out.length >= %d) break;
			}

			return out;
		})()
	`, limit), &raw))
	if err != nil {
		return nil, err
	}

	links := make([]resultLink, 0, len(raw))
	for _, item := range raw {
		links = append(links, resultLink{
			Name: item["name"],
			URL:  item["url"],
		})
	}

	return links, nil
}

func scrapePlaceDetail(ctx context.Context, query string, link resultLink) (Place, error) {
	place := Place{
		Query:   query,
		Name:    link.Name,
		MapsURL: link.URL,
	}

	var details map[string]string
	err := chromedp.Run(ctx,
		chromedp.Navigate(link.URL),
		chromedp.Sleep(2500*time.Millisecond),
		chromedp.EvaluateAsDevTools(extractPlaceDetailsJS, &details),
	)
	if err != nil {
		return place, err
	}

	if details["name"] != "" {
		place.Name = details["name"]
	}
	place.Address = details["address"]
	place.Phone = details["phone"]
	place.Website = details["website"]
	place.Email = details["email"]
	place.Rating = details["rating"]
	place.Reviews = details["reviews"]
	place.Category = details["category"]
	place.Status = details["status"]
	place.Hours = details["hours"]
	place.PlusCode = details["plus_code"]
	place.Latitude = details["latitude"]
	place.Longitude = details["longitude"]
	place.PlaceID = details["place_id"]
	place.GoogleID = details["google_id"]
	place.PriceLevel = details["price_level"]
	place.Description = details["description"]
	place.Amenities = details["amenities"]
	place.MenuURL = details["menu_url"]
	place.BookingURL = details["booking_url"]
	place.OrderURL = details["order_url"]

	return place, nil
}

const extractPlaceDetailsJS = `
(() => {
	const text = sel => {
		const el = document.querySelector(sel);
		return el ? el.textContent.trim() : '';
	};
	const attr = (sel, name) => {
		const el = document.querySelector(sel);
		return el ? (el.getAttribute(name) || '').trim() : '';
	};
	const clean = (value, prefix) => {
		let v = (value || '').trim();
		v = v.replace(new RegExp('^' + prefix + '\\s*:?\\s*', 'i'), '');
		v = v.replace(/^:\s*/, '');
		return v.trim();
	};

	const url = window.location.href;

	let latitude = '';
	let longitude = '';
	const coordMatch = url.match(/!3d(-?\d+\.?\d*)!4d(-?\d+\.?\d*)/);
	if (coordMatch) {
		latitude = coordMatch[1];
		longitude = coordMatch[2];
	} else {
		const atMatch = url.match(/@(-?\d+\.?\d*),(-?\d+\.?\d*)/);
		if (atMatch) {
			latitude = atMatch[1];
			longitude = atMatch[2];
		}
	}

	let place_id = '';
	const placeMatch = url.match(/!1s(0x[0-9a-f]+:0x[0-9a-f]+)/i);
	if (placeMatch) place_id = placeMatch[1];

	let google_id = '';
	const cidMatch = url.match(/!16s([^!&]+)/);
	if (cidMatch) google_id = decodeURIComponent(cidMatch[1]);

	const name = text('h1.DUwDvf') || text('h1[class*="fontHeadline"]') || text('h1');

	const ratingLabel = attr('span[role="img"][aria-label*="stars"]', 'aria-label')
		|| attr('span[aria-label*="star"]', 'aria-label');
	let rating = '';
	let reviews = '';
	if (ratingLabel) {
		const m = ratingLabel.match(/([\d.]+)\s*stars?/i);
		if (m) rating = m[1];
		const r = ratingLabel.match(/([\d,]+)\s*reviews?/i);
		if (r) reviews = r[1].replace(/,/g, '');
	}
	if (!reviews) {
		const reviewText = text('button[aria-label*="reviews"]') || text('span[aria-label*="reviews"]');
		const r = reviewText.match(/([\d,]+)/);
		if (r) reviews = r[1].replace(/,/g, '');
	}

	const addressBtn = document.querySelector('button[data-item-id="address"]');
	const phoneBtn = document.querySelector('button[data-item-id^="phone"]');
	const websiteLink = document.querySelector('a[data-item-id="authority"]');
	const plusBtn = document.querySelector('button[data-item-id="oloc"]');
	const hoursBtn = document.querySelector('button[data-item-id="oh"]');
	const menuLink = document.querySelector('a[data-item-id="menu"]');
	const bookingLink = document.querySelector('a[data-item-id="reservations"]');
	const orderLink = document.querySelector('a[data-item-id="action:3"]');
	const mailLink = document.querySelector('a[href^="mailto:"]');

	const address = addressBtn
		? clean(addressBtn.getAttribute('aria-label') || addressBtn.textContent || '', 'Address')
		: '';
	const phone = phoneBtn
		? clean(phoneBtn.getAttribute('aria-label') || phoneBtn.textContent || '', 'Phone')
		: '';
	const website = websiteLink ? (websiteLink.href || websiteLink.textContent || '').trim() : '';
	const email = mailLink ? (mailLink.href || '').replace(/^mailto:/i, '').split('?')[0].trim() : '';
	const plus_code = plusBtn
		? clean(plusBtn.getAttribute('aria-label') || plusBtn.textContent || '', 'Plus code')
		: '';

	let status = '';
	let hours = '';
	if (hoursBtn) {
		const label = hoursBtn.getAttribute('aria-label') || hoursBtn.textContent || '';
		const statusMatch = label.match(/\b(Open|Closed)\b/i);
		if (statusMatch) status = statusMatch[1];
		hours = clean(label, 'Hours').replace(/^[·•]\s*/, '').trim();
	}

	const categoryBtn = document.querySelector('button[jsaction*="category"]');
	const category = categoryBtn ? categoryBtn.textContent.trim() : text('button.DkEaL');

	let price_level = '';
	const priceEl = document.querySelector('[aria-label*="Price"], [aria-label*="price"]');
	if (priceEl) {
		price_level = clean(priceEl.getAttribute('aria-label') || priceEl.textContent || '', 'Price');
	}
	if (!price_level) {
		const priceText = text('.mgr77e');
		if (priceText && /^[₹$€£]{1,4}$/.test(priceText)) price_level = priceText;
	}

	let description = '';
	const aboutCandidates = [
		document.querySelector('[data-section-id="overview"]'),
		document.querySelector('.PYvSYb'),
		document.querySelector('.WeS02d'),
	];
	for (const el of aboutCandidates) {
		if (el && el.textContent.trim()) {
			description = el.textContent.trim().replace(/\s+/g, ' ').slice(0, 1000);
			break;
		}
	}

	const amenitySet = new Set();
	const skipAmenity = new Set([address, phone, plus_code, name, category].filter(Boolean));
	const amenitySelectors = [
		'[class*="Io6YTe"]',
		'.LT79uc span',
		'[role="group"] span',
		'button[data-item-id^="service"]',
	];
	for (const sel of amenitySelectors) {
		for (const el of document.querySelectorAll(sel)) {
			const value = (el.textContent || el.getAttribute('aria-label') || '').trim();
			if (!value || value.length <= 1 || value.length >= 80) continue;
			if (skipAmenity.has(value)) continue;
			if (/^[\d\s+\-()]+$/.test(value)) continue;
			if (/^[A-Z0-9+]{4,}$/.test(value)) continue;
			if (/\.(com|in|org|net)\b/i.test(value)) continue;
			amenitySet.add(value);
		}
	}
	const amenities = Array.from(amenitySet).slice(0, 30).join('; ');

	const menu_url = menuLink ? menuLink.href : '';
	const booking_url = bookingLink ? bookingLink.href : '';
	const order_url = orderLink ? orderLink.href : '';

	return {
		name, address, phone, website, email, rating, reviews, category,
		status, hours, plus_code, latitude, longitude, place_id, google_id,
		price_level, description, amenities, menu_url, booking_url, order_url,
	};
})()
`
