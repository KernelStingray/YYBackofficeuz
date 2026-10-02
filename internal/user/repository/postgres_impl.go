package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/user/domain"
	"github.com/hextec/yolyolakay-backend/internal/user/repository/postgres"
	"github.com/hextec/yolyolakay-backend/pkg/database"
	"github.com/jackc/pgx/v5/pgtype"
)

// === Helpers ===
func uuidToPg(id uuid.UUID) pgtype.UUID {
	return database.UUIDToPg(id)
}
func pgToUUID(id pgtype.UUID) uuid.UUID {
	return database.PgToUUID(id)
}

// === User Repository ===
type userRepo struct {
	q *postgres.Queries
}

func NewUserRepository(q *postgres.Queries) domain.UserRepository {
	return &userRepo{q: q}
}

func mapToDomain(u postgres.User) *domain.User {
	var gender *string
	if u.Gender.Valid {
		g := u.Gender.String
		gender = &g
	}
	var locale string
	if u.Locale.Valid && u.Locale.String != "" {
		locale = u.Locale.String
	} else {
		locale = "uz" // default
	}
	var passwordHash *string
	if u.PasswordHash.Valid {
		p := u.PasswordHash.String
		passwordHash = &p
	}

	var avatarURL *string
	if u.AvatarUrl.Valid {
		a := u.AvatarUrl.String
		avatarURL = &a
	}

	var kycRejectionReason *string
	if u.KycRejectionReason.Valid {
		krr := u.KycRejectionReason.String
		kycRejectionReason = &krr
	}

	return &domain.User{
		ID:                 pgToUUID(u.ID),
		FirstName:          u.FirstName,
		LastName:           u.LastName,
		Gender:             gender,
		PhoneNumber:        u.PhoneNumber,
		Role:               domain.UserRole(u.Role),
		KycStatus:          u.KycStatus.String,
		KycRejectionReason: kycRejectionReason,
		Status:             u.Status,
		Locale:             locale,
		PasswordHash:       passwordHash,
		Rating:             u.Rating,
		RatingCount:        u.RatingCount,
		AvatarURL:          avatarURL,
		TrustedDevices:     u.TrustedDevices,
		CreatedAt:          u.CreatedAt.Time,
		UpdatedAt:          u.UpdatedAt.Time,
	}
}

func (r *userRepo) GetAdminStats(ctx context.Context) (*domain.AdminStats, error) {
	totalUsers, err := r.q.GetTotalUsersCount(ctx)
	if err != nil {
		return nil, err
	}

	totalDrivers, err := r.q.GetTotalDriversCount(ctx)
	if err != nil {
		return nil, err
	}

	totalPassengers, err := r.q.GetTotalPassengersCount(ctx)
	if err != nil {
		return nil, err
	}

	chartDataPG, err := r.q.GetUsersChartData(ctx)
	if err != nil {
		return nil, err
	}

	var chartData []domain.ChartDataPoint
	for _, row := range chartDataPG {
		chartData = append(chartData, domain.ChartDataPoint{
			Date:  row.Date,
			Count: row.Count,
		})
	}

	return &domain.AdminStats{
		TotalUsers:      totalUsers,
		DriversCount:    totalDrivers,
		PassengersCount: totalPassengers,
		ChartData:       chartData,
	}, nil
}

func (r *userRepo) SearchAdminUsers(ctx context.Context, page, limit int32, search, role, status, kycStatus string) ([]*domain.User, int64, error) {
	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	searchParams := postgres.SearchUsersAdminParams{
		Search:    search,
		Role:      role,
		Status:    status,
		KycStatus: kycStatus,
		Limit:     limit,
		Offset:    offset,
	}

	countParams := postgres.CountSearchUsersAdminParams{
		Search:    search,
		Role:      role,
		Status:    status,
		KycStatus: kycStatus,
	}

	totalCount, err := r.q.CountSearchUsersAdmin(ctx, countParams)
	if err != nil {
		return nil, 0, err
	}

	usersPG, err := r.q.SearchUsersAdmin(ctx, searchParams)
	if err != nil {
		return nil, 0, err
	}

	var users []*domain.User
	for _, u := range usersPG {
		users = append(users, mapToDomain(u))
	}

	return users, totalCount, nil
}

func (r *userRepo) Create(ctx context.Context, firstName, lastName, phoneNumber string, role domain.UserRole, gender *string) (*domain.User, error) {
	pgGender := pgtype.Text{Valid: false}
	if gender != nil {
		pgGender = pgtype.Text{String: *gender, Valid: true}
	}

	u, err := r.q.CreateUser(ctx, postgres.CreateUserParams{
		FirstName:   firstName,
		LastName:    lastName,
		PhoneNumber: phoneNumber,
		Role:        string(role),
		Gender:      pgGender,
	})
	if err != nil {
		return nil, err
	}
	return mapToDomain(u), nil
}

func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	u, err := r.q.GetUserByID(ctx, uuidToPg(id))
	if err != nil {
		return nil, err
	}
	return mapToDomain(u), nil
}

func (r *userRepo) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	u, err := r.q.GetUserByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}
	return mapToDomain(u), nil
}

func (r *userRepo) GetAllUsers(ctx context.Context) ([]*domain.User, error) {
	users, err := r.q.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.User, len(users))
	for i, u := range users {
		result[i] = mapToDomain(u)
	}
	return result, nil
}

func (r *userRepo) UpdateKYC(ctx context.Context, id uuid.UUID, kycStatus string, firstName, lastName string) (*domain.User, error) {
	u, err := r.q.UpdateUserKYC(ctx, postgres.UpdateUserKYCParams{
		ID:        uuidToPg(id),
		KycStatus: pgtype.Text{String: kycStatus, Valid: true},
		FirstName: firstName,
		LastName:  lastName,
	})
	if err != nil {
		return nil, err
	}
	return mapToDomain(u), nil
}

func (r *userRepo) UpdateKYCStatusAndRole(ctx context.Context, id uuid.UUID, kycStatus string, role domain.UserRole, reason *string) (*domain.User, error) {
	pgReason := pgtype.Text{Valid: false}
	if reason != nil {
		pgReason = pgtype.Text{String: *reason, Valid: true}
	}

	u, err := r.q.UpdateUserKYCStatusAndRole(ctx, postgres.UpdateUserKYCStatusAndRoleParams{
		ID:                 uuidToPg(id),
		KycStatus:          pgtype.Text{String: kycStatus, Valid: true},
		Role:               string(role),
		KycRejectionReason: pgReason,
	})
	if err != nil {
		return nil, err
	}
	return mapToDomain(u), nil
}

func (r *userRepo) UpdateCarsModerationStatusByUserID(ctx context.Context, userID uuid.UUID, status string) error {
	return r.q.UpdateCarsModerationStatusByUserID(ctx, postgres.UpdateCarsModerationStatusByUserIDParams{
		UserID:           uuidToPg(userID),
		ModerationStatus: pgtype.Text{String: status, Valid: true},
	})
}

func (r *userRepo) UpdateProfile(ctx context.Context, id uuid.UUID, firstName, lastName string, gender *string, locale *string, avatarURL *string) (*domain.User, error) {
	pgGender := pgtype.Text{Valid: false}
	if gender != nil {
		pgGender = pgtype.Text{String: *gender, Valid: true}
	}

	pgLocale := pgtype.Text{Valid: false}
	if locale != nil {
		pgLocale = pgtype.Text{String: *locale, Valid: true}
	}

	pgAvatar := pgtype.Text{Valid: false}
	if avatarURL != nil {
		pgAvatar = pgtype.Text{String: *avatarURL, Valid: true}
	}

	u, err := r.q.UpdateUserProfile(ctx, postgres.UpdateUserProfileParams{
		ID:        uuidToPg(id),
		FirstName: firstName,
		LastName:  lastName,
		Gender:    pgGender,
		Locale:    pgLocale,
		AvatarUrl: pgAvatar,
	})
	if err != nil {
		return nil, err
	}
	return mapToDomain(u), nil
}

func (r *userRepo) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	_, err := r.q.UpdateUserPassword(ctx, postgres.UpdateUserPasswordParams{
		ID:           uuidToPg(id),
		PasswordHash: pgtype.Text{String: passwordHash, Valid: true},
	})
	return err
}

func (r *userRepo) AddTrustedDevice(ctx context.Context, id uuid.UUID, deviceID string) error {
	return r.q.AddTrustedDevice(ctx, postgres.AddTrustedDeviceParams{
		ID:          uuidToPg(id),
		ArrayAppend: deviceID,
	})
}

func (r *userRepo) UpdateRating(ctx context.Context, id uuid.UUID, newRating float64, newCount int32) error {
	err := r.q.UpdateUserRating(ctx, postgres.UpdateUserRatingParams{
		ID:          uuidToPg(id),
		Rating:      newRating,
		RatingCount: newCount,
	})
	return err
}

func (r *userRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.SoftDeleteUser(ctx, uuidToPg(id))
}

func (r *userRepo) UpdateUserStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.q.UpdateUserStatus(ctx, postgres.UpdateUserStatusParams{
		ID:     uuidToPg(id),
		Status: status,
	})
	return err
}

func (r *userRepo) SystemBlockUser(ctx context.Context, id uuid.UUID) error {
	return r.UpdateUserStatus(ctx, id, "blocked")
}

func (r *userRepo) BlockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	return r.q.BlockUser(ctx, postgres.BlockUserParams{
		BlockerID: uuidToPg(blockerID),
		BlockedID: uuidToPg(blockedID),
	})
}

func (r *userRepo) UnblockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	return r.q.UnblockUser(ctx, postgres.UnblockUserParams{
		BlockerID: uuidToPg(blockerID),
		BlockedID: uuidToPg(blockedID),
	})
}

func (r *userRepo) GetBlockedUsers(ctx context.Context, blockerID uuid.UUID) ([]uuid.UUID, error) {
	pgIds, err := r.q.GetBlockedUsers(ctx, uuidToPg(blockerID))
	if err != nil {
		return nil, err
	}
	var ids []uuid.UUID
	for _, id := range pgIds {
		ids = append(ids, pgToUUID(id))
	}
	return ids, nil
}

func (r *userRepo) CreateReport(ctx context.Context, reportID, reporterID, reportedID uuid.UUID, reason string) error {
	return r.q.CreateReport(ctx, postgres.CreateReportParams{
		ID:         uuidToPg(reportID),
		ReporterID: uuidToPg(reporterID),
		ReportedID: uuidToPg(reportedID),
		Reason:     reason,
	})
}

// === Car Repository ===
type carRepo struct {
	q *postgres.Queries
}

func NewCarRepository(q *postgres.Queries) domain.CarRepository {
	return &carRepo{q: q}
}

func mapCarToDomain(c postgres.Car) domain.Car {
	var techPassport *string
	if c.TechPassportPhoto.Valid {
		p := c.TechPassportPhoto.String
		techPassport = &p
	}
	var driverLicense *string
	if c.DriverLicensePhoto.Valid {
		p := c.DriverLicensePhoto.String
		driverLicense = &p
	}
	var techPassportFront *string
	if c.TechPassportFront.Valid {
		p := c.TechPassportFront.String
		techPassportFront = &p
	}
	var techPassportBack *string
	if c.TechPassportBack.Valid {
		p := c.TechPassportBack.String
		techPassportBack = &p
	}
	var driverLicenseFront *string
	if c.DriverLicenseFront.Valid {
		p := c.DriverLicenseFront.String
		driverLicenseFront = &p
	}
	var driverLicenseBack *string
	if c.DriverLicenseBack.Valid {
		p := c.DriverLicenseBack.String
		driverLicenseBack = &p
	}
	var carFrontPhoto *string
	if c.CarFrontPhoto.Valid {
		p := c.CarFrontPhoto.String
		carFrontPhoto = &p
	}
	var carBackPhoto *string
	if c.CarBackPhoto.Valid {
		p := c.CarBackPhoto.String
		carBackPhoto = &p
	}
	var carSidePhoto *string
	if c.CarSidePhoto.Valid {
		p := c.CarSidePhoto.String
		carSidePhoto = &p
	}
	status := c.ModerationStatus.String
	if !c.ModerationStatus.Valid || status == "" {
		status = "pending"
	}

	// Fallback to legacy single photo fields if newer specific angle fields are absent
	if techPassportFront == nil && techPassport != nil {
		techPassportFront = techPassport
	}
	if driverLicenseFront == nil && driverLicense != nil {
		driverLicenseFront = driverLicense
	}

	return domain.Car{
		ID:                 pgToUUID(c.ID),
		UserID:             pgToUUID(c.UserID),
		GovNumber:          c.GovNumber,
		Brand:              c.Brand,
		Model:              c.Model,
		Color:              c.Color.String,
		Year:               c.Year.Int32,
		TechPassportPhoto:  techPassport,
		DriverLicensePhoto: driverLicense,
		TechPassportFront:  techPassportFront,
		TechPassportBack:   techPassportBack,
		DriverLicenseFront: driverLicenseFront,
		DriverLicenseBack:  driverLicenseBack,
		CarFrontPhoto:      carFrontPhoto,
		CarBackPhoto:       carBackPhoto,
		CarSidePhoto:       carSidePhoto,
		ModerationStatus:   status,
		CreatedAt:          c.CreatedAt.Time,
		UpdatedAt:          c.UpdatedAt.Time,
	}
}

func (r *carRepo) Create(ctx context.Context, c *domain.Car) (*domain.Car, error) {
	pgText := func(s *string) pgtype.Text {
		if s != nil {
			return pgtype.Text{String: *s, Valid: true}
		}
		return pgtype.Text{Valid: false}
	}

	created, err := r.q.CreateCar(ctx, postgres.CreateCarParams{
		UserID:             uuidToPg(c.UserID),
		GovNumber:          c.GovNumber,
		Brand:              c.Brand,
		Model:              c.Model,
		Color:              pgtype.Text{String: c.Color, Valid: c.Color != ""},
		Year:               pgtype.Int4{Int32: c.Year, Valid: c.Year > 0},
		TechPassportPhoto:  pgText(c.TechPassportPhoto),
		DriverLicensePhoto: pgText(c.DriverLicensePhoto),
		TechPassportFront:  pgText(c.TechPassportFront),
		TechPassportBack:   pgText(c.TechPassportBack),
		DriverLicenseFront: pgText(c.DriverLicenseFront),
		DriverLicenseBack:  pgText(c.DriverLicenseBack),
		CarFrontPhoto:      pgText(c.CarFrontPhoto),
		CarBackPhoto:       pgText(c.CarBackPhoto),
		CarSidePhoto:       pgText(c.CarSidePhoto),
	})
	if err != nil {
		return nil, err
	}
	domainC := mapCarToDomain(created)
	return &domainC, nil
}

func (r *carRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Car, error) {
	res, err := r.q.GetCarsByUserID(ctx, uuidToPg(userID))
	if err != nil {
		return nil, err
	}
	list := make([]domain.Car, len(res))
	for i, c := range res {
		list[i] = mapCarToDomain(c)
	}
	return list, nil
}

func (r *carRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Car, error) {
	res, err := r.q.GetCarByID(ctx, uuidToPg(id))
	if err != nil {
		return nil, err
	}
	c := mapCarToDomain(res)
	return &c, nil
}

func (r *carRepo) Update(ctx context.Context, c *domain.Car) (*domain.Car, error) {
	updated, err := r.q.UpdateCar(ctx, postgres.UpdateCarParams{
		ID:        uuidToPg(c.ID),
		GovNumber: c.GovNumber,
		Brand:     c.Brand,
		Model:     c.Model,
		Color:     c.Color,
		Year:      c.Year,
	})
	if err != nil {
		return nil, err
	}
	domainC := mapCarToDomain(updated)
	return &domainC, nil
}

func (r *carRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteCar(ctx, uuidToPg(id))
}

