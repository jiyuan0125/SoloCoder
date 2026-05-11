package main

import (
	"strconv"

	"envmonitor/internal/common"
	"envmonitor/internal/core"
)

func strconvToInt(s string) (int, error) {
	return strconv.Atoi(s)
}

func convertGasOutlet(outlet *core.GasOutlet) *common.GasOutletResponse {
	return &common.GasOutletResponse{
		ID:       outlet.ID,
		Location: outlet.Location,
		Status:   int(outlet.Status),
	}
}

func convertMeasurement(m core.Measurement) common.MeasurementResponse {
	return common.MeasurementResponse{
		Value:      m.Value,
		Factor:     m.Factor,
		Unit:       m.Unit,
		IsValid:    m.IsValid,
		IsAbnormal: m.IsAbnormal,
		IsExceeded: m.IsExceeded,
	}
}

func convertGasReport(report *core.GasReport) *common.GasReportResponse {
	measurements := make(map[string]common.MeasurementResponse, len(report.Measurements))
	for k, v := range report.Measurements {
		measurements[k] = convertMeasurement(v)
	}

	return &common.GasReportResponse{
		ID:              report.ID,
		OutletID:        report.OutletID,
		ReportedAt:      report.ReportedAt,
		ReceivedAt:      report.ReceivedAt,
		DataStatus:      int(report.DataStatus),
		Measurements:    measurements,
		ExceededFactors: report.ExceededFactors,
		AbnormalFactors: report.AbnormalFactors,
	}
}

func convertWastewaterReport(report *core.WastewaterReport) *common.WastewaterReportResponse {
	measurements := make(map[string]common.MeasurementResponse, len(report.Measurements))
	for k, v := range report.Measurements {
		measurements[k] = convertMeasurement(v)
	}

	return &common.WastewaterReportResponse{
		ID:              report.ID,
		OutletID:        report.OutletID,
		ReportedAt:      report.ReportedAt,
		ReceivedAt:      report.ReceivedAt,
		DataStatus:      int(report.DataStatus),
		Measurements:    measurements,
		ExceededFactors: report.ExceededFactors,
		AbnormalFactors: report.AbnormalFactors,
	}
}

func convertDailyReport(report *core.DailyReport) *common.DailyReportResponse {
	data := make(map[string]common.DailyFactorDataResponse, len(report.Data))
	for k, v := range report.Data {
		data[k] = common.DailyFactorDataResponse{
			Max:      v.Max,
			Min:      v.Min,
			Average:  v.Average,
			Count:    v.Count,
			Unit:     v.Unit,
			Exceeded: v.Exceeded,
		}
	}

	return &common.DailyReportResponse{
		Date:     report.Date,
		OutletID: report.OutletID,
		Status:   int(report.Status),
		Data:     data,
	}
}

func convertSolidWasteRecord(record *core.SolidWasteRecord) *common.SolidWasteRecordResponse {
	return &common.SolidWasteRecordResponse{
		ID:              record.ID,
		Name:            record.Name,
		Category:        int(record.Category),
		Amount:          record.Amount,
		StorageLocation: record.StorageLocation,
		DisposalMethod:  record.DisposalMethod,
		GeneratedAt:     record.GeneratedAt,
		DisposedAt:      record.DisposedAt,
		Status:          int(record.Status),
	}
}

func convertAlarm(alarm *core.Alarm) *common.AlarmResponse {
	return &common.AlarmResponse{
		ID:          alarm.ID,
		Type:        int(alarm.Type),
		Level:       int(alarm.Level),
		EntityType:  int(alarm.EntityType),
		EntityID:    alarm.EntityID,
		RelatedData: alarm.RelatedData,
		Message:     alarm.Message,
		CreatedAt:   alarm.CreatedAt,
		ResolvedAt:  alarm.ResolvedAt,
		IsResolved:  alarm.IsResolved,
	}
}
