package repository

import (
	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/Andriiklymiuk/ung/internal/models"
	"gorm.io/gorm"
)

type ClientRepository struct {
	db *gorm.DB
}

func NewClientRepository() *ClientRepository {
	return &ClientRepository{db: db.GormDB}
}

func (r *ClientRepository) Create(client *models.Client) error {
	return r.db.Create(client).Error
}

func (r *ClientRepository) GetByID(id uint) (*models.Client, error) {
	var client models.Client
	err := r.db.First(&client, id).Error
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func (r *ClientRepository) List() ([]models.Client, error) {
	var clients []models.Client
	err := r.db.Order("id DESC").Find(&clients).Error
	return clients, err
}

func (r *ClientRepository) Update(client *models.Client) error {
	return r.db.Save(client).Error
}

func (r *ClientRepository) Delete(id uint) error {
	return r.db.Delete(&models.Client{}, id).Error
}

func (r *ClientRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.Client{}).Count(&count).Error
	return count, err
}
