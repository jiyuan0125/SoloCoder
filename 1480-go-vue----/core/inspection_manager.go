package core

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"vehicle-inspection/common"
)

type InspectionManager struct {
	processes       map[string]*common.InspectionProcess
	reports         map[string]*common.InspectionReport
	appointmentMgr  *AppointmentManager
	vehicleManager  *VehicleManager
	mu              sync.RWMutex
	idGen           *IDGenerator
}

func NewInspectionManager(am *AppointmentManager, vm *VehicleManager) *InspectionManager {
	return &InspectionManager{
		processes:      make(map[string]*common.InspectionProcess),
		reports:        make(map[string]*common.InspectionReport),
		appointmentMgr: am,
		vehicleManager: vm,
		idGen:          NewIDGenerator(),
	}
}

func (im *InspectionManager) StartInspection(req *common.StartInspectionRequest) (*common.InspectionProcess, error) {
	if req.AppointmentID == "" {
		return nil, errors.New("预约ID不能为空")
	}

	appointment, err := im.appointmentMgr.GetAppointment(req.AppointmentID)
	if err != nil {
		return nil, err
	}

	vehicle, err := im.vehicleManager.GetVehicle(appointment.PlateNumber)
	if err != nil {
		return nil, err
	}

	baseFee, err := GetBaseFee(vehicle.VehicleType)
	if err != nil {
		return nil, err
	}

	process := &common.InspectionProcess{
		ID:            im.generateProcessID(),
		AppointmentID: req.AppointmentID,
		PlateNumber:   appointment.PlateNumber,
		CurrentStep:   common.StepAppearance,
		Steps:         []common.InspectionStepRecord{},
		Status:        "进行中",
		RecheckCount:  0,
		TotalFeeFen:   baseFee,
		CreatedAt:     time.Now(),
	}

	im.mu.Lock()
	im.processes[process.ID] = process
	im.mu.Unlock()

	return process, nil
}

func (im *InspectionManager) CompleteStep(req *common.CompleteStepRequest) (*common.InspectionProcess, error) {
	if req.ProcessID == "" {
		return nil, errors.New("检测流程ID不能为空")
	}

	im.mu.Lock()
	defer im.mu.Unlock()

	process, exists := im.processes[req.ProcessID]
	if !exists {
		return nil, errors.New("检测流程不存在")
	}

	if process.Status != "进行中" {
		return nil, errors.New("检测流程已完成或已取消")
	}

	currentStep := process.CurrentStep
	isRecheck := false

	for i := len(process.Steps) - 1; i >= 0; i-- {
		if process.Steps[i].Step == currentStep && process.Steps[i].Result == common.ResultRecheck {
			isRecheck = true
			break
		}
	}

	stepRecord := common.InspectionStepRecord{
		Step:        currentStep,
		Result:      req.Result,
		FailItems:   req.FailItems,
		CompletedAt: time.Now(),
		IsRecheck:   isRecheck,
	}

	process.Steps = append(process.Steps, stepRecord)

	if req.Result == common.ResultPass {
		nextStep := getNextStep(currentStep)
		if nextStep == "" {
			process.Status = "已完成"
			im.vehicleManager.UpdateLastInspectionDate(process.PlateNumber, time.Now())
			im.generateReport(process)
		} else {
			process.CurrentStep = nextStep
		}
	} else if req.Result == common.ResultRecheck {
		process.Status = "待复检"
	} else if req.Result == common.ResultFail {
		process.Status = "不合格"
	}

	return process, nil
}

func (im *InspectionManager) StartRecheck(req *common.StartRecheckRequest) (*common.InspectionProcess, error) {
	if req.ProcessID == "" {
		return nil, errors.New("检测流程ID不能为空")
	}

	im.mu.Lock()
	defer im.mu.Unlock()

	process, exists := im.processes[req.ProcessID]
	if !exists {
		return nil, errors.New("检测流程不存在")
	}

	if process.Status != "待复检" {
		return nil, errors.New("当前状态不允许开始复检")
	}

	var failedStep common.InspectionStep
	for i := len(process.Steps) - 1; i >= 0; i-- {
		if process.Steps[i].Result == common.ResultRecheck {
			failedStep = process.Steps[i].Step
			break
		}
	}

	if failedStep == "" {
		return nil, errors.New("未找到需要复检的环节")
	}

	process.RecheckCount++
	process.Status = "进行中"
	process.CurrentStep = failedStep

	vehicle, err := im.vehicleManager.GetVehicle(process.PlateNumber)
	if err != nil {
		return nil, err
	}

	recheckFee := CalculateRecheckFee(process.TotalFeeFen, process.RecheckCount)
	process.TotalFeeFen += recheckFee

	_ = vehicle
	return process, nil
}

func (im *InspectionManager) GetProcess(processID string) (*common.InspectionProcess, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	process, exists := im.processes[processID]
	if !exists {
		return nil, errors.New("检测流程不存在")
	}
	return process, nil
}

func (im *InspectionManager) GetReport(reportID string) (*common.InspectionReport, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	report, exists := im.reports[reportID]
	if !exists {
		return nil, errors.New("检测报告不存在")
	}
	return report, nil
}

func (im *InspectionManager) ListProcesses() []*common.InspectionProcess {
	im.mu.RLock()
	defer im.mu.RUnlock()

	processes := make([]*common.InspectionProcess, 0, len(im.processes))
	for _, p := range im.processes {
		processes = append(processes, p)
	}
	return processes
}

func (im *InspectionManager) generateReport(process *common.InspectionProcess) (*common.InspectionReport, error) {
	vehicle, err := im.vehicleManager.GetVehicle(process.PlateNumber)
	if err != nil {
		return nil, err
	}

	report := &common.InspectionReport{
		ID:             im.generateReportID(),
		ProcessID:      process.ID,
		PlateNumber:    process.PlateNumber,
		VehicleType:    vehicle.VehicleType,
		RegisterDate:   vehicle.RegisterDate,
		InspectionDate: time.Now(),
		Steps:          process.Steps,
		OverallResult:  "合格",
		TotalFeeFen:    process.TotalFeeFen,
		IssuedAt:       time.Now(),
	}

	im.reports[report.ID] = report
	return report, nil
}

func (im *InspectionManager) generateProcessID() string {
	ts := time.Now().UnixNano()
	return fmt.Sprintf("PROC-%d", ts)
}

func (im *InspectionManager) generateReportID() string {
	ts := time.Now().UnixNano()
	return fmt.Sprintf("RPT-%d", ts)
}

func getNextStep(current common.InspectionStep) common.InspectionStep {
	switch current {
	case common.StepAppearance:
		return common.StepExhaust
	case common.StepExhaust:
		return common.StepSafety
	case common.StepSafety:
		return common.StepChassis
	case common.StepChassis:
		return ""
	default:
		return ""
	}
}
