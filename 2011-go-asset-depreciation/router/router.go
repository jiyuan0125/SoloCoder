package router

import (
	"asset-depreciation/handlers"
	"net/http"
	"regexp"
)

var (
	assetsPattern         = regexp.MustCompile(`^/api/assets/?$`)
	assetPattern          = regexp.MustCompile(`^/api/assets/(\d+)/?$`)
	assetStatusPattern    = regexp.MustCompile(`^/api/assets/(\d+)/status/?$`)
	assetHistoryPattern   = regexp.MustCompile(`^/api/assets/(\d+)/history/?$`)
	historyRemarkPattern  = regexp.MustCompile(`^/api/assets/(\d+)/history/(\d+)/remark/?$`)
	assetDeprPattern      = regexp.MustCompile(`^/api/assets/(\d+)/depreciation/?$`)
	assetScrapPattern     = regexp.MustCompile(`^/api/assets/(\d+)/scrap/?$`)
	departmentSummaryPattern = regexp.MustCompile(`^/api/departments/summary/?$`)
	deprTriggerPattern    = regexp.MustCompile(`^/api/depreciation/trigger/?$`)
)

func NewRouter() http.Handler {
	return http.HandlerFunc(routeRequest)
}

func routeRequest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"asset-depreciation"}`))
		return
	}

	if assetsPattern.MatchString(path) {
		handlers.AssetsHandler(w, r)
		return
	}

	if matches := assetPattern.FindStringSubmatch(path); matches != nil {
		handlers.AssetHandler(w, r)
		return
	}

	if matches := assetStatusPattern.FindStringSubmatch(path); matches != nil {
		handlers.AssetStatusHandler(w, r)
		return
	}

	if matches := assetHistoryPattern.FindStringSubmatch(path); matches != nil {
		handlers.AssetHistoryHandler(w, r)
		return
	}

	if matches := historyRemarkPattern.FindStringSubmatch(path); matches != nil {
		handlers.HistoryRemarkHandler(w, r)
		return
	}

	if matches := assetDeprPattern.FindStringSubmatch(path); matches != nil {
		handlers.AssetDepreciationHandler(w, r)
		return
	}

	if matches := assetScrapPattern.FindStringSubmatch(path); matches != nil {
		handlers.AssetScrapHandler(w, r)
		return
	}

	if departmentSummaryPattern.MatchString(path) {
		handlers.DepartmentSummaryHandler(w, r)
		return
	}

	if deprTriggerPattern.MatchString(path) {
		handlers.DepreciationTriggerHandler(w, r)
		return
	}

	http.NotFound(w, r)
}
