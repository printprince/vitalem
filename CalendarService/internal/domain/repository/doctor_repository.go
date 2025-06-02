package repository

import (
	"context"
	"fmt"

	"CalendarService/internal/domain/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DoctorRepository interface {
	GetDoctorByID(ctx context.Context, id uuid.UUID) (*models.Doctor, error)
	GetDoctorByEmail(ctx context.Context, email string) (*models.Doctor, error)
	GetOrCreateDoctor(ctx context.Context, id uuid.UUID, email string) (*models.Doctor, error)
}

type doctorRepo struct {
	pool *pgxpool.Pool
}

func NewDoctorRepo(pool *pgxpool.Pool) DoctorRepository {
	return &doctorRepo{pool: pool}
}

func (r *doctorRepo) GetDoctorByID(ctx context.Context, id uuid.UUID) (*models.Doctor, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, email, created_at, updated_at FROM doctors WHERE id=$1`, id)
	var d models.Doctor
	if err := row.Scan(&d.ID, &d.Email, &d.CreatedAt, &d.UpdatedAt); err != nil {
		return nil, fmt.Errorf("doctor not found: %w", err)
	}
	return &d, nil
}

func (r *doctorRepo) GetDoctorByEmail(ctx context.Context, email string) (*models.Doctor, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, email, created_at, updated_at FROM doctors WHERE email=$1`, email)
	var d models.Doctor
	if err := row.Scan(&d.ID, &d.Email, &d.CreatedAt, &d.UpdatedAt); err != nil {
		return nil, fmt.Errorf("doctor not found: %w", err)
	}
	return &d, nil
}

// id, email, name обязательно передаются из токена/запроса
func (r *doctorRepo) GetOrCreateDoctor(ctx context.Context, id uuid.UUID, email string) (*models.Doctor, error) {
	// Пробуем найти доктора
	doc, err := r.GetDoctorByID(ctx, id)
	if err == nil {
		return doc, nil
	}

	// Если не найден — создаём нового
	row := r.pool.QueryRow(ctx,
		`INSERT INTO doctors (id, email, created_at, updated_at)
         VALUES ($1, $2, NOW(), NOW()) RETURNING id, email, created_at, updated_at`,
		id, email,
	)
	var d models.Doctor
	if err := row.Scan(&d.ID, &d.Email, &d.CreatedAt, &d.UpdatedAt); err != nil {
		return nil, fmt.Errorf("failed to create doctor: %w", err)
	}
	return &d, nil
}
