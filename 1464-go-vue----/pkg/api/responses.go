package api

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type DeviceResponse struct {
	Device *Device `json:"device"`
}

type DevicesResponse struct {
	Devices []Device `json:"devices"`
}

type InspectionPointResponse struct {
	Point *InspectionPoint `json:"point"`
}

type InspectionPointsResponse struct {
	Points []InspectionPoint `json:"points"`
}

type InspectionRouteResponse struct {
	Route *InspectionRoute `json:"route"`
}

type InspectionRoutesResponse struct {
	Routes []InspectionRoute `json:"routes"`
}

type InspectionPlanResponse struct {
	Plan *InspectionPlan `json:"plan"`
}

type InspectionPlansResponse struct {
	Plans []InspectionPlan `json:"plans"`
}

type InspectionTaskResponse struct {
	Task *InspectionTask `json:"task"`
}

type InspectionTasksResponse struct {
	Tasks []InspectionTask `json:"tasks"`
}

type DrillPlanResponse struct {
	Plan *DrillPlan `json:"plan"`
}

type DrillPlansResponse struct {
	Plans []DrillPlan `json:"plans"`
}

type DrillRecordResponse struct {
	Record *DrillRecord `json:"record"`
}

type DrillRecordsResponse struct {
	Records []DrillRecord `json:"records"`
}

type ReminderResponse struct {
	Reminder *Reminder `json:"reminder"`
}

type RemindersResponse struct {
	Reminders []Reminder `json:"reminders"`
}

type TaskSummary struct {
	TaskID        string  `json:"task_id"`
	TotalPoints   int     `json:"total_points"`
	CheckedPoints int     `json:"checked_points"`
	NormalPoints  int     `json:"normal_points"`
	PassRate      float64 `json:"pass_rate"`
	AbnormalItems []AbnormalItem `json:"abnormal_items"`
}

type AbnormalItem struct {
	PointID   string `json:"point_id"`
	PointName string `json:"point_name"`
	Location  string `json:"location"`
}

type IDResponse struct {
	ID string `json:"id"`
}
