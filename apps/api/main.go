package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq"
)

var (
	buildLayerGeoJSONHashFn = BuildLayerGeoJSONHash
	newBlockchainServiceFn  = NewBlockchainService
	verifyHashOnChainFn     = func(svc *BlockchainService, ctx context.Context, layerType string, year uint16, currentGeoJSON []byte) (HashVerificationResult, error) {
		return svc.VerifyHashOnChain(ctx, layerType, year, currentGeoJSON)
	}
	commitHashToLedgerFn = func(svc *BlockchainService, ctx context.Context, layerType string, year uint16, sha256Hash string) (common.Hash, error) {
		return svc.CommitHashToLedger(ctx, layerType, year, sha256Hash)
	}
	appDB           *sql.DB
	defaultDataCity = "Mumbai"
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

type VerificationMetadata struct {
	IsVerified    bool   `json:"is_verified"`
	OnChainHash   string `json:"on_chain_hash"`
	StatusMessage string `json:"status_message"`
}

type LayerGeoJSON struct {
	Type     string    `json:"type"`
	Features []Feature `json:"features"`
}

type VerifiedLayerResponse struct {
	Data         LayerGeoJSON         `json:"data"`
	Verification VerificationMetadata `json:"verification"`
}

// FeatureCollection is the top-level GeoJSON response.
type FeatureCollection struct {
	Type                 string                `json:"type"`
	Features             []Feature             `json:"features"`
	VerificationMetadata *VerificationMetadata `json:"verification_metadata,omitempty"`
	TrustScore           int                   `json:"trust_score"`
	OnChainVerified      bool                  `json:"on_chain_verified"`
	VerifiedLayerType    string                `json:"verified_layer_type,omitempty"`
	OnChainTimestampUnix uint64                `json:"on_chain_timestamp_unix,omitempty"`
	OnChainSource        string                `json:"on_chain_source,omitempty"`
	ExpectedHash         string                `json:"expected_hash,omitempty"`
	OnChainHash          string                `json:"on_chain_hash,omitempty"`
	VerificationError    string                `json:"verification_error,omitempty"`
}

type commitLedgerRequest struct {
	LayerType string `json:"layer_type"`
	Year      int    `json:"year"`
	City      string `json:"city"`
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
	appDB = db
	defaultDataCity = strings.TrimSpace(getEnv("DEFAULT_CITY", "Mumbai"))

	app.Get("/api/v1/:city/layers", GetVerifiedLayer())
	app.Post("/api/v1/ledger/commit", adminAuthRequired(), NotarizeLayer())
	app.Post("/api/v1/admin/notarize", adminAuthRequired(), NotarizeLayer())
	app.Post("/api/admin/seal-history", adminAuthRequired(), NotarizeLayer())
	app.Get("/api/v1/ledger/status", handleLedgerStatus())
	app.Get("/api/v1/ledger/verify", handleVerifyLedger(db))

	port := getEnv("PORT", "4000")
	log.Fatal(app.Listen(":" + port))
}

func GetLayerData(layerType string, year int) (LayerGeoJSON, error) {
	return getLayerDataByCity(defaultDataCity, layerType, year)
}

func getLayerDataByCity(city, layerType string, year int) (LayerGeoJSON, error) {
	if appDB == nil {
		return LayerGeoJSON{}, errors.New("database not initialized")
	}
	if strings.TrimSpace(layerType) == "" {
		return LayerGeoJSON{}, errors.New("layer_type is required")
	}

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

	rows, err := appDB.Query(query, city, layerType, year)
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

func GetVerifiedLayer() fiber.Handler {
	return func(c *fiber.Ctx) error {
		city := strings.TrimSpace(c.Params("city"))
		if city == "" {
			city = defaultDataCity
		}

		layerType := strings.TrimSpace(c.Query("layer_type"))
		if layerType == "" {
			layerType = strings.TrimSpace(c.Query("trust_layer_type"))
		}
		if layerType == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "query parameter 'layer_type' is required"})
		}

		yearStr := strings.TrimSpace(c.Query("year"))
		if yearStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "query parameter 'year' is required"})
		}

		year, err := strconv.Atoi(yearStr)
		if err != nil || year < 1 || year > 9999 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "query parameter 'year' must be a valid 4-digit year"})
		}

		layerData, err := getLayerDataByCity(city, layerType, year)
		if err != nil {
			log.Printf("get layer data failed: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch layer data"})
		}

		verification := VerificationMetadata{
			IsVerified:    false,
			OnChainHash:   "",
			StatusMessage: "Verification Skipped",
		}

		if len(layerData.Features) == 0 {
			verification.StatusMessage = "Record Not Found"
			return c.JSON(VerifiedLayerResponse{Data: layerData, Verification: verification})
		}

		yearU16, err := parseYearToUint16(year)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "query parameter 'year' must be between 0 and 65535"})
		}

		payload, err := json.Marshal(layerData)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to encode layer payload"})
		}

		hash := GenerateSHA256Hash(payload)

		ctx, cancel := context.WithTimeout(c.UserContext(), 20*time.Second)
		defer cancel()

		svc, err := newBlockchainServiceFn(ctx)
		if err != nil {
			verification.StatusMessage = "Chain Unavailable"
			return c.JSON(VerifiedLayerResponse{Data: layerData, Verification: verification})
		}
		defer svc.Close()

		result, err := verifyHashOnChainFn(svc, ctx, layerType, yearU16, payload)
		if err != nil {
			verification.StatusMessage = "Verification Failed"
			return c.JSON(VerifiedLayerResponse{Data: layerData, Verification: verification})
		}

		verification.OnChainHash = result.OnChainHash
		if strings.TrimSpace(result.OnChainHash) == "" {
			verification.IsVerified = false
			verification.StatusMessage = "Record Not Found"
			return c.JSON(VerifiedLayerResponse{Data: layerData, Verification: verification})
		}

		verification.IsVerified = result.Match && strings.EqualFold(result.OnChainHash, hash)
		verification.StatusMessage = "Tamper Detected"
		if verification.IsVerified {
			verification.StatusMessage = "Match Found"
		}

		return c.JSON(VerifiedLayerResponse{Data: layerData, Verification: verification})
	}
}

func NotarizeLayer() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req commitLedgerRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
		}

		layerType := strings.TrimSpace(req.LayerType)
		if layerType == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "layer_type is required"})
		}

		year := req.Year
		yearU16, err := parseYearToUint16(year)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "year must be between 0 and 65535"})
		}

		city := strings.TrimSpace(req.City)
		if city == "" {
			city = defaultDataCity
		}

		layerData, err := getLayerDataByCity(city, layerType, year)
		if err != nil {
			log.Printf("notarize get layer data failed: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch layer data"})
		}
		if len(layerData.Features) == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "no artifacts found for requested city/layer/year"})
		}

		payload, err := json.Marshal(layerData)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to encode layer payload"})
		}
		hash := GenerateSHA256Hash(payload)

		ctx, cancel := context.WithTimeout(c.UserContext(), 20*time.Second)
		defer cancel()

		svc, err := newBlockchainServiceFn(ctx)
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "failed to initialize blockchain service"})
		}
		defer svc.Close()

		txHash, err := commitHashToLedgerFn(svc, ctx, layerType, yearU16, hash)
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "failed to commit hash to ledger", "details": err.Error()})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"status":      "committed",
			"city":        city,
			"layer_type":  layerType,
			"year":        yearU16,
			"sha256_hash": hash,
			"tx_hash":     txHash.Hex(),
		})
	}
}

// handleGetLayers returns a Fiber handler that queries urban_artifacts for the
// given city filtered by the requested year.
//
// Query parameters:
//
//	year (required) – the calendar year to filter on
//	layer_type      – optional comma-separated filter on layer_type
func handleGetLayers(db *sql.DB) fiber.Handler {
	appDB = db
	return GetVerifiedLayer()
}

func handleAdminSealHistory(db *sql.DB) fiber.Handler {
	appDB = db
	return NotarizeLayer()
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func handleCommitLedger(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req commitLedgerRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid JSON body",
			})
		}

		layerType := strings.TrimSpace(req.LayerType)
		if layerType == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "layer_type is required",
			})
		}

		year, err := parseYearToUint16(req.Year)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "year must be between 0 and 65535",
			})
		}

		city := strings.TrimSpace(req.City)
		if city == "" {
			city = "Mumbai"
		}

		ctx, cancel := context.WithTimeout(c.UserContext(), 20*time.Second)
		defer cancel()

		hash, payload, err := buildLayerGeoJSONHashFn(ctx, db, city, layerType, year)
		if err != nil {
			log.Printf("ledger commit hash build failed: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to build layer hash from database",
			})
		}

		if string(payload) == "[]" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "no artifacts found for requested city/layer/year",
			})
		}

		svc, err := newBlockchainServiceFn(ctx)
		if err != nil {
			log.Printf("ledger commit service init failed: %v", err)
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"error": "failed to initialize blockchain service",
			})
		}
		defer svc.Close()

		txHash, err := commitHashToLedgerFn(svc, ctx, layerType, year, hash)
		if err != nil {
			log.Printf("ledger commit failed: %v", err)
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"error":   "failed to commit hash to ledger",
				"details": err.Error(),
			})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"status":      "committed",
			"city":        city,
			"layer_type":  layerType,
			"year":        year,
			"sha256_hash": hash,
			"tx_hash":     txHash.Hex(),
		})
	}
}

func handleAdminNotarize(db *sql.DB) fiber.Handler {
	appDB = db
	return NotarizeLayer()
}

func adminAuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		expected := strings.TrimSpace(os.Getenv("ADMIN_API_KEY"))
		if expected == "" {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "admin route not configured: set ADMIN_API_KEY",
			})
		}

		provided := strings.TrimSpace(c.Get("X-Admin-Key"))
		if provided == "" {
			provided = parseBearerToken(c.Get("Authorization"))
		}

		if provided == "" || provided != expected {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized",
			})
		}

		return c.Next()
	}
}

func parseBearerToken(authHeader string) string {
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(strings.TrimSpace(authHeader), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func handleLedgerStatus() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
		defer cancel()

		svc, err := newBlockchainServiceFn(ctx)
		if err != nil {
			return c.JSON(fiber.Map{
				"connected": false,
				"error":     err.Error(),
			})
		}
		defer svc.Close()

		chainID := "unknown"
		if svc.chainID != nil {
			chainID = svc.chainID.String()
		}

		return c.JSON(fiber.Map{
			"connected":        true,
			"chain_id":         chainID,
			"contract_address": svc.contractAddress.Hex(),
			"source_ref":       svc.sourceRef,
		})
	}
}

func handleVerifyLedger(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		layerType := strings.TrimSpace(c.Query("layer_type"))
		if layerType == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "query parameter 'layer_type' is required",
			})
		}

		yearStr := strings.TrimSpace(c.Query("year"))
		if yearStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "query parameter 'year' is required",
			})
		}

		yearInt, err := strconv.Atoi(yearStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "query parameter 'year' must be a valid integer",
			})
		}

		year, err := parseYearToUint16(yearInt)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "query parameter 'year' must be between 0 and 65535",
			})
		}

		city := strings.TrimSpace(c.Query("city"))
		if city == "" {
			city = "Mumbai"
		}

		ctx, cancel := context.WithTimeout(c.UserContext(), 20*time.Second)
		defer cancel()

		hash, payload, err := buildLayerGeoJSONHashFn(ctx, db, city, layerType, year)
		if err != nil {
			log.Printf("ledger verify hash build failed: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to build layer hash from database",
			})
		}

		if string(payload) == "[]" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "no artifacts found for requested city/layer/year",
			})
		}

		svc, err := newBlockchainServiceFn(ctx)
		if err != nil {
			log.Printf("ledger verify service init failed: %v", err)
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"error": "failed to initialize blockchain service",
			})
		}
		defer svc.Close()

		result, err := verifyHashOnChainFn(svc, ctx, layerType, year, payload)
		if err != nil {
			log.Printf("ledger verify failed: %v", err)
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"error":   "failed to verify hash on chain",
				"details": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"city":                    city,
			"layer_type":              result.LayerType,
			"year":                    result.Year,
			"expected_hash":           hash,
			"on_chain_hash":           result.OnChainHash,
			"on_chain_source":         result.OnChainSource,
			"on_chain_timestamp_unix": result.OnChainUnixTime,
			"match":                   result.Match,
		})
	}
}
