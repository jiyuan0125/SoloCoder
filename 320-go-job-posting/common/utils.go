package common

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func GenerateID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func ValidateCreateJobRequest(req CreateJobRequest) error {
	if req.Title == "" {
		return fmt.Errorf("职位名称不能为空")
	}
	if len(req.Title) > 50 {
		return fmt.Errorf("职位名称不能超过50字")
	}
	if req.MinSalary < 0 {
		return fmt.Errorf("最低薪资不能为负数")
	}
	if req.MaxSalary < 0 {
		return fmt.Errorf("最高薪资不能为负数")
	}
	if req.MinSalary > req.MaxSalary {
		return fmt.Errorf("最低薪资不能大于最高薪资")
	}
	if req.Description == "" {
		return fmt.Errorf("职位描述不能为空")
	}
	if len(req.Description) > 2000 {
		return fmt.Errorf("职位描述不能超过2000字")
	}
	if req.Education != "" {
		validEducations := map[EducationLevel]bool{
			EducationAny:      true,
			EducationCollege:  true,
			EducationBachelor: true,
			EducationMaster:   true,
			EducationDoctor:   true,
		}
		if !validEducations[req.Education] {
			return fmt.Errorf("无效的学历要求")
		}
	}
	return nil
}

func ValidateApplyJobRequest(req ApplyJobRequest) error {
	if req.Name == "" {
		return fmt.Errorf("姓名不能为空")
	}
	if req.Phone == "" {
		return fmt.Errorf("手机号不能为空")
	}
	return nil
}

func ValidateUpdateApplicationStatusRequest(req UpdateApplicationStatusRequest) error {
	validStatuses := map[ApplicationStatus]bool{
		StatusUnderReview: true,
		StatusInterview:   true,
		StatusNotSuitable: true,
	}
	if !validStatuses[req.Status] {
		return fmt.Errorf("无效的状态")
	}
	if req.Status == StatusInterview {
		if req.InterviewTime == nil || *req.InterviewTime == "" {
			return fmt.Errorf("面试邀约需要指定面试时间")
		}
		if req.InterviewLocation == nil || *req.InterviewLocation == "" {
			return fmt.Errorf("面试邀约需要指定面试地点")
		}
	}
	return nil
}
