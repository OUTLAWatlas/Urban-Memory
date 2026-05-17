package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/urban-memory/api/controllers"
	"github.com/yourusername/urban-memory/api/services"
)

type requestNotaryPayload struct {
	LayerType string `json:"layer_type"`
	Year      int    `json:"year"`
}

type confirmNotaryPayload struct {
	LayerType string `json:"layer_type"`
	Year      int    `json:"year"`
	AdminID   int64  `json:"admin_id"`
	OTPCode   string `json:"otp_code"`
	City      string `json:"city,omitempty"`
}

type notaryPipelineResult struct {
	City       string
	LayerType  string
	Year       uint16
	Hash       string
	IPFSCID    string
	TxHash     string
	FeatureSet LayerGeoJSON
}

var otpCodePattern = regexp.MustCompile(`^\d{6}$`)

// RequestNotary issues a short-lived OTP challenge for the logged-in admin.
func RequestNotary(db *sql.DB, mailer services.MailSender) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if db == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database is not initialized"})
		}
		if mailer == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "mail sender is not configured"})
		}

		var req requestNotaryPayload
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
		}

		layerType := strings.TrimSpace(req.LayerType)
		if layerType == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "layer_type is required"})
		}

		year, err := parseYearToUint16(req.Year)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "year must be between 0 and 65535"})
		}

		claims, admin, err := validateLoggedInAdmin(db, c)
		if err != nil {
			log.Printf("[OTP][REQUEST] rejected layer_type=%s year=%d reason=%v", layerType, year, err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}

		log.Printf("[OTP][REQUEST] admin_id=%d email=%s role=%s layer_type=%s year=%d", admin.ID, admin.Email, admin.Role, layerType, year)

		expiresAt, err := services.GenerateAndStoreOTP(c.UserContext(), claims.AdminID, admin.Email, mailer)
		if err != nil {
			log.Printf("[OTP][REQUEST] admin_id=%d store_failed=%v", admin.ID, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate otp", "details": err.Error()})
		}

		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
			"message":    "MFA Verification Required. OTP sent to your registered channel.",
			"admin_id":   admin.ID,
			"email":      admin.Email,
			"role":      admin.Role,
			"layer_type": layerType,
			"year":       year,
			"expires_at": expiresAt.Format(time.RFC3339),
		})
	}
}

// ConfirmNotary validates the OTP and then executes the IPFS + blockchain sealing pipeline.
func ConfirmNotary(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if db == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database is not initialized"})
		}

		var req confirmNotaryPayload
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
		}

		layerType := strings.TrimSpace(req.LayerType)
		if layerType == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "layer_type is required"})
		}

		otpCode := strings.TrimSpace(req.OTPCode)
		if !otpCodePattern.MatchString(otpCode) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "otp_code must be a 6-digit numeric string"})
		}

		year, err := parseYearToUint16(req.Year)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "year must be between 0 and 65535"})
		}

		claims, admin, err := validateLoggedInAdmin(db, c)
		if err != nil {
			log.Printf("[OTP][CONFIRM] rejected layer_type=%s year=%d reason=%v", layerType, year, err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}
		if req.AdminID <= 0 || req.AdminID != claims.AdminID {
			log.Printf("[OTP][CONFIRM] admin mismatch token_admin_id=%d payload_admin_id=%d", claims.AdminID, req.AdminID)
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "admin_id does not match authenticated session"})
		}

		log.Printf("[OTP][CONFIRM] admin_id=%d layer_type=%s year=%d validating code", admin.ID, layerType, year)

		if _, err := services.VerifyAndConsumeOTP(c.UserContext(), req.AdminID, otpCode, services.OTPPurposeNotary); err != nil {
			log.Printf("[OTP][CONFIRM] admin_id=%d otp_validation_failed=%v", req.AdminID, err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized / Code Expired"})
		}

		log.Printf("[OTP][CONFIRM] admin_id=%d otp_validated purpose=%s", req.AdminID, services.OTPPurposeNotary)

		city := strings.TrimSpace(req.City)
		if city == "" {
			city = defaultDataCity
		}

		result, err := executeDecentralizedNotaryPipeline(c.UserContext(), db, city, layerType, int(year))
		if err != nil {
			log.Printf("[PIPELINE][ERROR] admin_id=%d layer_type=%s year=%d err=%v", req.AdminID, layerType, year, err)
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": err.Error()})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"status":      "sealed_decentralized",
			"message":     "OTP verified and decentralized notarization completed",
			"city":        result.City,
			"layer_type":  result.LayerType,
			"year":        result.Year,
			"sha256_hash": result.Hash,
			"ipfs_cid":    result.IPFSCID,
			"tx_hash":     result.TxHash,
		})
	}
}

func validateLoggedInAdmin(db *sql.DB, c *fiber.Ctx) (*controllers.AdminSessionClaims, *struct {
	ID    int64
	Email string
	Role  string
}, error) {
	token := strings.TrimSpace(parseBearerToken(c.Get("Authorization")))
	if token == "" {
		token = strings.TrimSpace(c.Get("X-Session-Token"))
	}
	if token == "" {
		return nil, nil, fmt.Errorf("session token is required")
	}

	claims, err := controllers.ValidateAdminSessionToken(token)
	if err != nil {
		return nil, nil, err
	}

	var admin struct {
		ID    int64
		Email string
		Role  string
	}
	err = db.QueryRowContext(c.UserContext(), `SELECT id, email, role FROM admins WHERE id = $1`, claims.AdminID).Scan(&admin.ID, &admin.Email, &admin.Role)
	if err == sql.ErrNoRows {
		return nil, nil, fmt.Errorf("admin not found")
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch admin: %w", err)
	}
	if strings.ToLower(strings.TrimSpace(admin.Role)) != string(controllers.AdminRoleApproved) && strings.ToLower(strings.TrimSpace(admin.Role)) != string(controllers.AdminRoleSuperAdmin) {
		return nil, nil, fmt.Errorf("admin role not approved for notarization")
	}
	if !strings.EqualFold(admin.Email, claims.Email) {
		return nil, nil, fmt.Errorf("session token email mismatch")
	}

	return &claims, &admin, nil
}

func executeDecentralizedNotaryPipeline(ctx context.Context, db *sql.DB, city, layerType string, year int) (notaryPipelineResult, error) {
	result := notaryPipelineResult{City: city, LayerType: layerType}

	log.Printf("[PIPELINE][FETCH] city=%s layer_type=%s year=%d", city, layerType, year)
	layerData, err := GetLayerDataByType(ctx, db, city, layerType, year)
	if err != nil {
		return result, fmt.Errorf("failed to fetch layer data: %w", err)
	}
	if len(layerData.Features) == 0 {
		return result, fmt.Errorf("no artifacts found for requested city/layer/year")
	}

	payload, err := json.Marshal(layerData)
	if err != nil {
		return result, fmt.Errorf("failed to encode layer payload: %w", err)
	}

	result.Hash = GenerateSHA256Hash(payload)
	result.FeatureSet = layerData
	log.Printf("[PIPELINE][HASH] city=%s layer_type=%s year=%d sha256=%s bytes=%d", city, layerType, year, result.Hash, len(payload))

	log.Printf("[PIPELINE][IPFS] uploading payload to Pinata")
	ipfsCID, err := services.UploadToIPFS(payload)
	if err != nil {
		return result, fmt.Errorf("failed to upload payload to ipfs: %w", err)
	}
	result.IPFSCID = ipfsCID
	log.Printf("[PIPELINE][IPFS] cid=%s", ipfsCID)

	yearU16, err := parseYearToUint16(year)
	if err != nil {
		return result, err
	}

	svc, err := newBlockchainServiceFn(ctx)
	if err != nil {
		return result, fmt.Errorf("%s: %w", blockchainUnavailableMessage(err), err)
	}
	defer svc.Close()

	log.Printf("[PIPELINE][CHAIN] committing hash and cid to ledger")
	txHash, err := commitHashToLedgerFn(svc, ctx, layerType, yearU16, result.Hash, result.IPFSCID)
	if err != nil {
		return result, fmt.Errorf("failed to commit hash and cid to ledger: %w", err)
	}
	result.TxHash = txHash.Hex()
	log.Printf("[PIPELINE][DONE] tx_hash=%s cid=%s sha256=%s", result.TxHash, result.IPFSCID, result.Hash)

	return result, nil
}

