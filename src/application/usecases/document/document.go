package document

import (
	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	"github.com/gbrayhan/microservices-go/src/domain/document"
	documentRepo "github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/document"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"go.uber.org/zap"
)

type IDocumentUseCase interface {
	UploadDocument(doc *document.DriverDocument) (*document.DriverDocument, error)
	GetDriverDocuments(driverID int) (*[]document.DriverDocument, error)
	GetDocumentByID(id int) (*document.DriverDocument, error)
	ApproveDocument(id int, verifiedBy int) error
	RejectDocument(id int, reason string, verifiedBy int) error
	GetDocumentSummary(driverID int) (*document.DriverDocumentSummary, error)
	GetPendingDocuments() (*[]document.DriverDocument, error)
	CheckExpiredDocuments() error
	DeleteDocument(id int) error
}

type DocumentUseCase struct {
	documentRepo *documentRepo.DocumentRepository
	logger       *logger.Logger
}

func NewDocumentUseCase(
	documentRepo *documentRepo.DocumentRepository,
	logger *logger.Logger,
) *DocumentUseCase {
	return &DocumentUseCase{
		documentRepo: documentRepo,
		logger:       logger,
	}
}

func (u *DocumentUseCase) UploadDocument(doc *document.DriverDocument) (*document.DriverDocument, error) {
	u.logger.Info("Uploading driver document", 
		zap.Int("driverID", doc.DriverID),
		zap.String("type", string(doc.DocumentType)))

	// Validate document
	if err := doc.Validate(); err != nil {
		u.logger.Error("Invalid document", zap.Error(err))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.ValidationError)
	}

	// Set initial status to pending
	doc.Status = document.DocumentStatusPending

	createdDoc, err := u.documentRepo.Create(doc)
	if err != nil {
		u.logger.Error("Error uploading document", zap.Error(err))
		return nil, err
	}

	u.logger.Info("Document uploaded successfully", zap.Int("id", createdDoc.ID))
	return createdDoc, nil
}

func (u *DocumentUseCase) GetDriverDocuments(driverID int) (*[]document.DriverDocument, error) {
	u.logger.Info("Getting driver documents", zap.Int("driverID", driverID))

	documents, err := u.documentRepo.GetDriverDocuments(driverID)
	if err != nil {
		u.logger.Error("Error getting driver documents", zap.Error(err))
		return nil, err
	}

	return documents, nil
}

func (u *DocumentUseCase) GetDocumentByID(id int) (*document.DriverDocument, error) {
	u.logger.Info("Getting document by ID", zap.Int("id", id))

	doc, err := u.documentRepo.GetByID(id)
	if err != nil {
		u.logger.Error("Error getting document", zap.Error(err))
		return nil, err
	}

	return doc, nil
}

func (u *DocumentUseCase) ApproveDocument(id int, verifiedBy int) error {
	u.logger.Info("Approving document", zap.Int("id", id), zap.Int("verifiedBy", verifiedBy))

	// Get document
	doc, err := u.documentRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Approve document
	doc.Approve(verifiedBy)

	// Update in database
	updates := map[string]interface{}{
		"status":       doc.Status,
		"verified_at":  doc.VerifiedAt,
		"verified_by":  doc.VerifiedBy,
		"rejection_reason": "",
	}

	_, err = u.documentRepo.Update(id, updates)
	if err != nil {
		u.logger.Error("Error approving document", zap.Error(err))
		return err
	}

	u.logger.Info("Document approved successfully", zap.Int("id", id))
	return nil
}

func (u *DocumentUseCase) RejectDocument(id int, reason string, verifiedBy int) error {
	u.logger.Info("Rejecting document", 
		zap.Int("id", id), 
		zap.String("reason", reason),
		zap.Int("verifiedBy", verifiedBy))

	// Get document
	doc, err := u.documentRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Reject document
	if err := doc.Reject(reason, verifiedBy); err != nil {
		return domainErrors.NewAppErrorWithType(domainErrors.ValidationError)
	}

	// Update in database
	updates := map[string]interface{}{
		"status":            doc.Status,
		"verified_at":       doc.VerifiedAt,
		"verified_by":       doc.VerifiedBy,
		"rejection_reason":  doc.RejectionReason,
	}

	_, err = u.documentRepo.Update(id, updates)
	if err != nil {
		u.logger.Error("Error rejecting document", zap.Error(err))
		return err
	}

	u.logger.Info("Document rejected successfully", zap.Int("id", id))
	return nil
}

func (u *DocumentUseCase) GetDocumentSummary(driverID int) (*document.DriverDocumentSummary, error) {
	u.logger.Info("Getting document summary for driver", zap.Int("driverID", driverID))

	documents, err := u.documentRepo.GetDriverDocuments(driverID)
	if err != nil {
		return nil, err
	}

	summary := &document.DriverDocumentSummary{
		DriverID: driverID,
	}

	// Count documents by status and type
	for _, doc := range *documents {
		switch doc.Status {
		case document.DocumentStatusPending:
			summary.PendingCount++
		case document.DocumentStatusRejected:
			summary.RejectedCount++
		case document.DocumentStatusExpired:
			summary.ExpiredCount++
		case document.DocumentStatusApproved:
			// Check document type
			switch doc.DocumentType {
			case document.DocumentTypeLicense:
				summary.LicenseVerified = !doc.IsExpired()
			case document.DocumentTypeInsurance:
				summary.InsuranceVerified = !doc.IsExpired()
			case document.DocumentTypeRegistration:
				summary.RegistrationVerified = !doc.IsExpired()
			case document.DocumentTypeBackground:
				summary.BackgroundVerified = !doc.IsExpired()
			case document.DocumentTypeProfilePhoto:
				summary.ProfilePhotoUploaded = true
			case document.DocumentTypeVehiclePhoto:
				summary.VehiclePhotoUploaded = true
			}
		}
	}

	// All documents valid if all required documents are verified
	summary.AllDocumentsValid = summary.LicenseVerified && 
		summary.InsuranceVerified && 
		summary.RegistrationVerified && 
		summary.BackgroundVerified && 
		summary.ProfilePhotoUploaded && 
		summary.VehiclePhotoUploaded

	return summary, nil
}

func (u *DocumentUseCase) GetPendingDocuments() (*[]document.DriverDocument, error) {
	u.logger.Info("Getting pending documents")

	documents, err := u.documentRepo.GetPendingDocuments()
	if err != nil {
		u.logger.Error("Error getting pending documents", zap.Error(err))
		return nil, err
	}

	return documents, nil
}

func (u *DocumentUseCase) CheckExpiredDocuments() error {
	u.logger.Info("Checking for expired documents")

	// Get documents that have expired
	expiredDocs, err := u.documentRepo.GetExpiredDocuments()
	if err != nil {
		return err
	}

	// Mark each as expired
	for _, doc := range *expiredDocs {
		if err := u.documentRepo.MarkAsExpired(doc.ID); err != nil {
			u.logger.Error("Error marking document as expired", 
				zap.Int("id", doc.ID), 
				zap.Error(err))
		}
	}

	u.logger.Info("Expired documents check completed", zap.Int("count", len(*expiredDocs)))
	return nil
}

func (u *DocumentUseCase) DeleteDocument(id int) error {
	u.logger.Info("Deleting document", zap.Int("id", id))

	err := u.documentRepo.Delete(id)
	if err != nil {
		u.logger.Error("Error deleting document", zap.Error(err))
		return err
	}

	u.logger.Info("Document deleted successfully", zap.Int("id", id))
	return nil
}
