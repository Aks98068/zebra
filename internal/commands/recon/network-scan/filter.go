package networkscan

import "strings"

type FilterAnalysis struct {
	Open         int
	Closed       int
	Filtered     int
	OpenFiltered int
	Unknown      int
	Observations []string
}

func AnalyzeFiltering(results []PortResult) FilterAnalysis {
	analysis := FilterAnalysis{
		Observations: make([]string, 0),
	}

	for _, result := range results {
		switch strings.ToLower(result.State) {
		case "open":
			analysis.Open++

		case "closed":
			analysis.Closed++

		case "filtered":
			analysis.Filtered++

		case "open|filtered":
			analysis.OpenFiltered++

		default:
			analysis.Unknown++
		}
	}

	if analysis.Filtered > 0 {
		analysis.Observations = append(
			analysis.Observations,
			"Some probes received no definitive response",
		)
	}

	if analysis.OpenFiltered > 0 {
		analysis.Observations = append(
			analysis.Observations,
			"Some UDP ports could not be distinguished between open and filtered",
		)
	}

	if analysis.Closed > 0 {
		analysis.Observations = append(
			analysis.Observations,
			"Some ports responded as closed",
		)
	}

	if analysis.Open > 0 {
		analysis.Observations = append(
			analysis.Observations,
			"Some ports accepted connections or returned service responses",
		)
	}

	return analysis
}
