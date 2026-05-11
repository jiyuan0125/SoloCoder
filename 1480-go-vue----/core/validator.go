package core

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"vehicle-inspection/common"
)

var plateRegex = regexp.MustCompile(`^[\u4e00-\u9fa5][A-Z][A-Z0-9]{5,6}$`)

func ValidatePlateNumber(plate string) error {
	plate = strings.TrimSpace(plate)
	if plate == "" {
		return errors.New("车牌号不能为空")
	}
	if !plateRegex.MatchString(plate) {
		return errors.New("车牌号格式不正确，格式应为：省份简称+字母+5-6位字母数字")
	}
	return nil
}

func ValidateVehicleType(vt common.VehicleType) error {
	switch vt {
	case common.VehicleTypeSedan, common.VehicleTypeSUV, common.VehicleTypeMPV,
		common.VehicleTypeTruck, common.VehicleTypePassenger:
		return nil
	default:
		return errors.New("无效的车辆类型")
	}
}

func ValidateDateNotInPast(t time.Time) error {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	date := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	if date.Before(today) {
		return errors.New("日期不能是过去的时间")
	}
	return nil
}

func ValidateVehicle(v *common.Vehicle) error {
	if err := ValidatePlateNumber(v.PlateNumber); err != nil {
		return err
	}
	if err := ValidateVehicleType(v.VehicleType); err != nil {
		return err
	}
	if v.RegisterDate.IsZero() {
		return errors.New("注册日期不能为空")
	}
	return nil
}
