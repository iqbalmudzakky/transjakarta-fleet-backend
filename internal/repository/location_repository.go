package repository

import (
	"database/sql"
	"transjakarta-fleet-backend/internal/models"
)

type LocationRepository struct {
	DB *sql.DB
}

func NewLocationRepository(db *sql.DB) *LocationRepository {
	return &LocationRepository{DB: db}
}

func (r *LocationRepository) GetLatestLocation(vehicleID string) (*models.VehicleLocation, error) {
	query := `
		SELECT vehicle_id, latitude, longitude, timestamp
		FROM vehicle_locations
		WHERE vehicle_id = $1
		ORDER BY timestamp DESC
		LIMIT 1
	`

	var loc models.VehicleLocation
	err := r.DB.QueryRow(query, vehicleID).Scan(
		&loc.VehicleID, 
		&loc.Latitude, 
		&loc.Longitude, 
		&loc.Timestamp,
	)
	if err != nil {
		return nil, err
	}
	return &loc, nil
}

func (r *LocationRepository) GetLocationHistory(vehicleID string, start, end int64) ([]models.VehicleLocation, error) {
	query := `
		SELECT vehicle_id, latitude, longitude, timestamp
		FROM vehicle_locations
		WHERE vehicle_id = $1 AND timestamp BETWEEN $2 AND $3
		ORDER BY timestamp ASC
	`
	rows, err := r.DB.Query(query, vehicleID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.VehicleLocation
	for rows.Next() {
		var loc models.VehicleLocation
		if err := rows.Scan(
			&loc.VehicleID,
			&loc.Latitude,
			&loc.Longitude,
			&loc.Timestamp,
		); err != nil {
			return nil, err
		}
		results = append(results, loc)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *LocationRepository) InsertLocation(loc *models.VehicleLocation) error {
	query := `
		INSERT INTO vehicle_locations (vehicle_id, latitude, longitude, timestamp)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.DB.Exec(query, loc.VehicleID, loc.Latitude, loc.Longitude, loc.Timestamp)
	return err
}