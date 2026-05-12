package metrics

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Aggregation string

const (
	AggSum   Aggregation = "sum"
	AggAvg   Aggregation = "avg"
	AggMax   Aggregation = "max"
	AggMin   Aggregation = "min"
	AggCount Aggregation = "count"
)

type Aggregates struct {
	Count int64   `json:"count"`
	Sum   float64 `json:"sum"`
	Avg   float64 `json:"avg"`
	Max   float64 `json:"max"`
	Min   float64 `json:"min"`
}

type DataPoint struct {
	Time  string  `json:"time"`
	Count int64   `json:"count"`
	Sum   float64 `json:"sum"`
	Avg   float64 `json:"avg"`
	Max   float64 `json:"max"`
	Min   float64 `json:"min"`
}

func parseLabelsFromQuery(query string) Labels {
	labels := make(Labels)
	if query == "" {
		return labels
	}
	pairs := strings.Split(query, ",")
	for _, p := range pairs {
		kv := strings.SplitN(p, "=", 2)
		if len(kv) == 2 {
			labels[kv[0]] = kv[1]
		}
	}
	return labels
}

func parseTimeRange(c *fiber.Ctx) (time.Time, time.Time) {
	now := time.Now()
	end := now

	durationStr := c.Query("duration", "1h")
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		duration = time.Hour
	}

	if startStr := c.Query("start"); startStr != "" {
		if start, err := time.Parse(time.RFC3339, startStr); err == nil {
			startTime := start
			if endStr := c.Query("end"); endStr != "" {
				if e, err := time.Parse(time.RFC3339, endStr); err == nil {
					end = e
				}
			}
			return startTime, end
		}
	}

	return now.Add(-duration), end
}

func calculateAggregates(buckets []*Bucket) Aggregates {
	if len(buckets) == 0 {
		return Aggregates{}
	}

	agg := Aggregates{
		Min: buckets[0].Min,
		Max: buckets[0].Max,
	}

	for _, b := range buckets {
		agg.Count += b.Count
		agg.Sum += b.Sum
		if b.Min < agg.Min {
			agg.Min = b.Min
		}
		if b.Max > agg.Max {
			agg.Max = b.Max
		}
	}

	if agg.Count > 0 {
		agg.Avg = agg.Sum / float64(agg.Count)
	}

	return agg
}

func buildDataPoints(buckets []*Bucket) []DataPoint {
	points := make([]DataPoint, 0, len(buckets))
	for _, b := range buckets {
		points = append(points, DataPoint{
			Time:  b.Start.Format(time.RFC3339),
			Count: b.Count,
			Sum:   b.Sum,
			Avg:   b.Avg(),
			Max:   b.Max,
			Min:   b.Min,
		})
	}
	return points
}

func HandleListMetrics(c *fiber.Ctx) error {
	metricsList := DefaultRegistry.List()
	result := make([]map[string]interface{}, 0, len(metricsList))

	for _, m := range metricsList {
		seriesList := m.ListSeries()
		result = append(result, map[string]interface{}{
			"name":          m.Name,
			"type":          m.Type,
			"help":          m.Help,
			"series_count":  len(seriesList),
			"labels":        seriesList,
		})
	}

	return c.JSON(fiber.Map{
		"metrics": result,
		"count":   len(result),
	})
}

func HandleGetMetric(c *fiber.Ctx) error {
	name := c.Params("name")
	m, ok := DefaultRegistry.Get(name)
	if !ok {
		return c.Status(404).JSON(fiber.Map{
			"error": "Metric not found",
		})
	}

	labelsFilter := parseLabelsFromQuery(c.Query("labels"))
	start, end := parseTimeRange(c)

	seriesData := m.Query(labelsFilter, start, end)
	aggParam := c.Query("agg", "")

	seriesList := make([]map[string]interface{}, 0, len(seriesData))

	for labelsKey, buckets := range seriesData {
		labels := parseLabelsKey(labelsKey)
		aggregates := calculateAggregates(buckets)

		if aggParam != "" {
			var value float64
			switch Aggregation(aggParam) {
			case AggSum:
				value = aggregates.Sum
			case AggAvg:
				value = aggregates.Avg
			case AggMax:
				value = aggregates.Max
			case AggMin:
				value = aggregates.Min
			case AggCount:
				value = float64(aggregates.Count)
			default:
				value = aggregates.Sum
			}

			seriesList = append(seriesList, map[string]interface{}{
				"labels": labels,
				"value":  value,
				"agg":    aggParam,
			})
		} else {
			seriesList = append(seriesList, map[string]interface{}{
				"labels":     labels,
				"data_points": buildDataPoints(buckets),
				"aggregates": aggregates,
			})
		}
	}

	return c.JSON(fiber.Map{
		"name":   m.Name,
		"type":   m.Type,
		"help":   m.Help,
		"start":  start.Format(time.RFC3339),
		"end":    end.Format(time.RFC3339),
		"series": seriesList,
	})
}

func RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/metrics")
	api.Get("/", HandleListMetrics)
	api.Get("/:name", HandleGetMetric)
}
