package osrm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Point struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type RouteResult struct {
	Distance float64 `json:"distance"` // Meters
	Duration float64 `json:"duration"` // Seconds
	Geometry []Point `json:"geometry"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = "http://osrm:5000"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type osrmResponse struct {
	Code   string      `json:"code"`
	Routes []osrmRoute `json:"routes"`
}

type osrmRoute struct {
	Distance float64      `json:"distance"`
	Duration float64      `json:"duration"`
	Geometry osrmGeometry `json:"geometry"`
}

type osrmGeometry struct {
	Coordinates [][]float64 `json:"coordinates"`
}

func (c *Client) GetRoute(ctx context.Context, points []Point) (*RouteResult, error) {
	if len(points) < 2 {
		return nil, errors.New("at least 2 points required")
	}

	var coords []string
	for _, p := range points {
		coords = append(coords, fmt.Sprintf("%f,%f", p.Longitude, p.Latitude))
	}
	coordsStr := strings.Join(coords, ";")

	url := fmt.Sprintf("%s/route/v1/driving/%s?overview=full&geometries=geojson", c.baseURL, coordsStr)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("osrm request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("osrm returned status: %s", resp.Status)
	}

	var res osrmResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to parse osrm response: %w", err)
	}

	if res.Code != "Ok" || len(res.Routes) == 0 {
		return nil, errors.New("no route found by osrm")
	}

	r := res.Routes[0]
	result := &RouteResult{
		Distance: r.Distance,
		Duration: r.Duration,
	}

	for _, coord := range r.Geometry.Coordinates {
		if len(coord) >= 2 {
			result.Geometry = append(result.Geometry, Point{
				Longitude: coord[0],
				Latitude:  coord[1],
			})
		}
	}

	return result, nil
}
