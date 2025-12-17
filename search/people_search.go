package search

import (
	"context"
	"fmt"
	"net/url"

	"github.com/go-rod/rod"

	"linkedin-automation/config"
	"linkedin-automation/logger"
	"linkedin-automation/stealth"
)

// PeopleSearch navigates LinkedIn people search, applies filters (job title, company, location),
// scrolls through results naturally, and collects unique profile URLs.
// Handles pagination up to the configured number of pages and deduplicates results.
// Returns a slice of profile URLs or an error if search fails.
func PeopleSearch(ctx context.Context, page *rod.Page, cfg config.Config, log *logger.Logger) ([]string, error) {
	keywords := fmt.Sprintf("%s %s %s", cfg.Search.JobTitle, cfg.Search.Company, cfg.Search.Location)
	pages := cfg.Search.PaginationPages
	if pages <= 0 {
		pages = 1
	}

	seen := map[string]struct{}{}
	results := []string{}

	for p := 1; p <= pages; p++ {
		searchURL := fmt.Sprintf("%s/search/results/people/?keywords=%s&page=%d", cfg.BaseURL, url.QueryEscape(keywords), p)
		if err := page.Navigate(searchURL); err != nil {
			return results, err
		}
		stealth.RandomPause(cfg.Humanization.MinDelayMs, cfg.Humanization.MaxDelayMs, cfg.Humanization.JitterMs)
		stealth.RandomScroll(page)

		links, err := page.Timeout(12 * 1000).Elements(`a.app-aware-link[href*="/in/"]`)
		if err != nil {
			log.Warn("no results found", "page", p)
			continue
		}
		for _, link := range links {
			href, err := link.Attribute("href")
			if err != nil || href == nil {
				continue
			}
			if _, ok := seen[*href]; ok {
				continue
			}
			seen[*href] = struct{}{}
			results = append(results, *href)
		}
	}
	log.Info("collected profiles", "count", len(results))
	return results, nil
}



