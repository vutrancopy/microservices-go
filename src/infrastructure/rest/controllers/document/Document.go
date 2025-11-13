package document

import (
	"net/http"
	"strconv"

	"github.com/gbrayhan/microservices-go/src/application/usecases/document"
	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	domainDocument "github.com/gbrayhan/microservices-go/src/domain/document"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type IDocumentController interface {
	UploadDocument(ctx *gin.Context)
	GetDriverDocuments(ctx *gin.Context)
	GetDocumentByID(ctx *gin.Context)
	ApproveDocument(ctx *gin.Context)
	RejectDocument(ctx *gin.Context)
	GetDocumentSummary(ctx *gin.Context)
	GetPendingDocuments(ctx *gin.Context)
	DeleteDocument(ctx *gin.Context)
}

type DocumentController struct {
	documentUseCase document.IDocumentUseCase
	logger          *logger.Logger
}

func NewDocumentController(documentUseCase document.IDocumentUseCase, logger *logger.Logger) *DocumentController {
	return &DocumentController{
		documentUseCase: documentUseCase,
		logger:          logger,
	}
}

type UploadDocumentRequest struct {
	DriverID       int    `json:"driverId" binding:"required"`
	DocumentType   string `json:"documentType" binding:"required"`
	DocumentNumber string `json:"documentNumber"`
	FileURL        string `json:"fileUrl" binding:"required"`
	ExpiryDate     string `json:"expiryDate"`
	Notes          string `json:"notes"`
}

type ApproveDocumentRequest struct {
	VerifiedBy int `json:"verifiedBy" binding:"required"`
}

type RejectDocumentRequest struct {
	Reason     string `json:"reason" binding:"required"`
	VerifiedBy int    `json:"verifiedBy" binding:"required"`
}

func (c *DocumentController) UploadDocument(ctx *gin.Context) {
	var req UploadDocumentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.logger.Info("Uploading driver document", 
		zap.Int("driverID", req.DriverID),
		zap.String("type", req.DocumentType))

	doc := &domainDocument.DriverDocument{
		DriverID:       req.DriverID,
		DocumentType:   domainDocument.DocumentType(req.DocumentType),
		DocumentNumber: req.DocumentNumber,
		FileURL:        req.FileURL,
		Notes:          req.Notes,
	}

	uploaded, err := c.documentUseCase.UploadDocument(doc)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Document uploaded successfully",
		"data":    uploaded,
	})
}

func (c *DocumentController) GetDriverDocuments(ctx *gin.Context) {
	driverIDParam := ctx.Param("driverId")
	driverID, err := strconv.Atoi(driverIDParam)
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.logger.Info("Getting driver documents", zap.Int("driverID", driverID))

	documents, err := c.documentUseCase.GetDriverDocuments(driverID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    documents,
	})
}

func (c *DocumentController) GetDocumentByID(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.logger.Info("Getting document by ID", zap.Int("id", id))

	doc, err := c.documentUseCase.GetDocumentByID(id)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    doc,
	})
}

func (c *DocumentController) ApproveDocument(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	var req ApproveDocumentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.logger.Info("Approving document", zap.Int("id", id), zap.Int("verifiedBy", req.VerifiedBy))

	err = c.documentUseCase.ApproveDocument(id, req.VerifiedBy)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Document approved successfully",
	})
}

func (c *DocumentController) RejectDocument(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	var req RejectDocumentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.logger.Info("Rejecting document", 
		zap.Int("id", id),
		zap.String("reason", req.Reason))

	err = c.documentUseCase.RejectDocument(id, req.Reason, req.VerifiedBy)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Document rejected",
	})
}

func (c *DocumentController) GetDocumentSummary(ctx *gin.Context) {
	driverIDParam := ctx.Param("driverId")
	driverID, err := strconv.Atoi(driverIDParam)
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.logger.Info("Getting document summary", zap.Int("driverID", driverID))

	summary, err := c.documentUseCase.GetDocumentSummary(driverID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    summary,
	})
}

func (c *DocumentController) GetPendingDocuments(ctx *gin.Context) {
	c.logger.Info("Getting pending documents")

	documents, err := c.documentUseCase.GetPendingDocuments()
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    documents,
	})
}

func (c *DocumentController) DeleteDocument(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		_ = ctx.Error(domainErrors.NewAppErrorWithType(domainErrors.ValidationError))
		return
	}

	c.logger.Info("Deleting document", zap.Int("id", id))

	err = c.documentUseCase.DeleteDocument(id)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Document deleted successfully",
	})
}
