package report

import (
	"encoding/json"
	"time"

	"github.com/mahmoudnasser1561/zero-credential-pipeline/stackcheck/internal/check"
)

type JSONReporter struct {
	Now func() time.Time
}

func NewJSON() *JSONReporter {
	return &JSONReporter{Now: time.Now}
}

type resultDTO struct {
	Name      string `json:"name"`
	Target    string `json:"target"`
	Status    string `json:"status"`
	LatencyMs int64  `json:"latency_ms"`
	Detail    string `json:"detail"`
}

type reportDTO struct {
	OverallStatus string      `json:"overall_status"`
	CheckedAt     string      `json:"checked_at"`
	Results       []resultDTO `json:"results"`
}

func (j *JSONReporter) Render(results []check.Result) string {
	dto := reportDTO{
		OverallStatus: string(Overall(results)),
		CheckedAt:     j.Now().UTC().Format(time.RFC3339),
	}

	for _, r := range results {
		dto.Results = append(dto.Results, resultDTO{
			Name:      r.Name,
			Target:    r.Target,
			Status:    string(r.Status),
			LatencyMs: r.Latency.Milliseconds(),
			Detail:    r.Detail,
		})
	}

	out, err := json.MarshalIndent(dto, "", "  ")
	if err != nil {
		// marshaling a fixed, simple struct of strings/ints cannot
		// realistically fail; surface it plainly rather than swallow it
		return `{"overall_status":"fail","checked_at":"","results":[],"marshal_error":"` + err.Error() + `"}`
	}

	return string(out)
}
