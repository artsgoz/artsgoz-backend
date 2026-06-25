package utils

import (
	"strconv"
	"strings"

	"github.com/artsgoz/artsgoz-backend/internal/modules/yellow_card/domain"
)

type SemesterKey struct {
	SemesterStr string
	SortKey     int
}

// ParseSemester converting "1/2566" or "1/66" into a sortable integer
func ParseSemester(sem string) int {
	parts := strings.Split(sem, "/")
	if len(parts) != 2 {
		return 0
	}
	semNum, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
	year, _ := strconv.Atoi(strings.TrimSpace(parts[1]))

	// Normalize 2-digit years if present (e.g. 57 or 66 -> 2557 or 2566)
	if year < 100 {
		year += 2500
	}

	return year*10 + semNum
}

func GetGradePoint(grade string) (float64, bool) {
	switch strings.ToUpper(strings.TrimSpace(grade)) {
	case "A":
		return 4.0, true
	case "B+":
		return 3.5, true
	case "B":
		return 3.0, true
	case "C+":
		return 2.5, true
	case "C":
		return 2.0, true
	case "D+":
		return 1.5, true
	case "D":
		return 1.0, true
	case "F":
		return 0.0, true
	}
	return 0.0, false
}

// CalculateGPA computes GPATerm statistics for a set of subjects.
func CalculateGPA(userID string, subjects []domain.YellowCardSubject) []domain.GPATerm {
	if len(subjects) == 0 {
		return nil
	}

	// Group subjects by semester
	semMap := make(map[string][]domain.YellowCardSubject)
	for _, s := range subjects {
		sem := strings.TrimSpace(s.Semester)
		if sem == "" {
			continue
		}
		semMap[sem] = append(semMap[sem], s)
	}

	// Extract unique semesters and sort them
	var semesters []SemesterKey
	for semStr := range semMap {
		semesters = append(semesters, SemesterKey{
			SemesterStr: semStr,
			SortKey:     ParseSemester(semStr),
		})
	}

	// Simple bubble sort/insertion sort for semesters chronologically
	for i := 0; i < len(semesters); i++ {
		for j := i + 1; j < len(semesters); j++ {
			if semesters[i].SortKey > semesters[j].SortKey {
				semesters[i], semesters[j] = semesters[j], semesters[i]
			}
		}
	}

	var gpaTerms []domain.GPATerm

	// Cumulative variables
	var cumulativeGradePoints float64
	var cumulativeGPACredits float64
	var cumulativeCA float64
	var cumulativeCG float64

	for _, sem := range semesters {
		subs := semMap[sem.SemesterStr]
		
		var termGradePoints float64
		var termGPACredits float64
		var termCA float64
		var termCG float64

		for _, s := range subs {
			credits, err := strconv.ParseFloat(strings.TrimSpace(s.Credits), 64)
			if err != nil {
				continue
			}

			grade := strings.ToUpper(strings.TrimSpace(s.Grade))
			if grade == "" {
				continue
			}

			// Credits Attempted (CA)
			if grade == "A" || grade == "B+" || grade == "B" || grade == "C+" || grade == "C" || grade == "D+" || grade == "D" || grade == "F" || grade == "S" || grade == "U" {
				termCA += credits
			}

			// Credits Earned (CG)
			if grade == "A" || grade == "B+" || grade == "B" || grade == "C+" || grade == "C" || grade == "D+" || grade == "D" || grade == "S" {
				termCG += credits
			}

			// GPA points
			if gp, exists := GetGradePoint(grade); exists {
				termGradePoints += credits * gp
				termGPACredits += credits
			}
		}

		// Calculate Term GPA
		var termGPA float64
		if termGPACredits > 0 {
			termGPA = termGradePoints / termGPACredits
		}

		// Update cumulative stats
		cumulativeGradePoints += termGradePoints
		cumulativeGPACredits += termGPACredits
		cumulativeCA += termCA
		cumulativeCG += termCG

		// Calculate Cumulative GPAX
		var gpax float64
		if cumulativeGPACredits > 0 {
			gpax = cumulativeGradePoints / cumulativeGPACredits
		}

		gpaTerms = append(gpaTerms, domain.GPATerm{
			UserID:   userID,
			Semester: sem.SemesterStr,
			CA:       termCA,
			CG:       termCG,
			GPA:      termGPA,
			CAX:      cumulativeCA,
			CGX:      cumulativeCG,
			GPAX:     gpax,
		})
	}

	return gpaTerms
}
