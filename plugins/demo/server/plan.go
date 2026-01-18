package server

import (
	"tgp/core/data"
	"tgp/core/http"
	"tgp/core/i18n"
)

// handlePlan обрабатывает демонстрацию визуализации плана.
func (s *Server) handlePlan(w http.ResponseWriter, r *http.Request) {

	var result PlanResult

	if plan, err := data.Get[executionPlan](s.request, "_execute_plan_"); err == nil {
		result.Plan = plan
		result.SVG = generatePlanSVG(plan)
		result.Message = i18n.Msg("Plan loaded and visualized")
		s.writeHTML(w, http.StatusOK, formatResult(result))
		return
	}

	result.Message = i18n.Msg("ExecutionPlan not found in request. This is normal if the plugin is not run through the scheduler.")
	result.SVG = generatePlanSVG(executionPlan{})
	s.writeHTML(w, http.StatusOK, formatResult(result))
}
