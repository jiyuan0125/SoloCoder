package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"recruitment/internal/services/candidate"
	"recruitment/internal/services/interview"
	"recruitment/internal/services/job"
	"recruitment/internal/services/offer"
	"recruitment/internal/store"
)

type Dependencies struct {
	Store        *store.Store
	JobSvc       *job.Service
	CandidateSvc *candidate.Service
	InterviewSvc *interview.Service
	OfferSvc     *offer.Service
}

func NewRootCmd(dataDir string, calendarDir string) *cobra.Command {
	deps, err := initDependencies(dataDir, calendarDir)
	if err != nil {
		fmt.Printf("Error initializing dependencies: %v\n", err)
	}

	rootCmd := &cobra.Command{
		Use:   "recruitment",
		Short: "招聘管理系统",
		Long: `招聘管理系统 - 管理职位发布、候选人投递、面试安排和 Offer 发放

数据存储: JSON 文件
配置目录: ` + dataDir,
	}

	rootCmd.AddCommand(NewJobCmd(deps))
	rootCmd.AddCommand(NewCandidateCmd(deps))
	rootCmd.AddCommand(NewInterviewCmd(deps))
	rootCmd.AddCommand(NewOfferCmd(deps))
	rootCmd.AddCommand(NewValidateCmd(deps))

	return rootCmd
}

func initDependencies(dataDir string, calendarDir string) (*Dependencies, error) {
	s, err := store.NewStore(dataDir)
	if err != nil {
		return nil, err
	}

	jobSvc := job.NewService(s)
	candidateSvc := candidate.NewService(s, jobSvc)
	interviewSvc := interview.NewService(s, candidateSvc, calendarDir)
	offerSvc := offer.NewService(s, candidateSvc, interviewSvc)

	return &Dependencies{
		Store:        s,
		JobSvc:       jobSvc,
		CandidateSvc: candidateSvc,
		InterviewSvc: interviewSvc,
		OfferSvc:     offerSvc,
	}, nil
}
