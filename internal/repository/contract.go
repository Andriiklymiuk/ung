package repository

import (
	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/Andriiklymiuk/ung/internal/models"
	"gorm.io/gorm"
)

type ContractRepository struct {
	db *gorm.DB
}

func NewContractRepository() *ContractRepository {
	return &ContractRepository{db: db.GormDB}
}

func (r *ContractRepository) Create(contract *models.Contract) error {
	return r.db.Create(contract).Error
}

func (r *ContractRepository) GetByID(id uint) (*models.Contract, error) {
	var contract models.Contract
	err := r.db.Preload("Client").First(&contract, id).Error
	if err != nil {
		return nil, err
	}
	return &contract, nil
}

func (r *ContractRepository) List() ([]models.Contract, error) {
	var contracts []models.Contract
	err := r.db.Preload("Client").Order("id DESC").Find(&contracts).Error
	return contracts, err
}

func (r *ContractRepository) ListActive() ([]models.Contract, error) {
	var contracts []models.Contract
	err := r.db.Preload("Client").Where("active = ?", true).Order("id DESC").Find(&contracts).Error
	return contracts, err
}

func (r *ContractRepository) Update(contract *models.Contract) error {
	return r.db.Save(contract).Error
}

func (r *ContractRepository) Delete(id uint) error {
	return r.db.Delete(&models.Contract{}, id).Error
}

func (r *ContractRepository) CountActive() (int64, error) {
	var count int64
	err := r.db.Model(&models.Contract{}).Where("active = ?", true).Count(&count).Error
	return count, err
}
