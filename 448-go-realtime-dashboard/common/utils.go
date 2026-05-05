package common

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"time"
)

func GenerateClientID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:12]
}

func GenerateAlertID() string {
	b := make([]byte, 12)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:16]
}

func EncodeMessage(msgType MessageType, payload interface{}) ([]byte, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	wsMsg := WebSocketMessage{
		Type:    msgType,
		Payload: payloadBytes,
	}

	return json.Marshal(wsMsg)
}

func DecodeMessage(data []byte) (*WebSocketMessage, error) {
	var msg WebSocketMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func IsWithinReconnectWindow(disconnectTime time.Time, now time.Time) bool {
	return now.Sub(disconnectTime) <= MaxReconnectWindow
}

func IsDataExpired(lastUpdate time.Time, now time.Time) bool {
	return now.Sub(lastUpdate) > ExpirationThreshold
}

func IsDataWithinRetention(timestamp time.Time, now time.Time) bool {
	retentionPeriod := time.Duration(DataRetentionDays) * 24 * time.Hour
	return now.Sub(timestamp) <= retentionPeriod
}

func CalculateTrendPoints(dataPoints []MetricDataPoint, now time.Time) []TrendDataPoint {
	if len(dataPoints) == 0 {
		return nil
	}

	startTime := now.Add(-TrendWindow)
	var filteredPoints []MetricDataPoint
	for _, dp := range dataPoints {
		if dp.Timestamp.After(startTime) {
			filteredPoints = append(filteredPoints, dp)
		}
	}

	if len(filteredPoints) == 0 {
		return nil
	}

	var trendPoints []TrendDataPoint
	for t := startTime; t.Before(now); t = t.Add(TrendInterval) {
		intervalEnd := t.Add(TrendInterval)
		var intervalPoints []MetricDataPoint
		for _, dp := range filteredPoints {
			if dp.Timestamp.After(t) && dp.Timestamp.Before(intervalEnd) {
				intervalPoints = append(intervalPoints, dp)
			}
		}

		if len(intervalPoints) > 0 {
			var sum float64
			for _, dp := range intervalPoints {
				sum += dp.Value
			}
			avg := sum / float64(len(intervalPoints))
			trendPoints = append(trendPoints, TrendDataPoint{
				Value:     avg,
				Timestamp: t,
				Period:    t.Format("2006-01-02 15:04"),
			})
		}
	}

	return trendPoints
}

func CalculateComparison(currentValue float64, yoYValue float64, moMValue float64) ComparisonData {
	result := ComparisonData{
		CurrentValue: currentValue,
	}

	if yoYValue != 0 {
		result.YoYPercent = ((currentValue - yoYValue) / yoYValue) * 100
		result.CompareValue = yoYValue
	}

	if moMValue != 0 {
		result.MoMPercent = ((currentValue - moMValue) / moMValue) * 100
	}

	return result
}
