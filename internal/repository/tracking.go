package repository

import (
	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/Andriiklymiuk/ung/internal/models"
	"gorm.io/gorm"
)

type TrackingSessionRepository struct {
	db *gorm.DB
}

func NewTrackingSessionRepository() *TrackingSessionRepository {
	return &TrackingSessionRepository{db: db.GormDB}
}

func (r *TrackingSessionRepository) Create(session *models.TrackingSession) error {
	return r.db.Create(session).Error
}

func (r *TrackingSessionRepository) GetByID(id uint) (*models.TrackingSession, error) {
	var session models.TrackingSession
	err := r.db.Preload("Contract").Preload("Contract.Client").First(&session, id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *TrackingSessionRepository) List() ([]models.TrackingSession, error) {
	var sessions []models.TrackingSession
	err := r.db.Preload("Contract").Preload("Contract.Client").
		Order("start_time DESC").Find(&sessions).Error
	return sessions, err
}

func (r *TrackingSessionRepository) GetByContractID(contractID uint) ([]models.TrackingSession, error) {
	var sessions []models.TrackingSession
	err := r.db.Where("contract_id = ?", contractID).
		Order("start_time DESC").Find(&sessions).Error
	return sessions, err
}

func (r *TrackingSessionRepository) Update(session *models.TrackingSession) error {
	return r.db.Save(session).Error
}

func (r *TrackingSessionRepository) Delete(id uint) error {
	return r.db.Delete(&models.TrackingSession{}, id).Error
}

func (r *TrackingSessionRepository) GetActiveSession() (*models.TrackingSession, error) {
	var session models.TrackingSession
	err := r.db.Where("end_time IS NULL").First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}
