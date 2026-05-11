package core

import (
	"encoding/csv"
	"io"
	"strconv"
	"time"
)

type ReportExporter struct {
	store   *Store
	stats   *StatsCalculator
}

func NewReportExporter(store *Store) *ReportExporter {
	return &ReportExporter{
		store: store,
		stats: NewStatsCalculator(store),
	}
}

func (re *ReportExporter) GenerateTaskReport(taskID string) (*TaskReport, error) {
	stats, err := re.stats.CalculateTaskStats(taskID)
	if err != nil {
		return nil, err
	}

	alerts, err := re.store.GetAlerts(taskID)
	if err != nil {
		return nil, err
	}

	alertCopies := make([]Alert, len(alerts))
	for i, a := range alerts {
		alertCopies[i] = *a
	}

	return &TaskReport{
		TaskID:     taskID,
		Stats:      *stats,
		Alerts:     alertCopies,
		ExportedAt: time.Now(),
	}, nil
}

func (re *ReportExporter) ExportSensorDataToCSV(taskID string, w io.Writer) error {
	data, err := re.store.GetSensorData(taskID)
	if err != nil {
		return err
	}

	writer := csv.NewWriter(w)
	defer writer.Flush()

	headers := []string{"timestamp", "temperature", "humidity", "latitude", "longitude", "speed"}
	if err := writer.Write(headers); err != nil {
		return err
	}

	for _, d := range data {
		row := []string{
			d.Timestamp.Format(time.RFC3339),
			strconv.FormatFloat(d.Temperature, 'f', 1, 64),
			strconv.FormatFloat(d.Humidity, 'f', 1, 64),
			strconv.FormatFloat(d.Latitude, 'f', 6, 64),
			strconv.FormatFloat(d.Longitude, 'f', 6, 64),
			strconv.FormatFloat(d.Speed, 'f', 1, 64),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

func (re *ReportExporter) ExportStatsToCSV(taskID string, w io.Writer) error {
	stats, err := re.stats.CalculateTaskStats(taskID)
	if err != nil {
		return err
	}

	writer := csv.NewWriter(w)
	defer writer.Flush()

	headers := []string{"metric", "value"}
	if err := writer.Write(headers); err != nil {
		return err
	}

	rows := [][]string{
		{"task_id", stats.TaskID},
		{"total_samples", strconv.Itoa(stats.TotalSamples)},
		{"valid_samples", strconv.Itoa(stats.ValidSamples)},
		{"pass_rate", strconv.FormatFloat(stats.PassRate*100, 'f', 2, 64) + "%"},
		{"need_recheck", strconv.FormatBool(stats.NeedRecheck)},
		{"temp_max", strconv.FormatFloat(stats.TempMax, 'f', 1, 64)},
		{"temp_min", strconv.FormatFloat(stats.TempMin, 'f', 1, 64)},
		{"temp_avg", strconv.FormatFloat(stats.TempAvg, 'f', 1, 64)},
		{"temp_stddev", strconv.FormatFloat(stats.TempStdDev, 'f', 2, 64)},
	}

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

func (re *ReportExporter) ExportAlertsToCSV(taskID string, w io.Writer) error {
	alerts, err := re.store.GetAlerts(taskID)
	if err != nil {
		return err
	}

	writer := csv.NewWriter(w)
	defer writer.Flush()

	headers := []string{"alert_id", "timestamp", "temperature", "level"}
	if err := writer.Write(headers); err != nil {
		return err
	}

	for _, a := range alerts {
		row := []string{
			a.ID,
			a.Timestamp.Format(time.RFC3339),
			strconv.FormatFloat(a.Temperature, 'f', 1, 64),
			string(a.Level),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}
