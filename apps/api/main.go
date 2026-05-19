package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/yourusername/urban-memory/api/controllers"
	"github.com/yourusername/urban-memory/api/services"
)

var (
	buildLayerGeoJSONHashFn = BuildLayerGeoJSONHash
	newBlockchainServiceFn  = NewBlockchainService
	verifyHashOnChainFn     = func(svc *BlockchainService, ctx context.Context, layerType string, year uint16, currentGeoJSON []byte) (HashVerificationResult, error) {
		return svc.VerifyHashOnChain(ctx, layerType, year, currentGeoJSON)
	}
	commitHashToLedgerFn = func(svc *BlockchainService, ctx context.Context, layerType string, year uint16, sha256Hash string, ipfsCID string) (common.Hash, error) {
		return svc.CommitHashToLedger(ctx, layerType, year, sha256Hash, ipfsCID)
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
	if err := godotenv.Load(); err != nil {
		log.Printf(".env not loaded (continuing with OS env): %v", err)
	}

	// Auto-load contract address from Hardhat deployment if not set
	if os.Getenv("LEDGER_CONTRACT_ADDRESS") == "" {
		if addr := loadDeployedContractAddress(); addr != "" {
			os.Setenv("LEDGER_CONTRACT_ADDRESS", addr)
			log.Printf("📝 Auto-loaded contract address from deployment: %s", addr)
		}
	}

	dsn := getEnv("DATABASE_URL", "postgres://pilot_user:pilot_password@localhost:5432/urban_memory_backend?sslmode=disable")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("database unreachable: %v", err)
	}
	mailer, err := services.NewSMTPMailerFromEnv()
	if err != nil {
		log.Fatalf("mail service configuration failed: %v", err)
	}
	log.Println("connected to PostGIS database")
	if err := controllers.EnsureBootstrapSuperAdmin(context.Background(), db); err != nil {
		log.Fatalf("failed to provision bootstrap super admin: %v", err)
	}
	log.Printf("bootstrap super admin ensured: username=%s role=%s",
		controllers.BootstrapSuperAdminIdentifier,
		controllers.AdminRoleSuperAdmin,
	)
	services.ConfigureOTPStorage(db)

	app := fiber.New(fiber.Config{
		AppName: "UrbanMemory API",
	})
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: func(_ string) bool { return true },
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Requested-With",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: true,
	}))
	app.Options("/*", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})
	appDB = db
	defaultDataCity = strings.TrimSpace(getEnv("DEFAULT_CITY", "Mumbai"))

	app.Post("/api/admin/register", controllers.RegisterAdmin(db))
	app.Post("/api/admin/login", controllers.LoginAdmin(db))
	app.Get("/api/admin/pending-users", controllers.ListPendingAdmins(db))
	app.Post("/api/admin/request-approve-user", controllers.RequestApproveAdminOTP(db, mailer))
	app.Post("/api/admin/approve-user", controllers.ApproveAdmin(db, mailer))
	app.Post("/api/v1/admin/request-password-change", controllers.RequestPasswordChange(db, mailer))
	app.Post("/api/v1/admin/confirm-password-change", controllers.ConfirmPasswordChange(db))

	app.Get("/api/v1/:city/layers", GetVerifiedLayer())
	app.Post("/api/v1/ledger/commit", adminAuthRequired(), NotarizeLayer())
	app.Post("/api/v1/admin/notarize", adminAuthRequired(), NotarizeLayer())
	app.Post("/api/v1/admin/request-notary", RequestNotary(db, mailer))
	app.Post("/api/v1/admin/confirm-notary", ConfirmNotary(db))
	app.Post("/api/v1/admin/seal-decentralized", ConfirmNotary(db))
	app.Post("/api/admin/seal-history", adminAuthRequired(), NotarizeLayer())
	app.Get("/api/v1/ledger/status", handleLedgerStatus())
	app.Get("/api/v1/ledger/verify", handleVerifyLedger(db))

	port := getEnv("PORT", "4000")
	if err := app.Listen(":" + port); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("failed to listen: %v", err)
	}
}

func GetLayerData(layerType string, year int) (LayerGeoJSON, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return GetLayerDataByType(ctx, appDB, defaultDataCity, layerType, year)
}

func GetVerifiedLayer() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse and validate query parameters
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

		yearU16, err := parseYearToUint16(year)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "query parameter 'year' must be between 0 and 65535"})
		}

		ctx, cancel := context.WithTimeout(c.UserContext(), 20*time.Second)
		defer cancel()

		// Fetch layer data using the new database helper
		layerData, err := GetLayerDataByType(ctx, appDB, city, layerType, year)
		if err != nil {
			log.Printf("get layer data failed: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch layer data"})
		}

		// Initialize default verification metadata
		verification := VerificationMetadata{
			IsVerified:    false,
			OnChainHash:   "",
			StatusMessage: "Verification Pending",
		}

		// Handle case where no features are found
		if len(layerData.Features) == 0 {
			verification.StatusMessage = "Record Not Found"
			return c.JSON(VerifiedLayerResponse{Data: layerData, Verification: verification})
		}

		// Marshal to compute hash
		payload, err := json.Marshal(layerData)
		if err != nil {
			log.Printf("marshal layer data failed: %v", err)
			verification.StatusMessage = "Serialization Error"
			return c.JSON(VerifiedLayerResponse{Data: layerData, Verification: verification})
		}

		expectedHash := GenerateSHA256Hash(payload)

		// Attempt blockchain verification (gracefully handle unavailability)
		svc, err := newBlockchainServiceFn(ctx)
		if err != nil {
			if !isBlockchainUnavailableError(err) {
				log.Printf("blockchain service init failed: %v", err)
			}
			verification.StatusMessage = "Chain Unavailable"
			verification.OnChainHash = fmt.Sprintf("(%s)", blockchainUnavailableMessage(err))
			return c.JSON(VerifiedLayerResponse{Data: layerData, Verification: verification})
		}
		defer svc.Close()

		// Verify hash on chain
		result, err := verifyHashOnChainFn(svc, ctx, layerType, yearU16, payload)
		if err != nil {
			log.Printf("blockchain verification failed: %v", err)
			// Check if this is a "Record Not Found" error from the blockchain
			if strings.Contains(err.Error(), "No cryptographic record found") {
				verification.StatusMessage = "Record Not Found"
				verification.OnChainHash = "(not notarized)"
				return c.JSON(VerifiedLayerResponse{Data: layerData, Verification: verification})
			}
			verification.StatusMessage = "Verification Error"
			verification.OnChainHash = "(verification failed)"
			return c.JSON(VerifiedLayerResponse{Data: layerData, Verification: verification})
		}

		// Process verification result
		verification.OnChainHash = result.OnChainHash
		if strings.TrimSpace(result.OnChainHash) == "" {
			verification.IsVerified = false
			verification.StatusMessage = "Record Not Found"
		} else {
			// Check if hashes match
			verification.IsVerified = result.Match && strings.EqualFold(result.OnChainHash, expectedHash)
			if verification.IsVerified {
				verification.StatusMessage = "Cryptographically Secured"
			} else {
				verification.StatusMessage = "Tamper Detected"
			}
		}

		return c.JSON(VerifiedLayerResponse{Data: layerData, Verification: verification})
	}
}

func NotarizeLayer() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse and validate request body
		var req commitLedgerRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
		}

		layerType := strings.TrimSpace(req.LayerType)
		if layerType == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "layer_type is required"})
		}

		yearU16, err := parseYearToUint16(req.Year)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "year must be between 0 and 65535"})
		}

		city := strings.TrimSpace(req.City)
		if city == "" {
			city = defaultDataCity
		}

		ctx, cancel := context.WithTimeout(c.UserContext(), 20*time.Second)
		defer cancel()

		// Fetch layer data using the new database helper
		layerData, err := GetLayerDataByType(ctx, appDB, city, layerType, req.Year)
		if err != nil {
			log.Printf("notarize get layer data failed: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch layer data"})
		}

		if len(layerData.Features) == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "no artifacts found for requested city/layer/year"})
		}

		// Marshal to compute hash
		payload, err := json.Marshal(layerData)
		if err != nil {
			log.Printf("notarize marshal failed: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to encode layer payload"})
		}

		sha256Hash := GenerateSHA256Hash(payload)

		// Initialize blockchain service and commit hash
		svc, err := newBlockchainServiceFn(ctx)
		if err != nil {
			if !isBlockchainUnavailableError(err) {
				log.Printf("notarize blockchain service init failed: %v", err)
			}
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": blockchainUnavailableMessage(err)})
		}
		defer svc.Close()

		txHash, err := commitHashToLedgerFn(svc, ctx, layerType, yearU16, sha256Hash, "legacy-no-cid")
		if err != nil {
			log.Printf("notarize commit failed: %v", err)
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"error":   "failed to commit hash to ledger",
				"details": err.Error(),
			})
		}

		log.Printf("successfully notarized: city=%s layer_type=%s year=%d hash=%s tx=%s",
			city, layerType, req.Year, sha256Hash, txHash.Hex())

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"status":      "committed",
			"city":        city,
			"layer_type":  layerType,
			"year":        yearU16,
			"sha256_hash": sha256Hash,
			"tx_hash":     txHash.Hex(),
			"message":     "Layer successfully notarized on blockchain",
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

// loadDeployedContractAddress reads the contract address from the Hardhat deployment JSON file.
// It looks for the file at ../../contracts/deployed-address.json (relative to the API binary location).
func loadDeployedContractAddress() string {
	// Try relative path from current working directory
	deploymentPath := filepath.Join("..", "..", "contracts", "deployed-address.json")

	data, err := os.ReadFile(deploymentPath)
	if err != nil {
		// File not found or unreadable; no contract address auto-loaded
		return ""
	}

	var deployment struct {
		Address string `json:"address"`
	}
	if err := json.Unmarshal(data, &deployment); err != nil {
		log.Printf("⚠️ Failed to parse deployment address file: %v", err)
		return ""
	}

	return strings.TrimSpace(deployment.Address)
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
			city = defaultDataCity
		}

		ctx, cancel := context.WithTimeout(c.UserContext(), 20*time.Second)
		defer cancel()

		hash, payload, err := GetLayerDataHashAndPayload(ctx, db, city, layerType, req.Year)
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
			if !isBlockchainUnavailableError(err) {
				log.Printf("ledger commit service init failed: %v", err)
			}
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": blockchainUnavailableMessage(err),
			})
		}
		defer svc.Close()

		txHash, err := commitHashToLedgerFn(svc, ctx, layerType, year, hash, "legacy-no-cid")
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

func SealDecentralizedLayer() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req commitLedgerRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
		}

		layerType := strings.TrimSpace(req.LayerType)
		if layerType == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "layer_type is required"})
		}

		yearU16, err := parseYearToUint16(req.Year)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "year must be between 0 and 65535"})
		}

		city := strings.TrimSpace(req.City)
		if city == "" {
			city = defaultDataCity
		}

		ctx, cancel := context.WithTimeout(c.UserContext(), 60*time.Second)
		defer cancel()

		layerData, err := GetLayerDataByType(ctx, appDB, city, layerType, req.Year)
		if err != nil {
			log.Printf("seal-decentralized get layer data failed: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch layer data"})
		}
		if len(layerData.Features) == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "no artifacts found for requested city/layer/year"})
		}

		payload, err := json.Marshal(layerData)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to encode layer payload"})
		}

		sha256Hash := GenerateSHA256Hash(payload)

		ipfsCID, err := services.UploadToIPFS(payload)
		if err != nil {
			log.Printf("seal-decentralized pinata upload failed: %v", err)
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"error":   "failed to upload payload to IPFS",
				"details": err.Error(),
			})
		}

		svc, err := newBlockchainServiceFn(ctx)
		if err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": blockchainUnavailableMessage(err)})
		}
		defer svc.Close()

		txHash, err := commitHashToLedgerFn(svc, ctx, layerType, yearU16, sha256Hash, ipfsCID)
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"error":   "failed to commit hash and CID to ledger",
				"details": err.Error(),
			})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"status":      "sealed_decentralized",
			"city":        city,
			"layer_type":  layerType,
			"year":        yearU16,
			"sha256_hash": sha256Hash,
			"ipfs_cid":    ipfsCID,
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
				"disabled":  errors.Is(err, ErrBlockchainDisabled),
				"error":     blockchainUnavailableMessage(err),
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
			city = defaultDataCity
		}

		ctx, cancel := context.WithTimeout(c.UserContext(), 20*time.Second)
		defer cancel()

		hash, payload, err := GetLayerDataHashAndPayload(ctx, db, city, layerType, yearInt)
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
			if !isBlockchainUnavailableError(err) {
				log.Printf("ledger verify service init failed: %v", err)
			}
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": blockchainUnavailableMessage(err),
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
			"on_chain_ipfs_cid":       result.OnChainIPFSCID,
			"on_chain_source":         result.OnChainSource,
			"on_chain_timestamp_unix": result.OnChainUnixTime,
			"match":                   result.Match,
		})
	}
}
