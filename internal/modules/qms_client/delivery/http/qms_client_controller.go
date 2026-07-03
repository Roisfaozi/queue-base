package http

import (
	"net/http"

	"github.com/Roisfaozi/queue-base/internal/modules/qms_client/model"
	"github.com/Roisfaozi/queue-base/internal/modules/qms_client/usecase"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/Roisfaozi/queue-base/pkg/response"
	"github.com/Roisfaozi/queue-base/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type QMSClientController struct {
	useCase  usecase.QMSClientAdminUseCase
	validate *validator.Validate
	log      *logrus.Logger
}

func NewQMSClientController(useCase usecase.QMSClientAdminUseCase, validate *validator.Validate, log *logrus.Logger) *QMSClientController {
	return &QMSClientController{useCase: useCase, validate: validate, log: log}
}

func (h *QMSClientController) Create(c *gin.Context) {
	var req model.QMSClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, exception.ErrBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(c, err, validation.FormatValidationErrors(err))
		return
	}
	res, err := h.useCase.CreateClient(c.Request.Context(), &req)
	if err != nil {
		response.HandleError(c, err, "failed to create qms client")
		return
	}
	response.Created(c, res)
}

func (h *QMSClientController) CreateCredential(c *gin.Context) {
	var req model.QMSClientCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, exception.ErrBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(c, err, validation.FormatValidationErrors(err))
		return
	}
	res, err := h.useCase.CreateCredential(c.Request.Context(), &req)
	if err != nil {
		response.HandleError(c, err, "failed to create qms client credential")
		return
	}
	response.Created(c, res)
}

func (h *QMSClientController) Delete(c *gin.Context) {
	c.Status(http.StatusNotImplemented)
}
