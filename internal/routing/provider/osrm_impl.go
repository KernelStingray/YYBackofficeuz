package provider

import (
	"context"

	"github.com/hextec/yolyolakay-backend/internal/routing/domain"
	"github.com/hextec/yolyolakay-backend/pkg/integration/osrm"
)

type osrmProvider struct {
	client *osrm.Client
}

func NewOSRMProvider(baseURL string) domain.RoutingProvider {
	return &osrmProvider{
		client: osrm.NewClient(baseURL),
	}
}

func (p *osrmProvider) GetRoute(ctx context.Context, points []domain.Coordinate) (*domain.Route, error) {
	var osrmPoints []osrm.Point
	for _, pt := range points {
		osrmPoints = append(osrmPoints, osrm.Point{
			Latitude:  pt.Latitude,
			Longitude: pt.Longitude,
		})
	}

	res, err := p.client.GetRoute(ctx, osrmPoints)
	if err != nil {
		return nil, err
	}

	route := &domain.Route{
		Distance: res.Distance,
		Duration: res.Duration,
	}

	for _, pt := range res.Geometry {
		route.Geometry = append(route.Geometry, domain.Coordinate{
			Latitude:  pt.Latitude,
			Longitude: pt.Longitude,
		})
	}

	return route, nil
}
