package http

import (
	"fmt"
	"net/http"
	"strconv"

	constants "github.com/pandusatrianura/code-with-umam-second-meeting/constant"
	"github.com/pandusatrianura/code-with-umam-second-meeting/internal/health/service"
	"github.com/pandusatrianura/code-with-umam-second-meeting/pkg/response"
)

type HealthHandler struct {
	service service.HealthService
}

func NewHealthHandler(service service.HealthService) *HealthHandler {
	return &HealthHandler{service: service}
}

// HealthCheckAPI godoc
// @Summary Get health status of API
// @Description Get health status of API
// @Tags healthcheck
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 503 {object} map[string]string
// @Router /api/health/service [get]
func (h *HealthHandler) API(w http.ResponseWriter, r *http.Request) {
	var result response.APIResponse
	svcHealthCheckResult := h.service.API()
	if svcHealthCheckResult.IsHealthy {
		result.Code = strconv.Itoa(constants.SuccessCode)
		result.Message = fmt.Sprintf("%s is healthy", svcHealthCheckResult.Name)
		response.WriteJSONResponse(w, http.StatusOK, result)
		return
	}

	result.Code = strconv.Itoa(constants.ErrorCode)
	result.Message = fmt.Sprintf("%s is not healthy", svcHealthCheckResult.Name)
	response.WriteJSONResponse(w, http.StatusServiceUnavailable, result)
	return
}

// HealthCheckDatabase godoc
// @Summary Get health status of Database
// @Description Get health status of Database
// @Tags healthcheck
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 503 {object} map[string]string
// @Router /api/health/db [get]
func (h *HealthHandler) DB(w http.ResponseWriter, r *http.Request) {
	var result response.APIResponse
	svcHealthCheckResult, err := h.service.DB()
	if svcHealthCheckResult.IsHealthy && err == nil {
		result.Code = strconv.Itoa(constants.SuccessCode)
		result.Message = fmt.Sprintf("%s is healthy", svcHealthCheckResult.Name)
		response.WriteJSONResponse(w, http.StatusOK, result)
		return
	}

	result.Code = strconv.Itoa(constants.ErrorCode)
	result.Message = fmt.Sprintf("%s is not healthy because %s", svcHealthCheckResult.Name, err.Error())
	response.WriteJSONResponse(w, http.StatusServiceUnavailable, result)
	return
}
