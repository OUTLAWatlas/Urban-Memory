package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq"
)

// Feature represents a single GeoJSON feature returned by the API.
type Feature struct {
	Type       string          `json:"type"`
	Properties FeatureProps    `json:"properties"`
	Geometry   json.RawMessage `json:"geometry"`
}

// FeatureProps holds the non-spatial attributes of an urban artifact.
type FeatureProps struct {
	ID        int     `json:"id"`
	CityName  string  `json:"city_name"`
	LayerType string  `json:"layer_type"`
	ValidFrom string  `json:"valid_from"`
	ValidTo   *string `json:"valid_to"`
	SourceRef *string `json:"source_ref,omitempty"`
}

// FeatureCollection is the top-level GeoJSON response.
type FeatureCollection struct {
	Type     string    `json:"type"`
	Features []Feature `json:"features"`
}

func main() {
	dsn := getEnv("DATABASE_URL", "postgres://pilot_user:pilot_password@localhost:5432/urban_memory_backend?sslmode=disable")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("database unreachable: %v", err)
	}
	log.Println("connected to PostGIS database")

	app := fiber.New(fiber.Config{
		AppName: "UrbanMemory API",
	})

	app.Get("/api/v1/:city/layers", handleGetLayers(db))

	port := getEnv("PORT", "4000")
	log.Fatal(app.Listen(":" + port))
}

// handleGetLayers returns a Fiber handler that queries urban_artifacts for the
// given city filtered by the requested year.
//
// Query parameters:
//
//	year (required) – the calendar year to filter on
//	layer_type      – optional comma-separated filter on layer_type
func handleGetLayers(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		city := c.Params("city")

		yearStr := c.Query("year")
		if yearStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "query parameter 'year' is required",
			})
		}

		year, err := strconv.Atoi(yearStr)
		if err != nil || year < 1 || year > 9999 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "query parameter 'year' must be a valid 4-digit year",
			})
		}

		// Build a date representing Jan-1 of the requested year so we can
		// check whether it falls inside [valid_from, valid_to].
		// A NULL valid_to is treated as "currently active" (no end date).
		query := `
			SELECT
				id,
				city_name,
				layer_type,
				valid_from,
				valid_to,
				source_ref,
				ST_AsGeoJSON(geom) AS geojson
			FROM urban_artifacts
			WHERE city_name ILIKE $1
			  AND valid_from <= make_date($2, 12, 31)
			  AND (valid_to IS NULL OR valid_to >= make_date($2, 1, 1))
			ORDER BY layer_type, valid_from;
		`

		rows, err := db.Query(query, city, year)
		if err != nil {
			log.Printf("query error: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to query layers",
			})
		}
		defer rows.Close()

		features := make([]Feature, 0)

		for rows.Next() {
			var (
				id        int
				cityName  string
				layerType string
				validFrom string
				validTo   *string
				sourceRef *string
				geoJSON   string
			)

			if err := rows.Scan(&id, &cityName, &layerType, &validFrom, &validTo, &sourceRef, &geoJSON); err != nil {
				log.Printf("row scan error: %v", err)
				continue
			}

			features = append(features, Feature{
				Type: "Feature",
				Properties: FeatureProps{
					ID:        id,
					CityName:  cityName,
					LayerType: layerType,
					ValidFrom: validFrom,
					ValidTo:   validTo,
					SourceRef: sourceRef,
				},
				Geometry: json.RawMessage(geoJSON),
			})
		}

		if err := rows.Err(); err != nil {
			log.Printf("rows iteration error: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "error reading results",
			})
		}

		return c.JSON(FeatureCollection{
			Type:     "FeatureCollection",
			Features: features,
		})
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
