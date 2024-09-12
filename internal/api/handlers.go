package api

import (
	"encoding/json"
	"github.com/abhikvarma/crontalk/internal/anthropic"
	"github.com/abhikvarma/crontalk/internal/cron_internal"
	"github.com/abhikvarma/crontalk/pkg/cronutil"
	"log"
	"net/http"
	"time"
)

type Handler struct {
	anthropicService *anthropic.Service
}

func NewHandler(as *anthropic.Service) *Handler {
	return &Handler{anthropicService: as}
}

type CronResponse struct {
	CronExpression string   `json:"cron_expression,omitempty"`
	NextRunTimes   []string `json:"next_run_times,omitempty"`
	ErrorMessage   string   `json:"error_message,omitempty"`
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

	llmCronResp, err := h.anthropicService.ProcessCronQuestion(r.Context(), input.CronQuestion)
	if err != nil {
		log.Printf("Error processing cron question: %v", err)
		http.Error(w, "Error processing cron questions", http.StatusInternalServerError)
		return
	}

	response := CronResponse{}
	if llmCronResp.Error != "" {
		response.ErrorMessage = llmCronResp.Error
		createJsonResponse(w, response, http.StatusOK)
		return
	}

	// todo: add a flow to fix using LLMs here, can loop in the users as well
	if err := cron_internal.ValidateCron(llmCronResp.Cron); err != nil {
		response.ErrorMessage = ":( Invalid cron expression generated: " + llmCronResp.Cron
		createJsonResponse(w, response, http.StatusOK)
		return
	}

	response.CronExpression = llmCronResp.Cron
	nextRunTimes, err := cronutil.GetNextRunTimes(llmCronResp.Cron, 5)
	if err != nil {
		log.Printf("Failed to calculate next run times for cron %s with error %v", llmCronResp, err)
	} else {
		response.NextRunTimes = make([]string, len(nextRunTimes))
		for i, t := range nextRunTimes {
			response.NextRunTimes[i] = t.Format(time.RFC3339)
		}
	}
	createJsonResponse(w, response, http.StatusOK)
}

func createJsonResponse(w http.ResponseWriter, response interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}
