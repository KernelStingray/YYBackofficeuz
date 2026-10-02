package photon

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Place struct {
	Name        string  `json:"name"`
	Country     string  `json:"country"`
	City        string  `json:"city"`
	Street      string  `json:"street"`
	Housenumber string  `json:"housenumber"`
	Postcode    string  `json:"postcode"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = "http://photon:2322"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type geoJSONResponse struct {
	Features []struct {
		Geometry struct {
			Coordinates []float64 `json:"coordinates"`
		} `json:"geometry"`
		Properties struct {
			Name        string `json:"name"`
			Country     string `json:"country"`
			City        string `json:"city"`
			Street      string `json:"street"`
			Housenumber string `json:"housenumber"`
			Postcode    string `json:"postcode"`
		} `json:"properties"`
	} `json:"features"`
}

func (c *Client) Geocode(ctx context.Context, query string) ([]Place, error) {
	reqURL := fmt.Sprintf("%s/api?q=%s&limit=5", c.baseURL, url.QueryEscape(query))
	return c.fetchPlaces(ctx, reqURL)
}

func (c *Client) ReverseGeocode(ctx context.Context, lat, lon float64) (*Place, error) {
	reqURL := fmt.Sprintf("%s/reverse?lat=%f&lon=%f", c.baseURL, lat, lon)
	places, err := c.fetchPlaces(ctx, reqURL)
	if err != nil {
		return nil, err
	}
	if len(places) == 0 {
		return nil, fmt.Errorf("no address found for coordinates (%f, %f)", lat, lon)
	}
	return &places[0], nil
}

func (c *Client) fetchPlaces(ctx context.Context, reqURL string) ([]Place, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("photon request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("photon returned status: %s", resp.Status)
	}

	var res geoJSONResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to parse photon response: %w", err)
	}

	var places []Place
	for _, feat := range res.Features {
		p := Place{
			Name:        feat.Properties.Name,
			Country:     feat.Properties.Country,
			City:        feat.Properties.City,
			Street:      feat.Properties.Street,
			Housenumber: feat.Properties.Housenumber,
			Postcode:    feat.Properties.Postcode,
		}
		if len(feat.Geometry.Coordinates) >= 2 {
			p.Longitude = feat.Geometry.Coordinates[0]
			p.Latitude = feat.Geometry.Coordinates[1]
		}
		places = append(places, p)
	}

	return places, nil
}
