package geolib

import "errors"

var (
	ErrInvalidLatitude  = errors.New("latitude must be between -90 and 90")
	ErrInvalidLongitude = errors.New("longitude must be between -180 and 180")
)

func ValidateLatitude(lat float64) error {
	if lat < -90 || lat > 90 {
		return ErrInvalidLatitude
	}
	return nil
}

func ValidateLongitude(lng float64) error {
	if lng < -180 || lng > 180 {
		return ErrInvalidLongitude
	}
	return nil
}

func ValidateCoordinate(lat, lng float64) error {
	if err := ValidateLatitude(lat); err != nil {
		return err
	}
	return ValidateLongitude(lng)
}
