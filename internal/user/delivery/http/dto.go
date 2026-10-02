package http

import "github.com/hextec/yolyolakay-backend/internal/user/domain"

type RegisterPassengerRequest struct {
	FirstName   string  `json:"first_name" binding:"required,min=2,max=50"`
	LastName    string  `json:"last_name" binding:"required,min=2,max=50"`
	PhoneNumber string  `json:"phone_number" binding:"required,e164"` // e164 validates international phone format like +998901234567
	Gender      *string `json:"gender"`
}

type UpdateProfileRequest struct {
	FirstName string  `json:"first_name" binding:"max=50"`
	LastName  string  `json:"last_name" binding:"max=50"`
	Gender    *string `json:"gender"`
	Locale    *string `json:"locale"`
	AvatarURL *string `json:"avatar_url"`
}

type UpgradeToDriverRequest struct {
	GovNumber          string  `json:"gov_number" binding:"required,max=20"`
	Brand              string  `json:"brand" binding:"required,max=50"`
	Model              string  `json:"model" binding:"required,max=50"`
	Color              string  `json:"color,omitempty" binding:"max=50"`
	Year               int32   `json:"year,omitempty"`
	AvatarURL          *string `json:"avatar_url,omitempty"`
	TechPassportPhoto  *string `json:"tech_passport_photo,omitempty"`
	DriverLicensePhoto *string `json:"driver_license_photo,omitempty"`
	TechPassportFront  *string `json:"tech_passport_front,omitempty"`
	TechPassportBack   *string `json:"tech_passport_back,omitempty"`
	DriverLicenseFront *string `json:"driver_license_front,omitempty"`
	DriverLicenseBack  *string `json:"driver_license_back,omitempty"`
	CarFrontPhoto      *string `json:"car_front_photo,omitempty"`
	CarBackPhoto       *string `json:"car_back_photo,omitempty"`
	CarSidePhoto       *string `json:"car_side_photo,omitempty"`
}

// ToCarDomain maps the DTO to the internal Domain structure
func (req *UpgradeToDriverRequest) ToCarDomain() *domain.Car {
	return &domain.Car{
		GovNumber:          req.GovNumber,
		Brand:              req.Brand,
		Model:              req.Model,
		Color:              req.Color,
		Year:               req.Year,
		TechPassportPhoto:  req.TechPassportPhoto,
		DriverLicensePhoto: req.DriverLicensePhoto,
		TechPassportFront:  req.TechPassportFront,
		TechPassportBack:   req.TechPassportBack,
		DriverLicenseFront: req.DriverLicenseFront,
		DriverLicenseBack:  req.DriverLicenseBack,
		CarFrontPhoto:      req.CarFrontPhoto,
		CarBackPhoto:       req.CarBackPhoto,
		CarSidePhoto:       req.CarSidePhoto,
	}
}

type CreateCarRequest struct {
	GovNumber          string  `json:"gov_number" binding:"required,max=20"`
	Brand              string  `json:"brand" binding:"required,max=50"`
	Model              string  `json:"model" binding:"required,max=50"`
	Color              string  `json:"color,omitempty" binding:"max=50"`
	Year               int32   `json:"year,omitempty"`
	TechPassportPhoto  *string `json:"tech_passport_photo,omitempty"`
	DriverLicensePhoto *string `json:"driver_license_photo,omitempty"`
	TechPassportFront  *string `json:"tech_passport_front,omitempty"`
	TechPassportBack   *string `json:"tech_passport_back,omitempty"`
	DriverLicenseFront *string `json:"driver_license_front,omitempty"`
	DriverLicenseBack  *string `json:"driver_license_back,omitempty"`
	CarFrontPhoto      *string `json:"car_front_photo,omitempty"`
	CarBackPhoto       *string `json:"car_back_photo,omitempty"`
	CarSidePhoto       *string `json:"car_side_photo,omitempty"`
}

func (req *CreateCarRequest) ToCarDomain() *domain.Car {
	return &domain.Car{
		GovNumber:          req.GovNumber,
		Brand:              req.Brand,
		Model:              req.Model,
		Color:              req.Color,
		Year:               req.Year,
		TechPassportPhoto:  req.TechPassportPhoto,
		DriverLicensePhoto: req.DriverLicensePhoto,
		TechPassportFront:  req.TechPassportFront,
		TechPassportBack:   req.TechPassportBack,
		DriverLicenseFront: req.DriverLicenseFront,
		DriverLicenseBack:  req.DriverLicenseBack,
		CarFrontPhoto:      req.CarFrontPhoto,
		CarBackPhoto:       req.CarBackPhoto,
		CarSidePhoto:       req.CarSidePhoto,
	}
}

type UpdateCarRequest struct {
	GovNumber string `json:"gov_number,omitempty"`
	Brand     string `json:"brand,omitempty"`
	Model     string `json:"model,omitempty"`
	Color     string `json:"color,omitempty"`
	Year      int32  `json:"year,omitempty"`
}

func (req *UpdateCarRequest) ToCarDomain() *domain.Car {
	return &domain.Car{
		GovNumber: req.GovNumber,
		Brand:     req.Brand,
		Model:     req.Model,
		Color:     req.Color,
		Year:      req.Year,
	}
}
