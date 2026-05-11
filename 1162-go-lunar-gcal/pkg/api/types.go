package api

type SolarToLunarRequest struct {
	Date string `json:"date"`
}

type SolarToLunarResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Data    struct {
		Year        int    `json:"year"`
		Month       int    `json:"month"`
		Day         int    `json:"day"`
		IsLeapMonth bool   `json:"is_leap_month"`
		Ganzhi      string `json:"ganzhi"`
		Shengxiao   string `json:"shengxiao"`
		MonthName   string `json:"month_name"`
		DayName     string `json:"day_name"`
		FullText    string `json:"full_text"`
	} `json:"data,omitempty"`
}

type LunarToSolarRequest struct {
	Year        int    `json:"year"`
	Month       int    `json:"month"`
	Day         int    `json:"day"`
	IsLeapMonth bool   `json:"is_leap_month"`
	MonthStr    string `json:"month_str,omitempty"`
	DayStr      string `json:"day_str,omitempty"`
}

type LunarToSolarResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Data    struct {
		Year     int    `json:"year"`
		Month    int    `json:"month"`
		Day      int    `json:"day"`
		DateStr  string `json:"date_str"`
		Weekday  string `json:"weekday"`
		FullText string `json:"full_text"`
	} `json:"data,omitempty"`
}

type JieqiRequest struct {
	Year int `json:"year"`
}

type JieqiItem struct {
	Name      string `json:"name"`
	Date      string `json:"date"`
	DateTime  string `json:"datetime"`
	Year      int    `json:"year"`
	Month     int    `json:"month"`
	Day       int    `json:"day"`
	Hour      int    `json:"hour"`
	Minute    int    `json:"minute"`
}

type JieqiResponse struct {
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Data    []JieqiItem `json:"data,omitempty"`
}

var weekdays = []string{"星期日", "星期一", "星期二", "星期三", "星期四", "星期五", "星期六"}

func WeekdayCN(w int) string {
	if w < 0 || w > 6 {
		return ""
	}
	return weekdays[w]
}
