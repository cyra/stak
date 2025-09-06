package search

import (
	"sort"
	"strings"

	"github.com/sahilm/fuzzy"
	"stak/internal/models"
)

type FuzzySearcher struct {
	entries []models.Entry
}

func NewFuzzySearcher() *FuzzySearcher {
	return &FuzzySearcher{
		entries: []models.Entry{},
	}
}

func (f *FuzzySearcher) SetEntries(entries []models.Entry) {
	f.entries = entries
}

func (f *FuzzySearcher) Search(query string) []models.Entry {
	if query == "" {
		return f.entries
	}

	var searchTargets []string
	entryMap := make(map[string]models.Entry)

	for _, entry := range f.entries {
		searchText := f.buildSearchText(entry)
		searchTargets = append(searchTargets, searchText)
		entryMap[searchText] = entry
	}

	matches := fuzzy.Find(query, searchTargets)
	
	var results []models.Entry
	for _, match := range matches {
		if entry, exists := entryMap[match.Str]; exists {
			results = append(results, entry)
		}
	}

	return results
}

func (f *FuzzySearcher) SearchLinks(query string) []models.Entry {
	linkEntries := f.filterByType(models.TypeLink)
	f.SetEntries(linkEntries)
	return f.Search(query)
}

func (f *FuzzySearcher) filterByType(entryType models.EntryType) []models.Entry {
	var filtered []models.Entry
	for _, entry := range f.entries {
		if entry.Type == entryType {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func (f *FuzzySearcher) buildSearchText(entry models.Entry) string {
	var parts []string
	
	parts = append(parts, entry.Content)
	
	if entry.URL != "" {
		parts = append(parts, entry.URL)
	}
	
	if entry.URLTitle != "" {
		parts = append(parts, entry.URLTitle)
	}
	
	parts = append(parts, entry.Tags...)
	
	parts = append(parts, string(entry.Type))
	
	return strings.Join(parts, " ")
}

type RankedResult struct {
	Entry models.Entry
	Score int
}

func (f *FuzzySearcher) RankedSearch(query string) []models.Entry {
	if query == "" {
		return f.sortByRecency(f.entries)
	}

	queryLower := strings.ToLower(query)
	var rankedResults []RankedResult

	for _, entry := range f.entries {
		score := f.calculateScore(entry, queryLower)
		if score > 0 {
			rankedResults = append(rankedResults, RankedResult{
				Entry: entry,
				Score: score,
			})
		}
	}

	sort.Slice(rankedResults, func(i, j int) bool {
		if rankedResults[i].Score == rankedResults[j].Score {
			return rankedResults[i].Entry.CreatedAt.After(rankedResults[j].Entry.CreatedAt)
		}
		return rankedResults[i].Score > rankedResults[j].Score
	})

	var results []models.Entry
	for _, result := range rankedResults {
		results = append(results, result.Entry)
	}

	return results
}

func (f *FuzzySearcher) calculateScore(entry models.Entry, query string) int {
	score := 0
	contentLower := strings.ToLower(entry.Content)

	if strings.Contains(contentLower, query) {
		score += 10
		if strings.HasPrefix(contentLower, query) {
			score += 5
		}
	}

	for _, tag := range entry.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			score += 8
		}
	}

	if entry.URL != "" && strings.Contains(strings.ToLower(entry.URL), query) {
		score += 6
	}

	if entry.URLTitle != "" && strings.Contains(strings.ToLower(entry.URLTitle), query) {
		score += 7
	}

	if strings.Contains(strings.ToLower(string(entry.Type)), query) {
		score += 3
	}

	return score
}

func (f *FuzzySearcher) sortByRecency(entries []models.Entry) []models.Entry {
	sorted := make([]models.Entry, len(entries))
	copy(sorted, entries)
	
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].CreatedAt.After(sorted[j].CreatedAt)
	})
	
	return sorted
}