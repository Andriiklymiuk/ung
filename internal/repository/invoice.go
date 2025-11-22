package repository

import (
	"time"

	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/Andriiklymiuk/ung/internal/models"
	"gorm.io/gorm"
)

type InvoiceRepository struct {
	db *gorm.DB
}

func NewInvoiceRepository() *InvoiceRepository {
	return &InvoiceRepository{db: db.GormDB}
}

func (r *InvoiceRepository) Create(invoice *models.Invoice) error {
	return r.db.Create(invoice).Error
}

func (r *InvoiceRepository) GetByID(id uint) (*models.Invoice, error) {
	var invoice models.Invoice
	err := r.db.First(&invoice, id).Error
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (r *InvoiceRepository) List() ([]models.Invoice, error) {
	var invoices []models.Invoice
	err := r.db.Order("id DESC").Find(&invoices).Error
	return invoices, err
}

func (r *InvoiceRepository) GetByMonth(year int, month int) ([]models.Invoice, error) {
	var invoices []models.Invoice
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	err := r.db.Where("issued_date >= ? AND issued_date < ?", startDate, endDate).
		Order("issued_date DESC").
		Find(&invoices).Error

	return invoices, err
}

func (r *InvoiceRepository) GetByStatus(status models.InvoiceStatus) ([]models.Invoice, error) {
	var invoices []models.Invoice
	err := r.db.Where("status = ?", status).Order("issued_date DESC").Find(&invoices).Error
	return invoices, err
}

func (r *InvoiceRepository) Update(invoice *models.Invoice) error {
	return r.db.Save(invoice).Error
}

func (r *InvoiceRepository) Delete(id uint) error {
	return r.db.Delete(&models.Invoice{}, id).Error
}

func (r *InvoiceRepository) CountByInvoiceNumPattern(pattern string) (int64, error) {
	var count int64
	err := r.db.Model(&models.Invoice{}).Where("invoice_num LIKE ?", pattern).Count(&count).Error
	return count, err
}

// InvoiceRecipientRepository handles invoice-client relationships
type InvoiceRecipientRepository struct {
	db *gorm.DB
}

func NewInvoiceRecipientRepository() *InvoiceRecipientRepository {
	return &InvoiceRecipientRepository{db: db.GormDB}
}

func (r *InvoiceRecipientRepository) Create(invoiceID, clientID uint) error {
	result := r.db.Exec("INSERT INTO invoice_recipients (invoice_id, client_id) VALUES (?, ?)", invoiceID, clientID)
	return result.Error
}

func (r *InvoiceRecipientRepository) GetClientByInvoiceID(invoiceID uint) (*models.Client, error) {
	var client models.Client
	err := r.db.Table("clients").
		Joins("JOIN invoice_recipients ON clients.id = invoice_recipients.client_id").
		Where("invoice_recipients.invoice_id = ?", invoiceID).
		First(&client).Error

	if err != nil {
		return nil, err
	}
	return &client, nil
}

// InvoiceLineItemRepository handles invoice line items
type InvoiceLineItemRepository struct {
	db *gorm.DB
}

func NewInvoiceLineItemRepository() *InvoiceLineItemRepository {
	return &InvoiceLineItemRepository{db: db.GormDB}
}

func (r *InvoiceLineItemRepository) Create(item *models.InvoiceLineItem) error {
	return r.db.Create(item).Error
}

func (r *InvoiceLineItemRepository) GetByInvoiceID(invoiceID uint) ([]models.InvoiceLineItem, error) {
	var items []models.InvoiceLineItem
	err := r.db.Where("invoice_id = ?", invoiceID).Order("id").Find(&items).Error
	return items, err
}

func (r *InvoiceLineItemRepository) Update(item *models.InvoiceLineItem) error {
	return r.db.Save(item).Error
}

func (r *InvoiceLineItemRepository) Delete(id uint) error {
	return r.db.Delete(&models.InvoiceLineItem{}, id).Error
}
