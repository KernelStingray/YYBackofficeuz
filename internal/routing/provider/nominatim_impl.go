package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/hextec/yolyolakay-backend/internal/routing/domain"
)

type nominatimProvider struct {
	baseURL string
	client  *http.Client
}

func NewNominatimProvider(baseURL string) domain.GeocodingProvider {
	return &nominatimProvider{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type nominatimResponse struct {
	DisplayName string `json:"display_name"`
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
}

func (p *nominatimProvider) GeocodeAddress(ctx context.Context, query string) ([]domain.GeocodedAddress, error) {
	reqURL := fmt.Sprintf("%s/search?q=%s&format=json&addressdetails=1&limit=5&countrycodes=uz", p.baseURL, url.QueryEscape(query))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Yolyolakay/1.0")
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("nominatim API returned status: %d", resp.StatusCode)
	}

	var nomRes []nominatimResponse
	if err := json.NewDecoder(resp.Body).Decode(&nomRes); err != nil {
		return nil, err
	}

	var results []domain.GeocodedAddress
	for _, r := range nomRes {
		lat, _ := strconv.ParseFloat(r.Lat, 64)
		lon, _ := strconv.ParseFloat(r.Lon, 64)
		results = append(results, domain.GeocodedAddress{
			DisplayName: r.DisplayName,
			Coordinate: domain.Coordinate{
				Latitude:  lat,
				Longitude: lon,
			},
		})
	}
	return results, nil
}

func (p *nominatimProvider) ReverseGeocode(ctx context.Context, coord domain.Coordinate) (*domain.GeocodedAddress, error) {
	reqURL := fmt.Sprintf("%s/reverse?lat=%f&lon=%f&format=json&addressdetails=1", p.baseURL, coord.Latitude, coord.Longitude)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Yolyolakay/1.0")
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("nominatim API returned status: %d", resp.StatusCode)
	}

	var nomRes nominatimResponse
	if err := json.NewDecoder(resp.Body).Decode(&nomRes); err != nil {
		return nil, err
	}

	lat, _ := strconv.ParseFloat(nomRes.Lat, 64)
	lon, _ := strconv.ParseFloat(nomRes.Lon, 64)

	return &domain.GeocodedAddress{
		DisplayName: nomRes.DisplayName,
		Coordinate: domain.Coordinate{
			Latitude:  lat,
			Longitude: lon,
		},
	}, nil
}
