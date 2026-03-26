package tmdb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const baseURL = "https://api.themoviedb.org/3"

// Client is a TMDB API client.
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// New creates a new TMDB client with the given API key.
func New(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

func (c *Client) get(endpoint string) (*http.Response, error) {
	req, err := http.NewRequest("GET", baseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")
	return c.httpClient.Do(req)
}

// dropLastWord removes the last whitespace-delimited word from s.
// Returns an empty string when no words remain.
func dropLastWord(s string) string {
	s = strings.TrimSpace(s)
	i := strings.LastIndex(s, " ")
	if i < 0 {
		return ""
	}
	return s[:i]
}

// SmartSearchTV searches for a TV show by progressively dropping the last
// word from query until results are found or the query is exhausted.
// Returns the results, the query that produced them, and any error.
// The matched query is empty when nothing was found.
func (c *Client) SmartSearchTV(query string) ([]Show, string, error) {
	for q := query; q != ""; q = dropLastWord(q) {
		shows, err := c.SearchTV(q)
		if err != nil {
			return nil, "", err
		}
		if len(shows) > 0 {
			return shows, q, nil
		}
	}
	return nil, "", nil
}

// SmartSearchMovie searches for a movie by progressively dropping the last
// word from query until results are found or the query is exhausted.
// Returns the results, the query that produced them, and any error.
// The matched query is empty when nothing was found.
func (c *Client) SmartSearchMovie(query string) ([]Movie, string, error) {
	for q := query; q != ""; q = dropLastWord(q) {
		movies, err := c.SearchMovie(q)
		if err != nil {
			return nil, "", err
		}
		if len(movies) > 0 {
			return movies, q, nil
		}
	}
	return nil, "", nil
}

// SearchTV searches for TV shows matching the query.
func (c *Client) SearchTV(query string) ([]Show, error) {
	resp, err := c.get("/search/tv?query=" + url.QueryEscape(query))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Results []Show `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Results, nil
}

// GetSeason fetches a season's episodes for the given show.
func (c *Client) GetSeason(showID, season int) (*Season, error) {
	resp, err := c.get(fmt.Sprintf("/tv/%d/season/%d", showID, season))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var s Season
	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return nil, err
	}
	return &s, nil
}

// SearchMovie searches for movies matching the query.
func (c *Client) SearchMovie(query string) ([]Movie, error) {
	resp, err := c.get("/search/movie?query=" + url.QueryEscape(query))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Results []Movie `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Results, nil
}

// SmartGetSeason fetches the season that best matches discTitle by name.
// It fetches full show details, runs MatchSeason against the season list,
// and falls back to fallbackSeason if no named season matches.
// Returns the season data, the season number used, and any error.
func (c *Client) SmartGetSeason(showID int, discTitle string, fallbackSeason int) (*Season, int, error) {
	details, err := c.GetShow(showID)
	if err != nil {
		return nil, 0, fmt.Errorf("fetching show details: %w", err)
	}

	seasonNum := fallbackSeason
	if matched := MatchSeason(discTitle, details.Seasons); matched > 0 {
		seasonNum = matched
	}

	season, err := c.GetSeason(showID, seasonNum)
	if err != nil {
		return nil, 0, err
	}
	return season, seasonNum, nil
}

// GetShow fetches full show details including the season list.
func (c *Client) GetShow(showID int) (*ShowDetails, error) {
	resp, err := c.get(fmt.Sprintf("/tv/%d", showID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var d ShowDetails
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return nil, err
	}
	return &d, nil
}

// GetMovie fetches full details for a movie by ID.
func (c *Client) GetMovie(movieID int) (*MovieDetails, error) {
	resp, err := c.get(fmt.Sprintf("/movie/%d", movieID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var m MovieDetails
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, err
	}
	return &m, nil
}
