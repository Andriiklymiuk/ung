package repository

import (
	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/Andriiklymiuk/ung/internal/models"
	"gorm.io/gorm"
)

type CompanyRepository struct {
	db *gorm.DB
}

func NewCompanyRepository() *CompanyRepository {
	return &CompanyRepository{db: db.GormDB}
}

func (r *CompanyRepository) Create(company *models.Company) error {
	return r.db.Create(company).Error
}

func (r *CompanyRepository) GetByID(id uint) (*models.Company, error) {
	var company models.Company
	err := r.db.First(&company, id).Error
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *CompanyRepository) List() ([]models.Company, error) {
	var companies []models.Company
	err := r.db.Order("id DESC").Find(&companies).Error
	return companies, err
}

func (r *CompanyRepository) Update(company *models.Company) error {
	return r.db.Save(company).Error
}

func (r *CompanyRepository) Delete(id uint) error {
	return r.db.Delete(&models.Company{}, id).Error
}

func (r *CompanyRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.Company{}).Count(&count).Error
	return count, err
}
