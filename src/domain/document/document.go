package document

import (
	"errors"
	"time"
)

// DocumentType defines the type of driver document
type DocumentType string

const (
	DocumentTypeLicense         DocumentType = "license"          // Driver's license
	DocumentTypeInsurance       DocumentType = "insurance"        // Vehicle insurance
	DocumentTypeRegistration    DocumentType = "registration"     // Vehicle registration
	DocumentTypeBackground      DocumentType = "background_check" // Background check
	DocumentTypeProfilePhoto    DocumentType = "profile_photo"    // Driver photo
	DocumentTypeVehiclePhoto    DocumentType = "vehicle_photo"    // Vehicle photo
	DocumentTypeIdentification  DocumentType = "identification"   // National ID/Passport
)

// DocumentStatus represents verification status
type DocumentStatus string

const (
	DocumentStatusPending  DocumentStatus = "pending"  // Awaiting review
	DocumentStatusApproved DocumentStatus = "approved" // Verified and approved
	DocumentStatusRejected DocumentStatus = "rejected" // Rejected
	DocumentStatusExpired  DocumentStatus = "expired"  // Document has expired
)

// DriverDocument represents a driver's uploaded document
type DriverDocument struct {
	ID            int
	DriverID      int
	DocumentType  DocumentType
	DocumentNumber string         // License number, insurance policy number, etc.
	FileURL       string          // URL to uploaded file (S3, CloudStorage, etc.)
	Status        DocumentStatus
	ExpiryDate    *time.Time      // For documents that expire
	VerifiedAt    *time.Time
	VerifiedBy    *int            // Admin user ID who verified
	RejectionReason string
	Notes         string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// IsExpired checks if document has expired
func (d *DriverDocument) IsExpired() bool {
	if d.ExpiryDate == nil {
		return false
	}
	return time.Now().After(*d.ExpiryDate)
}

// CanDriverOperate checks if driver can operate based on document status
func (d *DriverDocument) CanDriverOperate() bool {
	if d.Status != DocumentStatusApproved {
		return false
	}

	if d.IsExpired() {
		return false
	}

	return true
}

// Approve approves the document
func (d *DriverDocument) Approve(verifiedBy int) {
	d.Status = DocumentStatusApproved
	now := time.Now()
	d.VerifiedAt = &now
	d.VerifiedBy = &verifiedBy
	d.RejectionReason = ""
}

// Reject rejects the document with reason
func (d *DriverDocument) Reject(reason string, verifiedBy int) error {
	if reason == "" {
		return errors.New("rejection reason is required")
	}

	d.Status = DocumentStatusRejected
	now := time.Now()
	d.VerifiedAt = &now
	d.VerifiedBy = &verifiedBy
	d.RejectionReason = reason

	return nil
}

// MarkAsExpired marks document as expired
func (d *DriverDocument) MarkAsExpired() {
	d.Status = DocumentStatusExpired
}

// Validate validates the document
func (d *DriverDocument) Validate() error {
	if d.DriverID == 0 {
		return errors.New("driver ID is required")
	}

	if d.DocumentType == "" {
		return errors.New("document type is required")
	}

	if d.FileURL == "" {
		return errors.New("file URL is required")
	}

	// Validate document type
	switch d.DocumentType {
	case DocumentTypeLicense, DocumentTypeInsurance, DocumentTypeRegistration,
		DocumentTypeBackground, DocumentTypeProfilePhoto, DocumentTypeVehiclePhoto,
		DocumentTypeIdentification:
		// valid
	default:
		return errors.New("invalid document type")
	}

	// Validate status
	switch d.Status {
	case DocumentStatusPending, DocumentStatusApproved, DocumentStatusRejected, DocumentStatusExpired:
		// valid
	default:
		return errors.New("invalid document status")
	}

	return nil
}

// DriverDocumentSummary provides summary of driver's document verification status
type DriverDocumentSummary struct {
	DriverID           int
	AllDocumentsValid  bool
	LicenseVerified    bool
	InsuranceVerified  bool
	RegistrationVerified bool
	BackgroundVerified bool
	ProfilePhotoUploaded bool
	VehiclePhotoUploaded bool
	PendingCount       int
	RejectedCount      int
	ExpiredCount       int
}

// IDocumentService defines the interface for document operations
type IDocumentService interface {
	UploadDocument(document *DriverDocument) (*DriverDocument, error)
	GetDriverDocuments(driverID int) (*[]DriverDocument, error)
	GetDocumentByID(id int) (*DriverDocument, error)
	ApproveDocument(id int, verifiedBy int) error
	RejectDocument(id int, reason string, verifiedBy int) error
	GetDocumentByType(driverID int, docType DocumentType) (*DriverDocument, error)
	GetDriverDocumentSummary(driverID int) (*DriverDocumentSummary, error)
	GetPendingDocuments() (*[]DriverDocument, error)
	CheckExpiredDocuments() error // Background job to mark expired documents
	DeleteDocument(id int) error
}

// RequiredDocuments returns list of required documents for driver verification
var RequiredDocuments = []DocumentType{
	DocumentTypeLicense,
	DocumentTypeInsurance,
	DocumentTypeRegistration,
	DocumentTypeBackground,
	DocumentTypeProfilePhoto,
	DocumentTypeVehiclePhoto,
}
