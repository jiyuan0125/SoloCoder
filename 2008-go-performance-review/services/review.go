package services

import (
	"database/sql"
	"fmt"
	"math"
	"sort"

	"github.com/performance-review/database"
	"github.com/performance-review/models"
)

func ValidateScore(score float64) bool {
	return score >= 1 && score <= 5
}

func CalculateTotalScore(quality, efficiency, collaboration, innovation float64) float64 {
	total := quality*0.4 + efficiency*0.3 + collaboration*0.2 + innovation*0.1
	return math.Round(total*100) / 100
}

func CalculateLevel(totalScore float64) string {
	if totalScore >= 4.5 {
		return "S"
	} else if totalScore >= 4.0 {
		return "A"
	} else if totalScore >= 3.5 {
		return "B"
	} else if totalScore >= 3.0 {
		return "C"
	}
	return "D"
}

func GetPreviousQuarter(quarterID int64) (int64, error) {
	var year, quarter int
	err := database.DB.QueryRow(`
		SELECT year, quarter FROM quarters WHERE id = ?
	`, quarterID).Scan(&year, &quarter)
	if err != nil {
		return 0, err
	}

	prevQuarter := quarter - 1
	prevYear := year
	if prevQuarter < 1 {
		prevQuarter = 4
		prevYear = year - 1
	}

	var prevQuarterID int64
	err = database.DB.QueryRow(`
		SELECT id FROM quarters WHERE year = ? AND quarter = ?
	`, prevYear, prevQuarter).Scan(&prevQuarterID)
	if err != nil {
		return 0, err
	}

	return prevQuarterID, nil
}

func CheckConsecutiveD(employeeID, currentQuarterID int64) (bool, error) {
	var currentLevel string
	err := database.DB.QueryRow(`
		SELECT final_level FROM reviews WHERE employee_id = ? AND quarter_id = ? AND status = 'confirmed'
	`, employeeID, currentQuarterID).Scan(&currentLevel)
	if err != nil {
		return false, nil
	}

	if currentLevel != "D" {
		return false, nil
	}

	prevQuarterID, err := GetPreviousQuarter(currentQuarterID)
	if err != nil {
		return false, nil
	}

	var prevLevel string
	err = database.DB.QueryRow(`
		SELECT final_level FROM reviews WHERE employee_id = ? AND quarter_id = ? AND status = 'confirmed'
	`, employeeID, prevQuarterID).Scan(&prevLevel)
	if err != nil {
		return false, nil
	}

	return prevLevel == "D", nil
}

func GetTeamMembersByQuarter(quarterID int64) (map[int64][]models.Review, error) {
	rows, err := database.DB.Query(`
		SELECT r.id, r.employee_id, r.quarter_id, r.quality, r.efficiency, r.collaboration, 
		       r.innovation, r.total_score, r.level, r.final_level, r.status, r.need_improvement,
		       e.team_id
		FROM reviews r
		JOIN employees e ON r.employee_id = e.id
		WHERE r.quarter_id = ?
	`, quarterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teamReviews := make(map[int64][]models.Review)
	for rows.Next() {
		var review models.Review
		var teamID int64
		err := rows.Scan(
			&review.ID, &review.EmployeeID, &review.QuarterID, &review.Quality, &review.Efficiency,
			&review.Collaboration, &review.Innovation, &review.TotalScore, &review.Level,
			&review.FinalLevel, &review.Status, &review.NeedImprovement, &teamID,
		)
		if err != nil {
			return nil, err
		}
		teamReviews[teamID] = append(teamReviews[teamID], review)
	}

	return teamReviews, nil
}

func ApplySRestriction(quarterID int64) error {
	teamReviews, err := GetTeamMembersByQuarter(quarterID)
	if err != nil {
		return err
	}

	for teamID, reviews := range teamReviews {
		sort.Slice(reviews, func(i, j int) bool {
			return reviews[i].TotalScore > reviews[j].TotalScore
		})

		totalMembers := len(reviews)
		maxS := int(math.Floor(float64(totalMembers) * 0.1))

		sCount := 0
		for i, review := range reviews {
			finalLevel := review.Level
			if review.Level == "S" {
				if sCount < maxS {
					sCount++
				} else {
					finalLevel = "A"
				}
			}

			_, err := database.DB.Exec(`
				UPDATE reviews SET final_level = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?
			`, finalLevel, review.ID)
			if err != nil {
				return err
			}

			reviews[i].FinalLevel = finalLevel
		}

		fmt.Printf("Team %d: Total %d, max S %d, actual S %d\n", teamID, totalMembers, maxS, sCount)
	}

	return nil
}

func GetEmployeeReviewsByYear(employeeID int64, year int) ([]models.Review, error) {
	rows, err := database.DB.Query(`
		SELECT r.id, r.employee_id, r.quarter_id, r.quality, r.efficiency, r.collaboration,
		       r.innovation, r.total_score, r.level, r.final_level, r.status, r.need_improvement,
		       q.quarter
		FROM reviews r
		JOIN quarters q ON r.quarter_id = q.id
		WHERE r.employee_id = ? AND q.year = ? AND r.status = 'confirmed'
	`, employeeID, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []models.Review
	quarterMap := make(map[int]bool)
	for rows.Next() {
		var review models.Review
		var quarter int
		err := rows.Scan(
			&review.ID, &review.EmployeeID, &review.QuarterID, &review.Quality, &review.Efficiency,
			&review.Collaboration, &review.Innovation, &review.TotalScore, &review.Level,
			&review.FinalLevel, &review.Status, &review.NeedImprovement, &quarter,
		)
		if err != nil {
			return nil, err
		}
		quarterMap[quarter] = true
		reviews = append(reviews, review)
	}

	if len(quarterMap) != 4 {
		return nil, fmt.Errorf("employee missing quarters")
	}

	return reviews, nil
}

func CalculateAnnualReview(employeeID int64, year int) (*models.AnnualReview, error) {
	reviews, err := GetEmployeeReviewsByYear(employeeID, year)
	if err != nil {
		return nil, err
	}

	var totalScore float64
	hasD := false
	for _, r := range reviews {
		totalScore += r.TotalScore
		if r.FinalLevel == "D" {
			hasD = true
		}
	}

	avgScore := math.Round((totalScore/4)*100) / 100
	level := CalculateLevel(avgScore)

	if hasD && (level == "S" || level == "A") {
		level = "B"
	}

	return &models.AnnualReview{
		EmployeeID:   employeeID,
		Year:         year,
		AverageScore: avgScore,
		Level:        level,
	}, nil
}

func GetQuarterInfo(quarterID int64) (int, int, error) {
	var year, quarter int
	err := database.DB.QueryRow(`
		SELECT year, quarter FROM quarters WHERE id = ?
	`, quarterID).Scan(&year, &quarter)
	return year, quarter, err
}

func EmployeeExists(employeeID int64) bool {
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM employees WHERE id = ?", employeeID).Scan(&count)
	return err == nil && count > 0
}

func QuarterExists(quarterID int64) bool {
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM quarters WHERE id = ?", quarterID).Scan(&count)
	return err == nil && count > 0
}

func GetDraftReview(employeeID, quarterID int64) (*models.Review, error) {
	var review models.Review
	err := database.DB.QueryRow(`
		SELECT id, employee_id, quarter_id, quality, efficiency, collaboration, innovation,
		       total_score, level, final_level, status, need_improvement
		FROM reviews WHERE employee_id = ? AND quarter_id = ?
	`, employeeID, quarterID).Scan(
		&review.ID, &review.EmployeeID, &review.QuarterID, &review.Quality, &review.Efficiency,
		&review.Collaboration, &review.Innovation, &review.TotalScore, &review.Level,
		&review.FinalLevel, &review.Status, &review.NeedImprovement,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &review, nil
}
