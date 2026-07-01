package http

import (
	"github.com/Roisfaozi/queue-base/internal/modules/scanner/model"
	"github.com/Roisfaozi/queue-base/internal/modules/scanner/usecase"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/Roisfaozi/queue-base/pkg/response"
	"github.com/Roisfaozi/queue-base/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type ScannerController struct {
	useCase  usecase.ScannerUseCase
	validate *validator.Validate
	log      *logrus.Logger
}

func NewScannerController(useCase usecase.ScannerUseCase, validate *validator.Validate) *ScannerController {
	return NewScannerControllerWithLogger(useCase, validate, logrus.New())
}

func NewScannerControllerWithLogger(useCase usecase.ScannerUseCase, validate *validator.Validate, log *logrus.Logger) *ScannerController {
	if log == nil {
		log = logrus.New()
	}
	return &ScannerController{useCase: useCase, validate: validate, log: log}
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
	ctx := database.SetBranchContext(c.Request.Context(), req.BranchID)
	res, err := h.useCase.CheckIn(ctx, &usecase.CheckInRequest{
		Action:               req.Action,
		BranchID:             req.BranchID,
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
		h.log.Errorf("Scanner CheckIn failed: %v", err)
		response.HandleError(c, err, "failed to process scanner check-in")
		return
	}
	response.Success(c, res)
}
