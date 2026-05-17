package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// GetLayerDataByType queries PostGIS and returns GeoJSON features with consistent coordinate formatting
// for stable hashing. Coordinates are rounded to 6 decimal places (approximately 0.1 meter precision).
func GetLayerDataByType(ctx context.Context, db *sql.DB, cityName, layerType string, year int) (LayerGeoJSON, error) {
	if db == nil {
		return LayerGeoJSON{}, fmt.Errorf("database handle is nil")
	}

	if strings.TrimSpace(layerType) == "" {
		return LayerGeoJSON{}, fmt.Errorf("layer_type is required")
	}

	if strings.TrimSpace(cityName) == "" {
		cityName = defaultDataCity
	}

	// Use ST_SnapToGrid to consistently round coordinates to 6 decimal places (stable for hashing)
	query := `
		SELECT
			id,
			city_name,
			layer_type,
			valid_from,
			valid_to,
			source_ref,
			ST_AsGeoJSON(ST_SnapToGrid(geom, 0.000001), 6, 0) AS geojson
		FROM urban_artifacts
		WHERE city_name ILIKE $1
		  AND layer_type = $2
		  AND valid_from <= make_date($3, 12, 31)
		  AND (valid_to IS NULL OR valid_to >= make_date($3, 1, 1))
		ORDER BY id;
	`

	rows, err := db.QueryContext(ctx, query, cityName, layerType, year)
	if err != nil {
		return LayerGeoJSON{}, fmt.Errorf("query layer data: %w", err)
	}
	defer rows.Close()

	features := make([]Feature, 0)
	for rows.Next() {
		var (
			id        int
			cityName  string
			layerName string
			validFrom time.Time
			validTo   sql.NullTime
			sourceRef sql.NullString
			geoJSON   string
		)

		if err := rows.Scan(&id, &cityName, &layerName, &validFrom, &validTo, &sourceRef, &geoJSON); err != nil {
			return LayerGeoJSON{}, fmt.Errorf("scan layer row: %w", err)
		}

		validFromStr := validFrom.UTC().Format(time.RFC3339)
		var validToStr *string
		if validTo.Valid {
			formatted := validTo.Time.UTC().Format(time.RFC3339)
			validToStr = &formatted
		}

		var sourceRefPtr *string
		if sourceRef.Valid {
			src := sourceRef.String
			sourceRefPtr = &src
		}

		features = append(features, Feature{
			Type: "Feature",
			Properties: FeatureProps{
				ID:        id,
				CityName:  cityName,
				LayerType: layerName,
				ValidFrom: validFromStr,
				ValidTo:   validToStr,
				SourceRef: sourceRefPtr,
			},
			Geometry: json.RawMessage(geoJSON),
		})
	}

	if err := rows.Err(); err != nil {
		return LayerGeoJSON{}, fmt.Errorf("iterate layer rows: %w", err)
	}

	return LayerGeoJSON{Type: "FeatureCollection", Features: features}, nil
}

// GetLayerDataHashAndPayload queries PostGIS and returns both the SHA-256 hash and raw GeoJSON payload
// for efficient hashing with coordinate precision guarantees.
func GetLayerDataHashAndPayload(ctx context.Context, db *sql.DB, cityName, layerType string, year int) (string, []byte, error) {
	if db == nil {
		return "", nil, fmt.Errorf("database handle is nil")
	}

	if strings.TrimSpace(layerType) == "" {
		return "", nil, fmt.Errorf("layer_type is required")
	}

	if strings.TrimSpace(cityName) == "" {
		cityName = defaultDataCity
	}

	// Aggregate with consistent coordinate rounding for stable hashes
	query := `
		SELECT COALESCE(
			jsonb_agg(ST_AsGeoJSON(ST_SnapToGrid(geom, 0.000001), 6, 0)::jsonb ORDER BY id)::text,
			'[]'
		)
		FROM urban_artifacts
		WHERE city_name ILIKE $1
		  AND layer_type = $2
		  AND valid_from <= make_date($3, 12, 31)
		  AND (valid_to IS NULL OR valid_to >= make_date($3, 1, 1));
	`

	var payload string
	if err := db.QueryRowContext(ctx, query, cityName, layerType, year).Scan(&payload); err != nil {
		return "", nil, fmt.Errorf("query layer payload: %w", err)
	}

	data := []byte(payload)
	hash := GenerateSHA256Hash(data)
	return hash, data, nil
}

// LayerDataWithMetadata contains both the GeoJSON and verification metadata
type LayerDataWithMetadata struct {
	LayerData LayerGeoJSON
	Hash      string
	Payload   []byte
}

// GetLayerDataWithHash retrieves layer data and computes its hash atomically
func GetLayerDataWithHash(ctx context.Context, db *sql.DB, cityName, layerType string, year int) (LayerDataWithMetadata, error) {
	result := LayerDataWithMetadata{}

	layerData, err := GetLayerDataByType(ctx, db, cityName, layerType, year)
	if err != nil {
		return result, err
	}

	// Marshal to ensure consistent JSON encoding
	payload, err := json.Marshal(layerData)
	if err != nil {
		return result, fmt.Errorf("marshal layer data: %w", err)
	}

	hash := GenerateSHA256Hash(payload)
	result.LayerData = layerData
	result.Hash = hash
	result.Payload = payload

	return result, nil
}
