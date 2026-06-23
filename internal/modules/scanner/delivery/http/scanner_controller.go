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
	clientID := c.GetHeader("X-Client-ID")
	apiKey := c.GetHeader("X-API-Key")
	if clientID == "" || apiKey == "" {
		response.BadRequest(c, exception.ErrBadRequest, "missing scanner headers")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(c, err, validation.FormatValidationErrors(err))
		return
	}
	res, err := h.useCase.CheckIn(c.Request.Context(), &usecase.CheckInRequest{
		Action:               req.Action,
		ClientID:             clientID,
		APIKey:               apiKey,
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

// CheckIn godoc
// @Summary      Scanner check-in
// @Description  Scanner entrypoint for queue registration or forward transition using client headers.
// @Tags         scanner
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        X-Client-ID header string true "Scanner client ID"
// @Param        X-API-Key header string true "Scanner API key"
// @Param        body body model.CheckInRequest true "Scanner payload"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper
// @Router       /scanner/check-in [post]
