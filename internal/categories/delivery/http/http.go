package http

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	constants "github.com/pandusatrianura/code-with-umam-second-meeting/constant"
	"github.com/pandusatrianura/code-with-umam-second-meeting/internal/categories/entity"
	"github.com/pandusatrianura/code-with-umam-second-meeting/internal/categories/service"
	"github.com/pandusatrianura/code-with-umam-second-meeting/pkg/response"
)

type CategoryHandler struct {
	service service.CategoryService
}

func NewCategoryHandler(service service.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

// API godoc
// @Summary Get health status of categories API
// @Description Get health status of categories API
// @Tags categories
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/categories/health [get]
func (h *CategoryHandler) API(w http.ResponseWriter, r *http.Request) {
	var result response.APIResponse
	svcHealthCheckResult := h.service.API()
	if svcHealthCheckResult.IsHealthy {
		result.Code = strconv.Itoa(constants.SuccessCode)
		result.Message = fmt.Sprintf("%s is healthy", svcHealthCheckResult.Name)
		response.WriteJSONResponse(w, http.StatusOK, result)
		return
	} else {
		result.Code = strconv.Itoa(constants.ErrorCode)
		result.Message = fmt.Sprintf("%s is not healthy", svcHealthCheckResult.Name)
		response.WriteJSONResponse(w, http.StatusServiceUnavailable, result)
		return
	}
}

// CreateCategory godoc
// @Summary Create a new category
// @Description Create a new category
// @Tags categories
// @Accept json
// @Produce json
// @Param category body entity.RequestCategory true "Category Data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /api/categories [post]
func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var requestCategory entity.RequestCategory
	if err := response.ParseJSON(r, &requestCategory); err != nil {
		response.Error(w, http.StatusBadRequest, constants.ErrorCode, constants.ErrInvalidCategoryRequest, err)
		return
	}

	if err := h.service.CreateCategory(&requestCategory); err != nil {
		response.Error(w, http.StatusInternalServerError, constants.ErrorCode, "Category created failed", err)
		return
	}

	response.Success(w, http.StatusCreated, constants.SuccessCode, "Category created successfully", nil)
}

// UpdateCategory godoc
// @Summary Update a category
// @Description Update a category
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Param category body entity.RequestCategory true "Category Data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /api/categories/{id} [put]
func (h *CategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	var requestCategory entity.RequestCategory

	idStr := strings.TrimPrefix(r.URL.Path, "/categories/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, constants.ErrorCode, constants.ErrInvalidCategoryID, err)
		return
	}

	if err := response.ParseJSON(r, &requestCategory); err != nil {
		response.Error(w, http.StatusBadRequest, constants.ErrorCode, constants.ErrInvalidCategoryRequest, err)
		return
	}

	if err := h.service.UpdateCategory(int64(id), &requestCategory); err != nil {
		response.Error(w, http.StatusInternalServerError, constants.ErrorCode, "Category updated failed", err)
		return
	}

	response.Success(w, http.StatusOK, constants.SuccessCode, "Category updated successfully", nil)
}

// DeleteCategory godoc
// @Summary Delete a category
// @Description Delete a category
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /api/categories/{id} [delete]
func (h *CategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/categories/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, constants.ErrorCode, constants.ErrInvalidCategoryID, err)
		return
	}

	if err := h.service.DeleteCategory(int64(id)); err != nil {
		response.Error(w, http.StatusInternalServerError, constants.ErrorCode, "Category delete failed", err)
		return
	}

	response.Success(w, http.StatusOK, constants.SuccessCode, "Category deleted successfully", nil)
}

// GetCategoryByID godoc
// @Summary Get a category by ID
// @Description Get a category by ID
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /api/categories/{id} [get]
func (h *CategoryHandler) GetCategoryByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/categories/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, constants.ErrorCode, constants.ErrInvalidCategoryID, err)
		return
	}

	category, err := h.service.GetCategoryByID(int64(id))
	if err != nil {
		response.Error(w, http.StatusInternalServerError, constants.ErrorCode, "Category retrieved failed", err)
		return
	}

	response.Success(w, http.StatusOK, constants.SuccessCode, "Category retrieved successfully", category)
}

// GetAllCategories godoc
// @Summary Get all categories
// @Description Get all categories
// @Tags categories
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /api/categories [get]
func (h *CategoryHandler) GetAllCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.service.GetAllCategories()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, constants.ErrorCode, "Categories retrieved failed", err)
		return
	}

	response.Success(w, http.StatusOK, constants.SuccessCode, "Categories retrieved successfully", categories)
}
