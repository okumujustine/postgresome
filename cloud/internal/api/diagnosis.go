package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/okumujustine/postgresome/cloud/internal/diagnosis"
)

type runDiagnosisRequest struct {
	DatabaseInstanceID string `json:"database_instance_id"`
}

type runDiagnosisResponse struct {
	Status     string                `json:"status"`
	DetectedAt string                `json:"detected_at"`
	Findings   []dashboardFindingDTO `json:"findings"`
}

func (s *Server) handleRunDiagnosis(w http.ResponseWriter, r *http.Request) {
	var req runDiagnosisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.DatabaseInstanceID == "" {
		http.Error(w, "database_instance_id is required", http.StatusBadRequest)
		return
	}

	service := diagnosis.NewService(s.pool)
	result, err := service.RunForDatabase(r.Context(), req.DatabaseInstanceID)
	if err != nil {
		log.Printf("failed to run diagnosis: %v", err)
		http.Error(w, "failed to run diagnosis", http.StatusInternalServerError)
		return
	}

	findings := make([]dashboardFindingDTO, len(result.Findings))
	for i, finding := range result.Findings {
		findings[i] = dashboardFindingDTO{
			Severity:            finding.Severity,
			Category:            finding.Category,
			Title:               finding.Title,
			Message:             finding.Message,
			Recommendation:      finding.Recommendation,
			Status:              "open",
			ProblemSummary:      finding.ProblemSummary,
			EvidenceSummary:     finding.EvidenceSummary,
			ImpactSummary:       finding.ImpactSummary,
			SuggestedAction:     finding.SuggestedAction,
			ConfidenceLabel:     finding.ConfidenceLabel,
			ConfidenceScore:     finding.ConfidenceScore,
			BaselineValue:       finding.BaselineValue,
			BaselineLabel:       finding.BaselineLabel,
			ChangeSummary:       finding.ChangeSummary,
			VerificationHint:    finding.VerificationHint,
			VerificationStatus:  finding.VerificationStatus,
			VerificationSummary: finding.VerificationSummary,
			RuleKey:             finding.RuleKey,
			ResourceType:        finding.ResourceType,
			ResourceName:        finding.ResourceName,
			CurrentValue:        finding.CurrentValue,
			ThresholdValue:      finding.ThresholdValue,
			DetectedAt:          finding.DetectedAt,
			FirstSeenAt:         finding.DetectedAt,
			LastSeenAt:          finding.DetectedAt,
		}
	}

	writeJSON(w, http.StatusOK, runDiagnosisResponse{
		Status:     "accepted",
		DetectedAt: result.DetectedAt.Format(time.RFC3339Nano),
		Findings:   findings,
	})
}
