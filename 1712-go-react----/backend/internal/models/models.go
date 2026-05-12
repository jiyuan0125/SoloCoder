package models

import "time"

type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
)

type BloodType string

const (
	BloodTypeA  BloodType = "A"
	BloodTypeB  BloodType = "B"
	BloodTypeAB BloodType = "AB"
	BloodTypeO  BloodType = "O"
)

type Allergy struct {
	Allergen   string `json:"allergen"`
	Reaction   string `json:"reaction"`
}

type Guardian struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type Resident struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Gender          Gender    `json:"gender"`
	IDCard          string    `json:"idCard"`
	BirthDate       time.Time `json:"birthDate"`
	Phone           string    `json:"phone"`
	Address         string    `json:"address"`
	EmergencyContact string   `json:"emergencyContact"`
	CreateDate      time.Time `json:"createDate"`
	DoctorInCharge  string    `json:"doctorInCharge"`
	BloodType       BloodType `json:"bloodType"`
	Allergies       []Allergy `json:"allergies"`
	Guardian        *Guardian `json:"guardian,omitempty"`
	Community       string    `json:"community"`
	FamilyID        string    `json:"familyID,omitempty"`
}

type Checkup struct {
	ID            string    `json:"id"`
	ResidentID    string    `json:"residentID"`
	Date          time.Time `json:"date"`
	Height        float64   `json:"height"`
	Weight        float64   `json:"weight"`
	BMI           float64   `json:"bmi"`
	SystolicBP    int       `json:"systolicBP"`
	DiastolicBP   int       `json:"diastolicBP"`
	HeartRate     int       `json:"heartRate"`
	VisionLeft    float64   `json:"visionLeft"`
	VisionRight   float64   `json:"visionRight"`
	IsInitial     bool      `json:"isInitial"`
}

type ChronicDisease string

const (
	ChronicHypertension  ChronicDisease = "hypertension"
	ChronicDiabetes      ChronicDisease = "diabetes"
	ChronicCoronary      ChronicDisease = "coronary"
	ChronicStroke        ChronicDisease = "stroke"
	ChronicCOPD          ChronicDisease = "copd"
)

type FollowupHypertension struct {
	SystolicBP  int    `json:"systolicBP"`
	DiastolicBP int    `json:"diastolicBP"`
	Medication  string `json:"medication"`
}

type FollowupDiabetes struct {
	FastingBloodSugar    float64 `json:"fastingBloodSugar"`
	PostprandialBloodSugar float64 `json:"postprandialBloodSugar"`
	HbA1c                float64 `json:"hba1c"`
	Medication           string  `json:"medication"`
}

type FollowupCoronary struct {
	CardiacFunction string `json:"cardiacFunction"`
	Medication      string `json:"medication"`
}

type FollowupRecord struct {
	ID             string        `json:"id"`
	ResidentID     string        `json:"residentID"`
	Disease        ChronicDisease `json:"disease"`
	Date           time.Time     `json:"date"`
	NextDueDate    time.Time     `json:"nextDueDate"`
	Hypertension   *FollowupHypertension `json:"hypertension,omitempty"`
	Diabetes       *FollowupDiabetes     `json:"diabetes,omitempty"`
	Coronary       *FollowupCoronary     `json:"coronary,omitempty"`
	IsOverdue      bool          `json:"isOverdue"`
}

type FamilyMember struct {
	ResidentID string `json:"residentID"`
	Relation   string `json:"relation"`
}

type Family struct {
	ID             string         `json:"id"`
	FamilyNumber   string         `json:"familyNumber"`
	Address        string         `json:"address"`
	Contracted     bool           `json:"contracted"`
	Members        []FamilyMember `json:"members"`
}

type CommunityStats struct {
	TotalCount    int            `json:"totalCount"`
	MaleCount     int            `json:"maleCount"`
	FemaleCount   int            `json:"femaleCount"`
	AgeGroups     map[string]int `json:"ageGroups"`
}

type ChronicPatientExport struct {
	ResidentID        string    `json:"residentID"`
	Name              string    `json:"name"`
	Disease           string    `json:"disease"`
	LastFollowupDate  time.Time `json:"lastFollowupDate"`
	LastFollowupData  string    `json:"lastFollowupData"`
}

type CompleteArchiveExport struct {
	Resident    Resident         `json:"resident"`
	Checkups    []Checkup        `json:"checkups"`
	Followups   []FollowupRecord `json:"followups"`
}
