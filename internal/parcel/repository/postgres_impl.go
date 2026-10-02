package repository

import (
	"math/big"
	"strconv"
	"time"

	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hextec/yolyolakay-backend/internal/parcel/domain"
	sqlc "github.com/hextec/yolyolakay-backend/internal/parcel/repository/postgres"
	"github.com/hextec/yolyolakay-backend/pkg/database"
)

type postgresParcelRepository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

func NewPostgresParcelRepository(db *pgxpool.Pool) domain.ParcelRepository {
	return &postgresParcelRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func parseFloatFromInterface(i interface{}) float64 {
	if i == nil {
		return 0.0
	}
	switch v := i.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case string:
		f, _ := strconv.ParseFloat(v, 64)
		return f
	default:
		return 0.0
	}
}

func uuidToPg(id uuid.UUID) pgtype.UUID {
	return database.UUIDToPg(id)
}

func pgToUUID(id pgtype.UUID) uuid.UUID {
	return database.PgToUUID(id)
}

func (r *postgresParcelRepository) CreateParcel(ctx context.Context, p *domain.Parcel) error {
	var num pgtype.Numeric
	_ = num.Scan(fmt.Sprintf("%f", p.Price))

	pgP, err := r.queries.CreateParcel(ctx, sqlc.CreateParcelParams{
		ID:                     uuidToPg(p.ID),
		SenderID:               uuidToPg(p.SenderID),
		StMakepoint:            p.OriginLon,
		StMakepoint_2:          p.OriginLat,
		StMakepoint_3:          p.DestinationLon,
		StMakepoint_4:          p.DestinationLat,
		ReceiverPhoneEncrypted: p.ReceiverPhoneEncrypted,
		Price:                  num,
		OriginAddress:          pgtype.Text{String: p.OriginAddress, Valid: p.OriginAddress != ""},
		DestinationAddress:     pgtype.Text{String: p.DestinationAddress, Valid: p.DestinationAddress != ""},
		Description:            pgtype.Text{String: p.Description, Valid: p.Description != ""},
		Size:                   pgtype.Text{String: p.Size, Valid: p.Size != ""},
		DeliverBy: func() pgtype.Timestamptz {
			if p.DeliverBy != nil {
				return pgtype.Timestamptz{Time: *p.DeliverBy, Valid: true}
			}
			return pgtype.Timestamptz{Valid: false}
		}(),
		ReceiverName: pgtype.Text{String: p.ReceiverName, Valid: p.ReceiverName != ""},
		DropoffType:  pgtype.Text{String: p.DropoffType, Valid: p.DropoffType != ""},
	})
	if err != nil {
		return err
	}

	p.CreatedAt = pgP.CreatedAt.Time
	p.UpdatedAt = pgP.UpdatedAt.Time
	p.Status = domain.ParcelStatus(pgP.Status)
	p.ReceiverName = pgP.ReceiverName.String
	p.DropoffType = pgP.DropoffType.String
	return nil
}

func (r *postgresParcelRepository) GetParcelByID(ctx context.Context, id uuid.UUID) (*domain.Parcel, error) {
	pgP, err := r.queries.GetParcelByID(ctx, uuidToPg(id))
	if err != nil {
		return nil, err
	}

	val, _ := pgP.Price.Float64Value()
	price := val.Float64

	var driverID *uuid.UUID
	if pgP.DriverID.Valid {
		did := pgToUUID(pgP.DriverID)
		driverID = &did
	}

	return &domain.Parcel{
		ID:                     pgToUUID(pgP.ID),
		SenderID:               pgToUUID(pgP.SenderID),
		DriverID:               driverID,
		OriginLon:              parseFloatFromInterface(pgP.OriginLon),
		OriginLat:              parseFloatFromInterface(pgP.OriginLat),
		DestinationLon:         parseFloatFromInterface(pgP.DestinationLon),
		DestinationLat:         parseFloatFromInterface(pgP.DestinationLat),
		ReceiverPhoneEncrypted: pgP.ReceiverPhoneEncrypted,
		Price:                  price,
		Status:                 domain.ParcelStatus(pgP.Status),
		OriginAddress:          pgP.OriginAddress.String,
		DestinationAddress:     pgP.DestinationAddress.String,
		Description:            pgP.Description.String,
		Size:                   pgP.Size.String,
		ReceiverName:           pgP.ReceiverName.String,
		DropoffType:            pgP.DropoffType.String,
		DeliverBy: func() *time.Time {
			if pgP.DeliverBy.Valid {
				t := pgP.DeliverBy.Time
				return &t
			}
			return nil
		}(),
		CreatedAt: pgP.CreatedAt.Time,
		UpdatedAt: pgP.UpdatedAt.Time,
	}, nil
}

func (r *postgresParcelRepository) UpdateParcelStatus(ctx context.Context, id uuid.UUID, status domain.ParcelStatus, driverID *uuid.UUID) error {
	var pgDriverID pgtype.UUID
	if driverID != nil {
		pgDriverID = uuidToPg(*driverID)
	}

	_, err := r.queries.UpdateParcelStatus(ctx, sqlc.UpdateParcelStatusParams{
		Status:   string(status),
		DriverID: pgDriverID,
		ID:       uuidToPg(id),
	})
	return err
}

func (r *postgresParcelRepository) SearchParcels(ctx context.Context, params domain.ParcelSearchParams) ([]domain.Parcel, error) {
	if params.DestLon != nil && params.DestLat != nil {
		corridorMeters := 25000.0
		if params.CorridorMeters != nil && *params.CorridorMeters > 0 {
			corridorMeters = *params.CorridorMeters
		}

		sqlcCorridorParams := sqlc.SearchParcelsInCorridorParams{
			StMakepoint:   params.Lon,
			StMakepoint_2: params.Lat,
			StMakepoint_3: *params.DestLon,
			StMakepoint_4: *params.DestLat,
			StDwithin:     corridorMeters,
		}

		if params.PriceMin != nil {
			sqlcCorridorParams.PriceMin = pgtype.Numeric{Int: big.NewInt(int64(*params.PriceMin)), Valid: true}
		}
		if params.PriceMax != nil {
			sqlcCorridorParams.PriceMax = pgtype.Numeric{Int: big.NewInt(int64(*params.PriceMax)), Valid: true}
		}

		pgParcels, err := r.queries.SearchParcelsInCorridor(ctx, sqlcCorridorParams)
		if err != nil {
			return nil, err
		}

		var res []domain.Parcel
		for _, pgP := range pgParcels {
			val, _ := pgP.Price.Float64Value()
			price := val.Float64

			res = append(res, domain.Parcel{
				ID:                     pgToUUID(pgP.ID),
				SenderID:               pgToUUID(pgP.SenderID),
				OriginLon:              parseFloatFromInterface(pgP.OriginLon),
				OriginLat:              parseFloatFromInterface(pgP.OriginLat),
				DestinationLon:         parseFloatFromInterface(pgP.DestinationLon),
				DestinationLat:         parseFloatFromInterface(pgP.DestinationLat),
				ReceiverPhoneEncrypted: pgP.ReceiverPhoneEncrypted,
				Status:                 domain.ParcelStatus(pgP.Status),
				Price:                  price,
				OriginAddress:          pgP.OriginAddress.String,
				DestinationAddress:     pgP.DestinationAddress.String,
				Description:            pgP.Description.String,
				Size:                   pgP.Size.String,
				ReceiverName:           pgP.ReceiverName.String,
				DropoffType:            pgP.DropoffType.String,
				DeliverBy: func() *time.Time {
					if pgP.DeliverBy.Valid {
						t := pgP.DeliverBy.Time
						return &t
					}
					return nil
				}(),
				CreatedAt: pgP.CreatedAt.Time,
				UpdatedAt: pgP.UpdatedAt.Time,
			})
		}
		return res, nil
	}

	sqlcParams := sqlc.SearchParcelsParams{
		StMakepoint:   params.Lon,
		StMakepoint_2: params.Lat,
		StDwithin:     params.RadiusMeters,
	}

	if params.PriceMin != nil {
		sqlcParams.PriceMin = pgtype.Numeric{Int: big.NewInt(int64(*params.PriceMin)), Valid: true}
	}
	if params.PriceMax != nil {
		sqlcParams.PriceMax = pgtype.Numeric{Int: big.NewInt(int64(*params.PriceMax)), Valid: true}
	}

	pgParcels, err := r.queries.SearchParcels(ctx, sqlcParams)
	if err != nil {
		return nil, err
	}

	var res []domain.Parcel
	for _, pgP := range pgParcels {
		val, _ := pgP.Price.Float64Value()
		price := val.Float64

		res = append(res, domain.Parcel{
			ID:                     pgToUUID(pgP.ID),
			SenderID:               pgToUUID(pgP.SenderID),
			OriginLon:              parseFloatFromInterface(pgP.OriginLon),
			OriginLat:              parseFloatFromInterface(pgP.OriginLat),
			DestinationLon:         parseFloatFromInterface(pgP.DestinationLon),
			DestinationLat:         parseFloatFromInterface(pgP.DestinationLat),
			ReceiverPhoneEncrypted: pgP.ReceiverPhoneEncrypted,
			Status:                 domain.ParcelStatus(pgP.Status),
			Price:                  price,
			OriginAddress:          pgP.OriginAddress.String,
			DestinationAddress:     pgP.DestinationAddress.String,
			Description:            pgP.Description.String,
			Size:                   pgP.Size.String,
			ReceiverName:           pgP.ReceiverName.String,
			DropoffType:            pgP.DropoffType.String,
			DeliverBy: func() *time.Time {
				if pgP.DeliverBy.Valid {
					t := pgP.DeliverBy.Time
					return &t
				}
				return nil
			}(),
			CreatedAt: pgP.CreatedAt.Time,
			UpdatedAt: pgP.UpdatedAt.Time,
		})
	}

	return res, nil
}

func (r *postgresParcelRepository) GetAllParcels(ctx context.Context) ([]domain.Parcel, error) {
	pgParcels, err := r.queries.GetAllParcels(ctx)
	if err != nil {
		return nil, err
	}

	var res []domain.Parcel
	for _, pgP := range pgParcels {
		val, _ := pgP.Price.Float64Value()
		price := val.Float64

		var driverID *uuid.UUID
		if pgP.DriverID.Valid {
			did := pgToUUID(pgP.DriverID)
			driverID = &did
		}

		res = append(res, domain.Parcel{
			ID:                     pgToUUID(pgP.ID),
			SenderID:               pgToUUID(pgP.SenderID),
			DriverID:               driverID,
			OriginLon:              parseFloatFromInterface(pgP.OriginLon),
			OriginLat:              parseFloatFromInterface(pgP.OriginLat),
			DestinationLon:         parseFloatFromInterface(pgP.DestinationLon),
			DestinationLat:         parseFloatFromInterface(pgP.DestinationLat),
			ReceiverPhoneEncrypted: pgP.ReceiverPhoneEncrypted,
			Status:                 domain.ParcelStatus(pgP.Status),
			Price:                  price,
			OriginAddress:          pgP.OriginAddress.String,
			DestinationAddress:     pgP.DestinationAddress.String,
			Description:            pgP.Description.String,
			Size:                   pgP.Size.String,
			ReceiverName:           pgP.ReceiverName.String,
			DropoffType:            pgP.DropoffType.String,
			DeliverBy: func() *time.Time {
				if pgP.DeliverBy.Valid {
					t := pgP.DeliverBy.Time
					return &t
				}
				return nil
			}(),
			CreatedAt: pgP.CreatedAt.Time,
			UpdatedAt: pgP.UpdatedAt.Time,
		})
	}

	return res, nil
}

func (r *postgresParcelRepository) GetParcelsBySenderID(ctx context.Context, senderID uuid.UUID) ([]domain.Parcel, error) {
	pgParcels, err := r.queries.GetParcelsBySenderID(ctx, uuidToPg(senderID))
	if err != nil {
		return nil, err
	}

	var res []domain.Parcel
	for _, pgP := range pgParcels {
		val, _ := pgP.Price.Float64Value()
		price := val.Float64

		var driverID *uuid.UUID
		if pgP.DriverID.Valid {
			did := pgToUUID(pgP.DriverID)
			driverID = &did
		}

		res = append(res, domain.Parcel{
			ID:                     pgToUUID(pgP.ID),
			SenderID:               pgToUUID(pgP.SenderID),
			DriverID:               driverID,
			OriginLon:              parseFloatFromInterface(pgP.OriginLon),
			OriginLat:              parseFloatFromInterface(pgP.OriginLat),
			DestinationLon:         parseFloatFromInterface(pgP.DestinationLon),
			DestinationLat:         parseFloatFromInterface(pgP.DestinationLat),
			ReceiverPhoneEncrypted: pgP.ReceiverPhoneEncrypted,
			Status:                 domain.ParcelStatus(pgP.Status),
			Price:                  price,
			OriginAddress:          pgP.OriginAddress.String,
			DestinationAddress:     pgP.DestinationAddress.String,
			Description:            pgP.Description.String,
			Size:                   pgP.Size.String,
			ReceiverName:           pgP.ReceiverName.String,
			DropoffType:            pgP.DropoffType.String,
			DeliverBy: func() *time.Time {
				if pgP.DeliverBy.Valid {
					t := pgP.DeliverBy.Time
					return &t
				}
				return nil
			}(),
			ResponsesCount: pgP.ResponsesCount,
			CreatedAt:      pgP.CreatedAt.Time,
			UpdatedAt:      pgP.UpdatedAt.Time,
		})
	}

	return res, nil
}

func (r *postgresParcelRepository) GetParcelsByDriverID(ctx context.Context, driverID uuid.UUID) ([]domain.Parcel, error) {
	pgParcels, err := r.queries.GetParcelsByDriverID(ctx, uuidToPg(driverID))
	if err != nil {
		return nil, err
	}

	var res []domain.Parcel
	for _, pgP := range pgParcels {
		val, _ := pgP.Price.Float64Value()
		price := val.Float64

		var did *uuid.UUID
		if pgP.DriverID.Valid {
			d := pgToUUID(pgP.DriverID)
			did = &d
		}

		res = append(res, domain.Parcel{
			ID:                     pgToUUID(pgP.ID),
			SenderID:               pgToUUID(pgP.SenderID),
			DriverID:               did,
			OriginLon:              parseFloatFromInterface(pgP.OriginLon),
			OriginLat:              parseFloatFromInterface(pgP.OriginLat),
			DestinationLon:         parseFloatFromInterface(pgP.DestinationLon),
			DestinationLat:         parseFloatFromInterface(pgP.DestinationLat),
			ReceiverPhoneEncrypted: pgP.ReceiverPhoneEncrypted,
			Status:                 domain.ParcelStatus(pgP.Status),
			Price:                  price,
			OriginAddress:          pgP.OriginAddress.String,
			DestinationAddress:     pgP.DestinationAddress.String,
			Description:            pgP.Description.String,
			Size:                   pgP.Size.String,
			ReceiverName:           pgP.ReceiverName.String,
			DropoffType:            pgP.DropoffType.String,
			DeliverBy: func() *time.Time {
				if pgP.DeliverBy.Valid {
					t := pgP.DeliverBy.Time
					return &t
				}
				return nil
			}(),
			CreatedAt: pgP.CreatedAt.Time,
			UpdatedAt: pgP.UpdatedAt.Time,
		})
	}

	return res, nil
}

func (r *postgresParcelRepository) CreateParcelResponse(ctx context.Context, response *domain.ParcelResponse) error {
	var pricePg pgtype.Numeric
	pricePg.Scan(strconv.FormatFloat(response.OfferedPrice, 'f', -1, 64))

	res, err := r.queries.CreateParcelResponse(ctx, sqlc.CreateParcelResponseParams{
		ID:           uuidToPg(response.ID),
		ParcelID:     uuidToPg(response.ParcelID),
		DriverID:     uuidToPg(response.DriverID),
		Status:       string(response.Status),
		OfferedPrice: pricePg,
	})
	if err != nil {
		return err
	}

	response.CreatedAt = res.CreatedAt.Time
	response.UpdatedAt = res.UpdatedAt.Time
	return nil
}

func (r *postgresParcelRepository) GetParcelResponsesByParcelID(ctx context.Context, parcelID uuid.UUID) ([]domain.ParcelResponse, error) {
	pgResp, err := r.queries.GetParcelResponsesByParcelID(ctx, uuidToPg(parcelID))
	if err != nil {
		return nil, err
	}

	var res []domain.ParcelResponse
	for _, pr := range pgResp {
		val, _ := pr.OfferedPrice.Float64Value()
		price := val.Float64

		res = append(res, domain.ParcelResponse{
			ID:           pgToUUID(pr.ID),
			ParcelID:     pgToUUID(pr.ParcelID),
			DriverID:     pgToUUID(pr.DriverID),
			Status:       domain.ParcelResponseStatus(pr.Status),
			OfferedPrice: price,
			CreatedAt:    pr.CreatedAt.Time,
			UpdatedAt:    pr.UpdatedAt.Time,
		})
	}
	return res, nil
}

func (r *postgresParcelRepository) UpdateParcelResponseStatus(ctx context.Context, responseID uuid.UUID, status domain.ParcelResponseStatus) error {
	_, err := r.db.Exec(ctx, "UPDATE parcel_responses SET status = $1, updated_at = NOW() WHERE id = $2", string(status), responseID)
	return err
}

func (r *postgresParcelRepository) RejectOtherParcelResponses(ctx context.Context, parcelID uuid.UUID, acceptedResponseID uuid.UUID) error {
	_, err := r.db.Exec(ctx, "UPDATE parcel_responses SET status = 'rejected', updated_at = NOW() WHERE parcel_id = $1 AND id != $2 AND status = 'pending'", parcelID, acceptedResponseID)
	return err
}

func (r *postgresParcelRepository) GetUserStats(ctx context.Context, userID uuid.UUID) (*domain.UserParcelStats, error) {
	sentCount, err := r.queries.CountUserParcelsAsSender(ctx, uuidToPg(userID))
	if err != nil {
		return nil, err
	}

	deliveredCount, err := r.queries.CountUserParcelsAsDriver(ctx, uuidToPg(userID))
	if err != nil {
		return nil, err
	}

	return &domain.UserParcelStats{
		TotalSent:      sentCount,
		TotalDelivered: deliveredCount,
	}, nil
}

func (r *postgresParcelRepository) GetAdminStats(ctx context.Context) (*domain.AdminStats, error) {
	totalCount, err := r.queries.GetTotalParcelsCount(ctx)
	if err != nil {
		totalCount = 0
	}

	activeCount, err := r.queries.GetActiveParcelsCount(ctx)
	if err != nil {
		activeCount = 0
	}

	deliveredCount, err := r.queries.GetDeliveredParcelsCount(ctx)
	if err != nil {
		deliveredCount = 0
	}

	chartRows, err := r.queries.GetParcelsChartData(ctx)
	var chartData []domain.ChartDataPoint
	if err == nil {
		for _, row := range chartRows {
			chartData = append(chartData, domain.ChartDataPoint{
				Date:  row.Date,
				Count: row.Count,
			})
		}
	}

	return &domain.AdminStats{
		TotalParcels:     totalCount,
		ActiveParcels:    activeCount,
		DeliveredParcels: deliveredCount,
		ChartData:        chartData,
	}, nil
}
