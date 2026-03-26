package tmdb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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
