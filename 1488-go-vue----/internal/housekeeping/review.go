package housekeeping

import (
	"math"
	"sort"
)

func CalculateAverageRating(reviews []Review) float64 {
	if len(reviews) == 0 {
		return 5.0
	}

	sort.Slice(reviews, func(i, j int) bool {
		return reviews[i].Time.After(reviews[j].Time)
	})

	recentCount := 20
	if len(reviews) < recentCount {
		recentCount = len(reviews)
	}

	recentReviews := reviews[:recentCount]

	var weightedSum float64
	var totalWeight float64

	for i, review := range recentReviews {
		var weight float64
		if i < 5 {
			weight = 1.2
		} else {
			weight = 1.0
		}
		weightedSum += review.Rating * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 5.0
	}

	avg := weightedSum / totalWeight
	return math.Round(avg*10) / 10
}

func AddReview(aunt *Aunt, review Review) {
	aunt.mu.Lock()
	defer aunt.mu.Unlock()

	aunt.Reviews = append(aunt.Reviews, review)
	aunt.Rating = CalculateAverageRating(aunt.Reviews)
}
