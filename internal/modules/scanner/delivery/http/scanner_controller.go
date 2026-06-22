package http

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/scanner/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/scanner/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ScannerController struct {
	useCase  usecase.ScannerUseCase
	validate *validator.Validate
}

func NewScannerController(useCase usecase.ScannerUseCase, validate *validator.Validate) *ScannerController {
	return &ScannerController{useCase: useCase, validate: validate}
}

func (h *ScannerController) CheckIn(c *gin.Context) {
	var req model.CheckInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, exception.ErrBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(c, err, validation.FormatValidationErrors(err))
		return
	}
	res, err := h.useCase.CheckIn(c.Request.Context(), &usecase.CheckInRequest{
		Action:               req.Action,
		ClientID:             req.ClientID,
		APIKey:               req.APIKey,
		ServiceID:            req.ServiceID,
		PatientID:            req.PatientID,
		PatientName:          req.PatientName,
		QueueID:              req.QueueID,
		DestinationServiceID: req.DestinationServiceID,
		DestinationCounterID: req.DestinationCounterID,
	})
	if err != nil {
		response.HandleError(c, err, "failed to process scanner check-in")
		return
	}
	response.Success(c, res)
}
