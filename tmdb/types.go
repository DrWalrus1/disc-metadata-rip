package tmdb

import "strings"

// Show represents a TV show search result.
type Show struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// SeasonSummary is the season entry returned in show details.
type SeasonSummary struct {
	SeasonNumber int    `json:"season_number"`
	Name         string `json:"name"`
}

// ShowDetails holds full show metadata including the season list.
type ShowDetails struct {
	ID      int             `json:"id"`
	Name    string          `json:"name"`
	Seasons []SeasonSummary `json:"seasons"`
}

// MatchSeason returns the season number whose name best matches discTitle,
// using a case-insensitive substring search. Generic "Season N" names are
// skipped. Returns 0 if no named season matches.
func MatchSeason(discTitle string, seasons []SeasonSummary) int {
	lower := strings.ToLower(discTitle)
	bestLen, bestSeason := 0, 0
	for _, s := range seasons {
		name := strings.TrimSpace(s.Name)
		if name == "" || s.SeasonNumber == 0 {
			continue
		}
		// Skip generic "Season N" names — they won't appear in disc titles.
		lname := strings.ToLower(name)
		if strings.HasPrefix(lname, "season ") {
			continue
		}
		if strings.Contains(lower, lname) && len(name) > bestLen {
			bestLen = len(name)
			bestSeason = s.SeasonNumber
		}
	}
	return bestSeason
}

// Season represents a TV season with its episodes.
type Season struct {
	Episodes []Episode `json:"episodes"`
}

// Episode represents a single TV episode.
type Episode struct {
	ID            int    `json:"id"`
	EpisodeNumber int    `json:"episode_number"`
	Name          string `json:"name"`
	Overview      string `json:"overview"`
	Runtime       int    `json:"runtime"`
}

// Movie represents a movie search result.
type Movie struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// MovieDetails represents full movie metadata.
type MovieDetails struct {
	Title    string `json:"title"`
	Runtime  int    `json:"runtime"`
	Overview string `json:"overview"`
}

// EpisodesForDisc returns the slice of episodes starting at
// startEpisode (1-indexed). If startEpisode is 0, starts from episode 1.
func EpisodesForDisc(season *Season, startEpisode, numEpisodes int) []Episode {
	startIdx := 0
	if startEpisode > 1 {
		startIdx = startEpisode - 1
	}
	if startIdx >= len(season.Episodes) {
		return nil
	}
	end := startIdx + numEpisodes
	if end > len(season.Episodes) {
		end = len(season.Episodes)
	}
	return season.Episodes[startIdx:end]
}
