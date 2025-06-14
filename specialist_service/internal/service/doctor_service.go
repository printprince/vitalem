package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/printprince/vitalem/logger_service/pkg/logger"
	"github.com/printprince/vitalem/specialist_service/internal/models"
	"github.com/printprince/vitalem/specialist_service/internal/repository"
)

type DoctorService interface {
	CreateDoctor(ctx context.Context, req *models.DoctorCreateRequest) (*models.DoctorResponse, error)
	GetDoctorByID(ctx context.Context, id uuid.UUID) (*models.DoctorResponse, error)
	GetDoctorByUserID(ctx context.Context, userID uuid.UUID) (*models.DoctorResponse, error)
	GetAllDoctors(ctx context.Context) ([]*models.DoctorResponse, error)
	UpdateDoctor(ctx context.Context, id uuid.UUID, req *models.DoctorCreateRequest) (*models.DoctorResponse, error)
	UpdateDoctorProfile(ctx context.Context, userID uuid.UUID, req *models.DoctorCreateRequest) (*models.DoctorResponse, error)
	DeleteDoctor(ctx context.Context, id uuid.UUID) error
}

type doctorService struct {
	doctorRepo repository.DoctorRepository
	logger     *logger.Client
}

func NewDoctorService(doctorRepo repository.DoctorRepository, logger *logger.Client) DoctorService {
	return &doctorService{
		doctorRepo: doctorRepo,
		logger:     logger,
	}
}

func (s *doctorService) CreateDoctor(ctx context.Context, req *models.DoctorCreateRequest) (*models.DoctorResponse, error) {
	doctor := req.ToDoctor()

	createdDoctor, err := s.doctorRepo.Create(ctx, doctor)
	if err != nil {
		s.logger.Error("Failed to create doctor", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	return createdDoctor.ToDoctorResponse(), nil
}

func (s *doctorService) GetDoctorByID(ctx context.Context, id uuid.UUID) (*models.DoctorResponse, error) {
	doctor, err := s.doctorRepo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get doctor by ID", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
		return nil, err
	}

	if doctor == nil {
		return nil, nil
	}

	return doctor.ToDoctorResponse(), nil
}

func (s *doctorService) GetDoctorByUserID(ctx context.Context, userID uuid.UUID) (*models.DoctorResponse, error) {
	doctor, err := s.doctorRepo.FindByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get doctor by user ID", map[string]interface{}{
			"error":  err.Error(),
			"userID": userID,
		})
		return nil, err
	}

	if doctor == nil {
		return nil, nil
	}

	return doctor.ToDoctorResponse(), nil
}

func (s *doctorService) GetAllDoctors(ctx context.Context) ([]*models.DoctorResponse, error) {
	doctors, err := s.doctorRepo.FindAll(ctx)
	if err != nil {
		s.logger.Error("Failed to get all doctors", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	var response []*models.DoctorResponse
	for _, doctor := range doctors {
		response = append(response, doctor.ToDoctorResponse())
	}

	return response, nil
}

func (s *doctorService) UpdateDoctor(ctx context.Context, id uuid.UUID, req *models.DoctorCreateRequest) (*models.DoctorResponse, error) {
	doctor, err := s.doctorRepo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to find doctor for update", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
		return nil, err
	}

	if doctor == nil {
		return nil, nil
	}

	// Обновляем поля
	doctor.FirstName = req.FirstName
	doctor.MiddleName = req.MiddleName
	doctor.LastName = req.LastName
	doctor.Description = req.Description
	doctor.Email = req.Email
	doctor.Phone = req.Phone
	doctor.AvatarURL = req.AvatarURL
	doctor.Roles = req.Roles
	doctor.Price = req.Price
	doctor.Education = req.Education
	doctor.Certificates = req.Certificates

	updatedDoctor, err := s.doctorRepo.Update(ctx, doctor)
	if err != nil {
		s.logger.Error("Failed to update doctor", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
		return nil, err
	}

	return updatedDoctor.ToDoctorResponse(), nil
}

// UpdateDoctorProfile обновляет профиль врача по ID пользователя
func (s *doctorService) UpdateDoctorProfile(ctx context.Context, userID uuid.UUID, req *models.DoctorCreateRequest) (*models.DoctorResponse, error) {
	doctor, err := s.doctorRepo.FindByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to find doctor by user ID for profile update", map[string]interface{}{
			"error":  err.Error(),
			"userID": userID,
		})
		return nil, err
	}

	if doctor == nil {
		// Если профиль не найден, но у нас есть данные для создания, создаем новый
		s.logger.Info("Creating new doctor profile for user", map[string]interface{}{
			"userID": userID,
		})

		newDoctor := req.ToDoctor()
		newDoctor.UserID = userID // Устанавливаем ID пользователя

		createdDoctor, err := s.doctorRepo.Create(ctx, newDoctor)
		if err != nil {
			s.logger.Error("Failed to create doctor profile", map[string]interface{}{
				"error":  err.Error(),
				"userID": userID,
			})
			return nil, err
		}

		return createdDoctor.ToDoctorResponse(), nil
	}

	// Обновляем поля существующего профиля
	doctor.FirstName = req.FirstName
	doctor.MiddleName = req.MiddleName
	doctor.LastName = req.LastName
	doctor.Description = req.Description
	doctor.Email = req.Email
	doctor.Phone = req.Phone
	doctor.AvatarURL = req.AvatarURL
	doctor.Roles = req.Roles
	doctor.Price = req.Price
	doctor.Education = req.Education
	doctor.Certificates = req.Certificates

	updatedDoctor, err := s.doctorRepo.Update(ctx, doctor)
	if err != nil {
		s.logger.Error("Failed to update doctor profile", map[string]interface{}{
			"error":  err.Error(),
			"userID": userID,
		})
		return nil, err
	}

	s.logger.Info("Doctor profile updated successfully", map[string]interface{}{
		"userID":   userID,
		"doctorID": doctor.ID,
	})

	return updatedDoctor.ToDoctorResponse(), nil
}

func (s *doctorService) DeleteDoctor(ctx context.Context, id uuid.UUID) error {
	err := s.doctorRepo.Delete(ctx, id)
	if err != nil {
		s.logger.Error("Failed to delete doctor", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
		return err
	}
	return nil
}
