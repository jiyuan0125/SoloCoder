package export

import (
	"encoding/csv"
	"io"
	"strconv"
	"time"

	"quality-trace/pkg/common"
	"quality-trace/pkg/core/batch"
	"quality-trace/pkg/core/inspection"
	"quality-trace/pkg/core/process"
)

type Service struct {
	batchSvc     *batch.Service
	processSvc   *process.Service
	inspectionSvc *inspection.Service
}

func NewService(b *batch.Service, p *process.Service, i *inspection.Service) *Service {
	return &Service{
		batchSvc:     b,
		processSvc:   p,
		inspectionSvc: i,
	}
}

func formatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

func formatOptionalTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return formatTime(*t)
}

func (s *Service) ExportBatches(w io.Writer, filter common.ListBatchesRequest) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	header := []string{
		"batch_id", "product_name", "factory_code", "plan_quantity",
		"actual_quantity", "start_time", "end_time", "status",
		"is_abnormal", "is_one_time_pass", "rework_count",
		"created_at", "updated_at",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	batches := s.batchSvc.ListBatches(filter)
	for _, b := range batches {
		row := []string{
			b.BatchID,
			b.ProductName,
			b.FactoryCode,
			strconv.Itoa(b.PlanQuantity),
			strconv.Itoa(b.ActualQuantity),
			formatTime(b.StartTime),
			formatOptionalTime(b.EndTime),
			string(b.Status),
			strconv.FormatBool(b.IsAbnormal),
			strconv.FormatBool(b.IsOneTimePass),
			strconv.Itoa(b.ReworkCount),
			formatTime(b.CreatedAt),
			formatTime(b.UpdatedAt),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) ExportProcessRecords(w io.Writer, batchID string) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	header := []string{
		"id", "batch_id", "sequence", "process_name", "operator",
		"start_time", "end_time", "result", "is_rework",
		"created_at", "updated_at",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	var records []*common.ProcessRecord
	if batchID != "" {
		records = s.processSvc.ListProcessRecords(batchID)
	} else {
		records = s.processSvc.GetAllProcessRecords()
	}

	for _, r := range records {
		row := []string{
			strconv.FormatInt(r.ID, 10),
			r.BatchID,
			strconv.Itoa(r.Sequence),
			r.ProcessName,
			r.Operator,
			formatTime(r.StartTime),
			formatOptionalTime(r.EndTime),
			string(r.Result),
			strconv.FormatBool(r.IsRework),
			formatTime(r.CreatedAt),
			formatTime(r.UpdatedAt),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) ExportInspectionRecords(w io.Writer, filter common.ListInspectionRecordsRequest) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	header := []string{
		"id", "batch_id", "inspection_type", "metric_name",
		"actual_value", "lower_limit", "upper_limit", "result",
		"inspected_by", "inspected_at", "created_at",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	records := s.inspectionSvc.ListInspectionRecords(filter)
	for _, r := range records {
		row := []string{
			strconv.FormatInt(r.ID, 10),
			r.BatchID,
			string(r.InspectionType),
			r.MetricName,
			strconv.FormatFloat(r.ActualValue, 'f', -1, 64),
			strconv.FormatFloat(r.LowerLimit, 'f', -1, 64),
			strconv.FormatFloat(r.UpperLimit, 'f', -1, 64),
			string(r.Result),
			r.InspectedBy,
			formatTime(r.InspectedAt),
			formatTime(r.CreatedAt),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}
