package service

import (
	"abtest/api"
	"abtest/server/store"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
)

var (
	ErrExperimentNotFound     = errors.New("experiment not found")
	ErrExperimentAlreadyRunning = errors.New("experiment already running")
	ErrExperimentNotRunning   = errors.New("experiment not running")
	ErrExperimentArchived     = errors.New("experiment is archived")
	ErrCannotModifyTraffic    = errors.New("cannot modify traffic after experiment starts")
	ErrGrayMin24Hours         = errors.New("gray release requires 24 hours observation before adjusting traffic")
	ErrInsufficientSample     = errors.New("insufficient sample size, experiment needs to run at least 7 days")
	ErrInvalidPercentage      = errors.New("invalid percentage, must be 0-100 and sum to 100")
)

const (
	MinExperimentDurationDays = 7
	GrayInitialControlPct     = 95
	GrayInitialTreatmentPct   = 5
	GrayMinObservationHours   = 24
	ConfidenceLevel           = 0.95
)

type Service struct {
	store *store.Store
}

func NewService(s *store.Store) *Service {
	return &Service{store: s}
}

func (s *Service) CreateExperiment(req api.CreateExperimentRequest) (*api.Experiment, error) {
	if req.Type == api.ExperimentTypeGray {
		req.ControlPercentage = GrayInitialControlPct
		req.TreatmentPercentage = GrayInitialTreatmentPct
	}

	if !isValidPercentage(req.ControlPercentage, req.TreatmentPercentage) {
		return nil, ErrInvalidPercentage
	}

	exp := &api.Experiment{
		ID:                 generateID(),
		Name:               req.Name,
		Type:               req.Type,
		Status:             api.ExperimentStatusDraft,
		ControlPercentage:  req.ControlPercentage,
		TreatmentPercentage: req.TreatmentPercentage,
		CreatedAt:          time.Now(),
		MinDurationDays:    MinExperimentDurationDays,
	}

	if err := s.store.CreateExperiment(exp); err != nil {
		return nil, err
	}

	return exp, nil
}

func (s *Service) StartExperiment(experimentID string) (*api.Experiment, error) {
	exp, exists := s.store.GetExperiment(experimentID)
	if !exists {
		return nil, ErrExperimentNotFound
	}

	if exp.Status == api.ExperimentStatusArchived {
		return nil, ErrExperimentArchived
	}

	if exp.Status == api.ExperimentStatusRunning {
		return exp, nil
	}

	now := time.Now()
	exp.Status = api.ExperimentStatusRunning
	exp.StartedAt = &now

	if err := s.store.UpdateExperiment(exp); err != nil {
		return nil, err
	}

	return exp, nil
}

func (s *Service) EndExperiment(experimentID string) (*api.Experiment, error) {
	exp, exists := s.store.GetExperiment(experimentID)
	if !exists {
		return nil, ErrExperimentNotFound
	}

	if exp.Status == api.ExperimentStatusArchived {
		return nil, ErrExperimentArchived
	}

	if exp.Status != api.ExperimentStatusRunning {
		return nil, ErrExperimentNotRunning
	}

	now := time.Now()
	exp.Status = api.ExperimentStatusPaused
	exp.EndedAt = &now

	if err := s.store.UpdateExperiment(exp); err != nil {
		return nil, err
	}

	return exp, nil
}

func (s *Service) ArchiveExperiment(experimentID string) (*api.Experiment, error) {
	exp, exists := s.store.GetExperiment(experimentID)
	if !exists {
		return nil, ErrExperimentNotFound
	}

	if exp.Status == api.ExperimentStatusArchived {
		return exp, nil
	}

	if exp.Status == api.ExperimentStatusRunning {
		return nil, errors.New("cannot archive running experiment, end it first")
	}

	exp.Status = api.ExperimentStatusArchived

	if err := s.store.UpdateExperiment(exp); err != nil {
		return nil, err
	}

	return exp, nil
}

func (s *Service) UpdateTraffic(experimentID string, controlPct, treatmentPct int) (*api.Experiment, error) {
	exp, exists := s.store.GetExperiment(experimentID)
	if !exists {
		return nil, ErrExperimentNotFound
	}

	if exp.Status == api.ExperimentStatusArchived {
		return nil, ErrExperimentArchived
	}

	if !isValidPercentage(controlPct, treatmentPct) {
		return nil, ErrInvalidPercentage
	}

	if exp.Status == api.ExperimentStatusRunning {
		if exp.Type == api.ExperimentTypeNormal {
			return nil, ErrCannotModifyTraffic
		}

		if exp.Type == api.ExperimentTypeGray && exp.StartedAt != nil {
			elapsed := time.Since(*exp.StartedAt)
			if elapsed < time.Hour*time.Duration(GrayMinObservationHours) {
				return nil, ErrGrayMin24Hours
			}
		}
	}

	exp.ControlPercentage = controlPct
	exp.TreatmentPercentage = treatmentPct

	if err := s.store.UpdateExperiment(exp); err != nil {
		return nil, err
	}

	return exp, nil
}

func (s *Service) AssignUser(experimentID, userID string) (api.GroupType, error) {
	exp, exists := s.store.GetExperiment(experimentID)
	if !exists {
		return "", ErrExperimentNotFound
	}

	if exp.Status == api.ExperimentStatusArchived {
		return "", ErrExperimentArchived
	}

	if exp.Status != api.ExperimentStatusRunning {
		return "", ErrExperimentNotRunning
	}

	if group, exists := s.store.GetAssignment(experimentID, userID); exists {
		return group, nil
	}

	group := s.calculateGroup(userID, exp.ControlPercentage, exp.TreatmentPercentage)

	assignedGroup := s.store.GetOrCreateAssignment(experimentID, userID, func() api.GroupType {
		return group
	})

	return assignedGroup, nil
}

func (s *Service) calculateGroup(userID string, controlPct, treatmentPct int) api.GroupType {
	hash := sha256.Sum256([]byte(userID))
	value := binary.BigEndian.Uint64(hash[:8])
	normalized := float64(value) / float64(math.MaxUint64) * 100

	if normalized < float64(controlPct) {
		return api.GroupControl
	}
	return api.GroupTreatment
}

func (s *Service) RecordMetrics(experimentID, userID string, converted bool, stayDuration float64) error {
	group, exists := s.store.GetAssignment(experimentID, userID)
	if !exists {
		return errors.New("user not assigned to any group in this experiment")
	}

	s.store.RecordMetrics(experimentID, group, userID, converted, stayDuration)
	return nil
}

func (s *Service) GetStats(experimentID string) (*api.ExperimentStatsResponse, error) {
	exp, exists := s.store.GetExperiment(experimentID)
	if !exists {
		return nil, ErrExperimentNotFound
	}

	controlMetrics, treatmentMetrics := s.store.GetMetrics(experimentID)

	controlStats := calculateGroupStats(api.GroupControl, controlMetrics)
	treatmentStats := calculateGroupStats(api.GroupTreatment, treatmentMetrics)

	sampleSufficient := s.isSampleSufficient(exp)

	var significanceTests []api.SignificanceTest

	if len(controlMetrics) > 0 && len(treatmentMetrics) > 0 {
		convTest := s.testConversionSignificance(controlMetrics, treatmentMetrics)
		convTest.Metric = "conversion_rate"
		significanceTests = append(significanceTests, convTest)

		durationTest := s.testDurationSignificance(controlMetrics, treatmentMetrics)
		durationTest.Metric = "avg_stay_duration"
		significanceTests = append(significanceTests, durationTest)
	}

	return &api.ExperimentStatsResponse{
		Success:         true,
		Experiment:      *exp,
		ControlStats:    controlStats,
		TreatmentStats:  treatmentStats,
		Significance:    significanceTests,
		SampleSufficient: sampleSufficient,
	}, nil
}

func (s *Service) isSampleSufficient(exp *api.Experiment) bool {
	if exp.StartedAt == nil {
		return false
	}

	elapsed := time.Since(*exp.StartedAt)
	minDuration := time.Hour * 24 * time.Duration(exp.MinDurationDays)

	return elapsed >= minDuration
}

func (s *Service) GetUserAssignments(userID string) ([]api.UserAssignment, error) {
	assignments := s.store.GetUserAssignments(userID)

	var result []api.UserAssignment
	for expID, group := range assignments {
		if exp, exists := s.store.GetExperiment(expID); exists {
			result = append(result, api.UserAssignment{
				ExperimentID:   expID,
				ExperimentName: exp.Name,
				Group:          group,
			})
		}
	}

	return result, nil
}

func (s *Service) GetExperiment(experimentID string) (*api.Experiment, bool) {
	return s.store.GetExperiment(experimentID)
}

func (s *Service) ListExperiments(status *api.ExperimentStatus) []api.Experiment {
	return s.store.ListExperiments(status)
}

func isValidPercentage(control, treatment int) bool {
	if control < 0 || control > 100 || treatment < 0 || treatment > 100 {
		return false
	}
	return control+treatment == 100
}

func generateID() string {
	return uuid.New().String()
}

func calculateGroupStats(group api.GroupType, metrics []*store.UserMetrics) api.GroupStats {
	total := len(metrics)
	if total == 0 {
		return api.GroupStats{
			Group:          group,
			TotalUsers:     0,
			ConvertedUsers: 0,
			ConversionRate: 0,
			AvgStayDuration: 0,
		}
	}

	converted := 0
	totalDuration := 0.0

	for _, m := range metrics {
		if m.Converted {
			converted++
		}
		totalDuration += m.StayDuration
	}

	return api.GroupStats{
		Group:          group,
		TotalUsers:     total,
		ConvertedUsers: converted,
		ConversionRate: float64(converted) / float64(total),
		AvgStayDuration: totalDuration / float64(total),
	}
}

func (s *Service) testConversionSignificance(control, treatment []*store.UserMetrics) api.SignificanceTest {
	controlConverted := 0
	for _, m := range control {
		if m.Converted {
			controlConverted++
		}
	}

	treatmentConverted := 0
	for _, m := range treatment {
		if m.Converted {
			treatmentConverted++
		}
	}

	controlRate := float64(controlConverted) / float64(len(control))
	treatmentRate := float64(treatmentConverted) / float64(len(treatment))

	pooledP := float64(controlConverted+treatmentConverted) / float64(len(control)+len(treatment))
	se := math.Sqrt(pooledP * (1 - pooledP) * (1.0/float64(len(control)) + 1.0/float64(len(treatment))))

	if se == 0 {
		return api.SignificanceTest{
			IsSignificant:   false,
			PValue:          1.0,
			ConfidenceLevel: ConfidenceLevel,
		}
	}

	z := (treatmentRate - controlRate) / se
	pValue := normalDistributionCDF(-math.Abs(z)) * 2

	return api.SignificanceTest{
		IsSignificant:   pValue < (1 - ConfidenceLevel),
		PValue:          pValue,
		ConfidenceLevel: ConfidenceLevel,
	}
}

func (s *Service) testDurationSignificance(control, treatment []*store.UserMetrics) api.SignificanceTest {
	controlMean, controlVar := calculateMeanAndVariance(control)
	treatmentMean, treatmentVar := calculateMeanAndVariance(treatment)

	n1 := float64(len(control))
	n2 := float64(len(treatment))

	se := math.Sqrt(controlVar/n1 + treatmentVar/n2)
	if se == 0 {
		return api.SignificanceTest{
			IsSignificant:   false,
			PValue:          1.0,
			ConfidenceLevel: ConfidenceLevel,
		}
	}

	t := (treatmentMean - controlMean) / se
	df := (controlVar/n1 + treatmentVar/n2) * (controlVar/n1 + treatmentVar/n2) /
		((controlVar*controlVar)/(n1*n1*(n1-1)) + (treatmentVar*treatmentVar)/(n2*n2*(n2-1)))

	pValue := tDistributionCDF(-math.Abs(t), int(df)) * 2

	return api.SignificanceTest{
		IsSignificant:   pValue < (1 - ConfidenceLevel),
		PValue:          pValue,
		ConfidenceLevel: ConfidenceLevel,
	}
}

func calculateMeanAndVariance(metrics []*store.UserMetrics) (mean, variance float64) {
	if len(metrics) == 0 {
		return 0, 0
	}

	sum := 0.0
	for _, m := range metrics {
		sum += m.StayDuration
	}
	mean = sum / float64(len(metrics))

	sumSq := 0.0
	for _, m := range metrics {
		diff := m.StayDuration - mean
		sumSq += diff * diff
	}
	variance = sumSq / float64(len(metrics)-1)
	if len(metrics) <= 1 {
		variance = 0
	}

	return mean, variance
}

func normalDistributionCDF(x float64) float64 {
	const (
		a1 = 0.254829592
		a2 = -0.284496736
		a3 = 1.421413741
		a4 = -1.453152027
		a5 = 1.061405429
		p  = 0.3275911
	)

	sign := 1.0
	if x < 0 {
		sign = -1.0
		x = -x
	}

	t := 1.0 / (1.0 + p*x)
	y := 1.0 - (((((a5*t+a4)*t)+a3)*t+a2)*t+a1)*t*math.Exp(-x*x/2.0)

	return 0.5 * (1.0 + sign*y)
}

func tDistributionCDF(t float64, df int) float64 {
	if df <= 0 {
		return 0.5
	}

	if df >= 100 {
		return normalDistributionCDF(t)
	}

	x := float64(df) / (float64(df) + t*t)
	a := float64(df) / 2.0
	b := 0.5

	return 0.5 * (1 + sign(t)*(1-betaIncomplete(x, a, b)))
}

func betaIncomplete(x, a, b float64) float64 {
	const eps = 1e-10
	const maxIter = 100

	if x <= 0 {
		return 0
	}
	if x >= 1 {
		return 1
	}

	lgammaA, _ := math.Lgamma(a)
	lgammaB, _ := math.Lgamma(b)
	lgammaAB, _ := math.Lgamma(a + b)
	lbeta := lgammaA + lgammaB - lgammaAB
	bt := math.Exp(a*math.Log(x) + b*math.Log(1-x) - lbeta)

	var bt1, ap, am, bp, az, bz float64
	var i int
	var m float64

	if x < (a+1)/(a+b+2) {
		bt1 = bt / a
		ap = a
		bp = a + b + 1
		az = 1.0
		am = 1.0
		for i = 0; i < maxIter; i++ {
			m = float64(i + 1)
			bz = az + bt1
			if math.Abs(bz-az) < eps*math.Abs(bz) {
				return bz
			}
			ap += 1
			bp += 1
			bt1 *= m * (b - m) * x / ((ap - 1) * bp)
			az = bz + bt1
			if math.Abs(az-bz) < eps*math.Abs(az) {
				return az
			}
			bt1 *= m * (a - m) * x / (am * (bp - 1))
			am += 1
			ap += 1
			bp += 1
		}
		return az
	}

	bt1 = bt / b
	ap = b
	bp = a + b + 1
	az = 1.0
	am = 1.0
	for i = 0; i < maxIter; i++ {
		m = float64(i + 1)
		bz = az - bt1
		if math.Abs(bz-az) < eps*math.Abs(bz) {
			return 1 - bz
		}
		ap += 1
		bp += 1
		bt1 *= m * (a - m) * (1 - x) / ((ap - 1) * bp)
		az = bz - bt1
		if math.Abs(az-bz) < eps*math.Abs(az) {
			return 1 - az
		}
		bt1 *= m * (b - m) * (1 - x) / (am * (bp - 1))
		am += 1
		ap += 1
		bp += 1
	}
	return 1 - az
}

func sign(x float64) float64 {
	if x >= 0 {
		return 1
	}
	return -1
}

func (s *Service) GetExperimentRunningDuration(experimentID string) (time.Duration, error) {
	exp, exists := s.store.GetExperiment(experimentID)
	if !exists {
		return 0, ErrExperimentNotFound
	}

	if exp.StartedAt == nil {
		return 0, nil
	}

	if exp.EndedAt != nil {
		return exp.EndedAt.Sub(*exp.StartedAt), nil
	}

	return time.Since(*exp.StartedAt), nil
}

func (s *Service) GetRemainingHoursForGray(experimentID string) (float64, error) {
	exp, exists := s.store.GetExperiment(experimentID)
	if !exists {
		return 0, ErrExperimentNotFound
	}

	if exp.Type != api.ExperimentTypeGray {
		return 0, nil
	}

	if exp.StartedAt == nil {
		return float64(GrayMinObservationHours), nil
	}

	elapsed := time.Since(*exp.StartedAt)
	remaining := time.Hour*time.Duration(GrayMinObservationHours) - elapsed
	if remaining <= 0 {
		return 0, nil
	}

	return remaining.Hours(), nil
}

func (s *Service) IsExperimentModifiable(experimentID string) (bool, string) {
	exp, exists := s.store.GetExperiment(experimentID)
	if !exists {
		return false, "experiment not found"
	}

	if exp.Status == api.ExperimentStatusArchived {
		return false, "experiment is archived"
	}

	if exp.Status == api.ExperimentStatusRunning && exp.Type == api.ExperimentTypeNormal {
		return false, "cannot modify normal experiment after it starts"
	}

	if exp.Status == api.ExperimentStatusRunning && exp.Type == api.ExperimentTypeGray {
		if exp.StartedAt != nil {
			elapsed := time.Since(*exp.StartedAt)
			if elapsed < time.Hour*time.Duration(GrayMinObservationHours) {
				remaining := time.Hour*time.Duration(GrayMinObservationHours) - elapsed
				return false, fmt.Sprintf("gray release requires %.1f more hours observation", remaining.Hours())
			}
		}
	}

	return true, ""
}
