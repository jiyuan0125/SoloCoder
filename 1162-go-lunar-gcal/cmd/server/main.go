package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"go-lunar-gcal/pkg/api"
	"go-lunar-gcal/pkg/jieqi"
	"go-lunar-gcal/pkg/lunar"
)

func getPort() string {
	port := "8100"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}
	flagPort := flag.String("port", "", "服务器监听端口")
	flag.Parse()
	if *flagPort != "" {
		port = *flagPort
	}
	return port
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func solarToLunarHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
			"success": false,
			"error":   "仅支持 POST 请求",
		})
		return
	}

	var req api.SolarToLunarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "请求体解析失败: " + err.Error(),
		})
		return
	}

	if req.Date == "" {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "日期不能为空",
		})
		return
	}

	t, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "日期格式错误，应为 YYYY-MM-DD",
		})
		return
	}

	lunarDate, err := lunar.SolarToLunar(t)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	resp := api.SolarToLunarResponse{
		Success: true,
	}
	resp.Data.Year = lunarDate.Year
	resp.Data.Month = lunarDate.Month
	resp.Data.Day = lunarDate.Day
	resp.Data.IsLeapMonth = lunarDate.IsLeapMonth
	resp.Data.Ganzhi = lunarDate.YearGanzhi()
	resp.Data.Shengxiao = lunarDate.YearShengxiao()
	resp.Data.MonthName = lunarDate.MonthName()
	resp.Data.DayName = lunarDate.DayName()
	resp.Data.FullText = lunarDate.String()

	writeJSON(w, http.StatusOK, resp)
}

func lunarToSolarHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
			"success": false,
			"error":   "仅支持 POST 请求",
		})
		return
	}

	var req api.LunarToSolarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "请求体解析失败: " + err.Error(),
		})
		return
	}

	var lunarDate *lunar.LunarDate
	var err error

	if req.MonthStr != "" && req.DayStr != "" {
		lunarDate, err = lunar.ParseLunarDate(strconv.Itoa(req.Year), req.MonthStr, req.DayStr)
	} else {
		lunarDate = &lunar.LunarDate{
			Year:        req.Year,
			Month:       req.Month,
			Day:         req.Day,
			IsLeapMonth: req.IsLeapMonth,
		}
	}

	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	solarDate, err := lunar.LunarToSolar(*lunarDate)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	resp := api.LunarToSolarResponse{
		Success: true,
	}
	solarDate = solarDate.In(time.FixedZone("CST", 8*3600))
	resp.Data.Year = solarDate.Year()
	resp.Data.Month = int(solarDate.Month())
	resp.Data.Day = solarDate.Day()
	resp.Data.DateStr = solarDate.Format("2006-01-02")
	resp.Data.Weekday = api.WeekdayCN(int(solarDate.Weekday()))
	resp.Data.FullText = fmt.Sprintf("%s %s", solarDate.Format("2006年01月02日"), resp.Data.Weekday)

	writeJSON(w, http.StatusOK, resp)
}

func jieqiHandler(w http.ResponseWriter, r *http.Request) {
	var year int
	var err error

	if r.Method == http.MethodPost {
		var req api.JieqiRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"error":   "请求体解析失败: " + err.Error(),
			})
			return
		}
		year = req.Year
	} else if r.Method == http.MethodGet {
		yearStr := r.URL.Query().Get("year")
		if yearStr == "" {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"error":   "年份不能为空",
			})
			return
		}
		year, err = strconv.Atoi(yearStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"error":   "年份格式错误",
			})
			return
		}
	} else {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
			"success": false,
			"error":   "仅支持 GET 或 POST 请求",
		})
		return
	}

	jieqiList, err := jieqi.GetJieqi(year)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	items := make([]api.JieqiItem, 0, 24)
	for _, jq := range jieqiList {
		t := jq.Time.In(time.FixedZone("CST", 8*3600))
		items = append(items, api.JieqiItem{
			Name:     jq.Name,
			Date:     t.Format("2006-01-02"),
			DateTime: t.Format("2006-01-02 15:04"),
			Year:     t.Year(),
			Month:    int(t.Month()),
			Day:      t.Day(),
			Hour:     t.Hour(),
			Minute:   t.Minute(),
		})
	}

	resp := api.JieqiResponse{
		Success: true,
		Data:    items,
	}

	writeJSON(w, http.StatusOK, resp)
}

func main() {
	port := getPort()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/solar-to-lunar", solarToLunarHandler)
	mux.HandleFunc("/api/lunar-to-solar", lunarToSolarHandler)
	mux.HandleFunc("/api/jieqi", jieqiHandler)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"name":    "农历公历互转服务",
			"version": "1.0.0",
			"endpoints": map[string]string{
				"/api/solar-to-lunar": "POST - 公历转农历",
				"/api/lunar-to-solar": "POST - 农历转公历",
				"/api/jieqi":          "GET/POST - 二十四节气",
			},
		})
	})

	addr := ":" + port
	fmt.Printf("服务器启动中，监听 %s\n", addr)
	fmt.Printf("可用接口:\n")
	fmt.Printf("  POST %s/api/solar-to-lunar - 公历转农历\n", addr)
	fmt.Printf("  POST %s/api/lunar-to-solar - 农历转公历\n", addr)
	fmt.Printf("  GET/POST %s/api/jieqi?year=YYYY - 二十四节气\n", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "服务器启动失败: %v\n", err)
		os.Exit(1)
	}
}
