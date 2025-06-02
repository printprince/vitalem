package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"CalendarService/internal/domain/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EventRepository interface {
	CreateEvent(ctx context.Context, event *models.Event) error
	GetEventByID(ctx context.Context, id uuid.UUID) (*models.Event, error)
	GetEventsBySpecialist(ctx context.Context, specialistID uuid.UUID) ([]*models.Event, error)
	UpdateEvent(ctx context.Context, event *models.Event) error
	DeleteEvent(ctx context.Context, id uuid.UUID) error
	BookEvent(ctx context.Context, eventID uuid.UUID, patientID uuid.UUID) error
	CancelEvent(ctx context.Context, eventID uuid.UUID) error
	GetAllEvents(ctx context.Context) ([]*models.Event, error)
}

type eventRepo struct {
	pool *pgxpool.Pool
}

func NewEventRepository(pool *pgxpool.Pool) EventRepository {
	return &eventRepo{pool: pool}
}

func (r *eventRepo) CreateEvent(ctx context.Context, event *models.Event) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO events 
		(id, title, description, start_time, end_time, specialist_id, patient_id, status, appointment_type, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		event.ID, event.Title, event.Description, event.StartTime, event.EndTime,
		event.SpecialistID, event.PatientID, event.Status, event.AppointmentType, event.CreatedAt, event.UpdatedAt,
	)
	return err
}

func (r *eventRepo) GetEventByID(ctx context.Context, id uuid.UUID) (*models.Event, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, title, description, start_time, end_time, specialist_id, patient_id, status, appointment_type, created_at, updated_at
		FROM events WHERE id=$1`, id,
	)
	event := &models.Event{}
	var patientID *uuid.UUID
	err := row.Scan(
		&event.ID, &event.Title, &event.Description, &event.StartTime, &event.EndTime,
		&event.SpecialistID, &patientID, &event.Status, &event.AppointmentType, &event.CreatedAt, &event.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	event.PatientID = patientID
	return event, nil
}

func (r *eventRepo) GetEventsBySpecialist(ctx context.Context, specialistID uuid.UUID) ([]*models.Event, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, title, description, start_time, end_time, specialist_id, patient_id, status, appointment_type, created_at, updated_at
		FROM events WHERE specialist_id=$1`, specialistID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		event := &models.Event{}
		var patientID *uuid.UUID
		if err := rows.Scan(
			&event.ID, &event.Title, &event.Description, &event.StartTime, &event.EndTime,
			&event.SpecialistID, &patientID, &event.Status, &event.AppointmentType, &event.CreatedAt, &event.UpdatedAt,
		); err != nil {
			return nil, err
		}
		event.PatientID = patientID
		events = append(events, event)
	}

	return events, nil
}

func (r *eventRepo) UpdateEvent(ctx context.Context, event *models.Event) error {
	event.UpdatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`UPDATE events SET
			title=$1,
			description=$2,
			start_time=$3,
			end_time=$4,
			specialist_id=$5,
			patient_id=$6,
			status=$7,
			appointment_type=$8,
			updated_at=$9
		WHERE id=$10`,
		event.Title,
		event.Description,
		event.StartTime,
		event.EndTime,
		event.SpecialistID,
		event.PatientID,
		event.Status,
		event.AppointmentType,
		event.UpdatedAt,
		event.ID,
	)
	return err
}

func (r *eventRepo) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM events WHERE id=$1`, id)
	return err
}

// Бронирование события (устанавливаем patient_id и меняем статус)
func (r *eventRepo) BookEvent(ctx context.Context, eventID uuid.UUID, patientID uuid.UUID) error {
	commandTag, err := r.pool.Exec(ctx,
		`UPDATE events SET patient_id=$1, status='booked', updated_at=$2 WHERE id=$3 AND status='available'`,
		patientID, time.Now(), eventID,
	)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return errors.New("event is not available for booking")
	}
	return nil
}

// Отмена бронирования события (очищаем patient_id и меняем статус)
func (r *eventRepo) CancelEvent(ctx context.Context, eventID uuid.UUID) error {
	commandTag, err := r.pool.Exec(ctx,
		`UPDATE events SET patient_id=NULL, status='canceled', updated_at=$1 WHERE id=$2 AND status='booked'`,
		time.Now(), eventID,
	)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return errors.New("event is not booked or already canceled")
	}
	return nil
}

func (r *eventRepo) GetAllEvents(ctx context.Context) ([]*models.Event, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, title, description, start_time, end_time, specialist_id, patient_id, status, appointment_type, created_at, updated_at
		FROM events`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		event := &models.Event{}
		var patientID *uuid.UUID
		if err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.Description,
			&event.StartTime,
			&event.EndTime,
			&event.SpecialistID,
			&patientID,
			&event.Status,
			&event.AppointmentType,
			&event.CreatedAt,
			&event.UpdatedAt,
		); err != nil {
			return nil, err
		}
		event.PatientID = patientID
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

// Получить события с фильтрацией по специалисту и статусу
func (r *eventRepo) GetEventsByFilter(ctx context.Context, specialistID uuid.UUID, status string) ([]*models.Event, error) {
	query := `SELECT id, title, description, start_time, end_time, specialist_id, patient_id, status, appointment_type, created_at, updated_at FROM events WHERE 1=1`
	args := []interface{}{}
	argPos := 1

	if specialistID != uuid.Nil {
		query += fmt.Sprintf(" AND specialist_id = $%d", argPos)
		args = append(args, specialistID)
		argPos++
	}

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, status)
		argPos++
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		event := &models.Event{}
		var patientID *uuid.UUID
		if err := rows.Scan(
			&event.ID, &event.Title, &event.Description, &event.StartTime, &event.EndTime,
			&event.SpecialistID, &patientID, &event.Status, &event.AppointmentType, &event.CreatedAt, &event.UpdatedAt,
		); err != nil {
			return nil, err
		}
		event.PatientID = patientID
		events = append(events, event)
	}

	return events, nil
}

// BeginTx начинает транзакцию
func (r *eventRepo) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}

func (r *eventRepo) GetEventByIDTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*models.Event, error) {
	row := tx.QueryRow(ctx,
		`SELECT id, title, description, start_time, end_time, specialist_id, patient_id, status, appointment_type, created_at, updated_at
		FROM events WHERE id=$1`, id,
	)
	event := &models.Event{}
	var patientID *uuid.UUID
	err := row.Scan(
		&event.ID, &event.Title, &event.Description, &event.StartTime, &event.EndTime,
		&event.SpecialistID, &patientID, &event.Status, &event.AppointmentType, &event.CreatedAt, &event.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	event.PatientID = patientID
	return event, nil
}

func (r *eventRepo) UpdateEventTx(ctx context.Context, tx pgx.Tx, event *models.Event) error {
	event.UpdatedAt = time.Now()
	_, err := tx.Exec(ctx,
		`UPDATE events SET
			title=$1, description=$2, start_time=$3, end_time=$4,
			specialist_id=$5, patient_id=$6, status=$7, appointment_type=$8, updated_at=$9
		WHERE id=$10`,
		event.Title, event.Description, event.StartTime, event.EndTime,
		event.SpecialistID, event.PatientID, event.Status, event.AppointmentType, event.UpdatedAt, event.ID,
	)
	return err
}
