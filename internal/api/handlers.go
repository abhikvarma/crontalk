package api

import (
	"encoding/json"
	"github.com/abhikvarma/crontalk/internal/anthropic"
	"github.com/abhikvarma/crontalk/internal/cron_internal"
	"github.com/abhikvarma/crontalk/pkg/cronutil"
	"net/http"
	"time"
)

type Handler struct {
	anthropicService *anthropic.Service
}

func NewHandler(as *anthropic.Service) *Handler {
	return &Handler{anthropicService: as}
}

func (h *Handler) HandleCronRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		CronQuestion string `json:"cron_question"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	cronExpression, err := h.anthropicService.ProcessNaturalLanguage(r.Context(), input.CronQuestion)
	if err != nil {
		http.Error(w, "Failed to process natural language", http.StatusInternalServerError)
		return
	}

	if err := cron_internal.ValidateCron(cronExpression); err != nil {
		http.Error(w, "Invalid cron_internal expression generated", http.StatusInternalServerError)
		return
	}

	nextRunTimes, err := cronutil.GetNextRunTimes(cronExpression, 5)
	if err != nil {
		http.Error(w, "Failed to calculate next run times", http.StatusInternalServerError)
		return
	}

	response := struct {
		CronExpression string   `json:"cron_expression"`
		NextRunTimes   []string `json:"next_run_times"`
	}{
		CronExpression: cronExpression,
		NextRunTimes:   make([]string, len(nextRunTimes)),
	}

	for i, t := range nextRunTimes {
		response.NextRunTimes[i] = t.Format(time.RFC3339)
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
		return
	}
}
