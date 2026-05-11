package testdata

//go:enum Status
const (
	// 未知状态
	StatusUnknown Status = 0
	// 活跃状态
	StatusActive Status = 1
	// 停用状态
	StatusInactive Status = 2
	_                Status = 99 // 保留
	// 删除状态
	StatusDeleted Status = 100
)

//go:enum Gender
const (
	GenderInvalid Gender = -1
	GenderMale    Gender = 0
	GenderFemale  Gender = 1
	GenderOther   Gender = 2
)

//go:enum DeviceStatus
const (
	DeviceOnline  DeviceStatus = 1
	DeviceAlive   DeviceStatus = 1
	DeviceOffline DeviceStatus = 0
)

//go:enum LargeValue
const (
	LargeValue1 LargeValue = 2147483647
	LargeValue2 LargeValue = 2147483648
)
