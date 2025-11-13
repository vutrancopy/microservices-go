package document

import (
	"encoding/json"
	"time"

	
	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	"github.com/gbrayhan/microservices-go/src/domain/document"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// DocumentRepository handles driver document data access
type DocumentRepository struct {
	DB     *gorm.DB
	Logger *logger.Logger
}

type DocumentRepositoryInterface interface {
	Create(doc *document.DriverDocument) (*document.DriverDocument, error)
	GetByID(id int) (*document.DriverDocument, error)
	GetDriverDocuments(driverID int) (*[]document.DriverDocument, error)
	GetByType(driverID int, docType document.DocumentType) (*document.DriverDocument, error)
	Update(id int, updates map[string]interface{}) (*document.DriverDocument, error)
	Delete(id int) error
	GetPendingDocuments() (*[]document.DriverDocument, error)
	GetExpiredDocuments() (*[]document.DriverDocument, error)
	MarkAsExpired(id int) error
}

// NewDocumentRepository creates a new document repository
func NewDocumentRepository(db *gorm.DB, logger *logger.Logger) *DocumentRepository {
	return &DocumentRepository{
		DB:     db,
		Logger: logger,
	}
}

func (r *DocumentRepository) Create(doc *document.DriverDocument) (*document.DriverDocument, error) {
	result := r.DB.Create(doc)

	if result.Error != nil {
		var gormErr domainErrors.GormErr
		if err := json.Unmarshal([]byte(result.Error.Error()), &gormErr); err == nil {
			if gormErr.Number == 1062 {
				return nil, domainErrors.NewAppErrorWithType(domainErrors.ResourceAlreadyExists)
			}
		}
		r.Logger.Error("Error creating document", zap.Error(result.Error))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return doc, nil
}

func (r *DocumentRepository) GetByID(id int) (*document.DriverDocument, error) {
	var doc document.DriverDocument
	result := r.DB.First(&doc, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
		}
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return &doc, nil
}

func (r *DocumentRepository) GetDriverDocuments(driverID int) (*[]document.DriverDocument, error) {
	var documents []document.DriverDocument
	result := r.DB.Where("driver_id = ?", driverID).
		Order("created_at DESC").
		Find(&documents)

	if result.Error != nil {
		r.Logger.Error("Error getting driver documents", 
			zap.Int("driverID", driverID), 
			zap.Error(result.Error))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return &documents, nil
}

func (r *DocumentRepository) GetByType(driverID int, docType document.DocumentType) (*document.DriverDocument, error) {
	var doc document.DriverDocument
	result := r.DB.Where("driver_id = ? AND document_type = ?", driverID, docType).
		Order("created_at DESC").
		First(&doc)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
		}
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return &doc, nil
}

func (r *DocumentRepository) Update(id int, updates map[string]interface{}) (*document.DriverDocument, error) {
	var doc document.DriverDocument
	if err := r.DB.First(&doc, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
		}
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	result := r.DB.Model(&doc).Updates(updates)
	if result.Error != nil {
		r.Logger.Error("Error updating document", zap.Int("id", id), zap.Error(result.Error))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return &doc, nil
}

func (r *DocumentRepository) Delete(id int) error {
	result := r.DB.Delete(&document.DriverDocument{}, id)
	if result.Error != nil {
		return domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	if result.RowsAffected == 0 {
		return domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}
	return nil
}

func (r *DocumentRepository) GetPendingDocuments() (*[]document.DriverDocument, error) {
	var documents []document.DriverDocument
	result := r.DB.Where("status = ?", document.DocumentStatusPending).
		Order("created_at ASC").
		Find(&documents)

	if result.Error != nil {
		r.Logger.Error("Error getting pending documents", zap.Error(result.Error))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return &documents, nil
}

func (r *DocumentRepository) GetExpiredDocuments() (*[]document.DriverDocument, error) {
	var documents []document.DriverDocument
	now := time.Now()

	result := r.DB.Where("status = ? AND expiry_date IS NOT NULL AND expiry_date < ?", 
		document.DocumentStatusApproved, now).
		Find(&documents)

	if result.Error != nil {
		r.Logger.Error("Error getting expired documents", zap.Error(result.Error))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	return &documents, nil
}

func (r *DocumentRepository) MarkAsExpired(id int) error {
	result := r.DB.Model(&document.DriverDocument{}).
		Where("id = ?", id).
		Update("status", document.DocumentStatusExpired)

	if result.Error != nil {
		return domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	if result.RowsAffected == 0 {
		return domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}
	return nil
}

