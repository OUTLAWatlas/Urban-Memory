package controllers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/mail"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/urban-memory/api/services"
	"golang.org/x/crypto/bcrypt"
)

type AdminRole string

const (
	AdminRoleSuperAdmin AdminRole = "super_admin"
	AdminRoleApproved   AdminRole = "approved"
	AdminRolePending    AdminRole = "pending"

	BootstrapSuperAdminIdentifier = "super@localhost"
	BootstrapSuperAdminPassword   = "123"
	lockedAwaitingApprovalHash    = "LOCKED_AWAITING_APPROVAL"
	registrationSubmittedMessage  = "Access request successfully submitted. A system administrator will review your application."
)

type RegisterAdminRequest struct {
	Email string `json:"email"`
}

type ApproveAdminRequest struct {
	Email string `json:"email"`
}

type RequestApproveAdminOTPRequest struct {
	Email string `json:"email"`
}

type LoginAdminRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerAdminResponse struct {
	Message string `json:"message"`
}

type approveAdminResponse struct {
	Message string    `json:"message"`
	Email   string    `json:"email"`
	Role    AdminRole `json:"role"`
}

type loginAdminResponse struct {
	Message      string    `json:"message"`
	AdminID      int64     `json:"admin_id"`
	Email        string    `json:"email"`
	Role         AdminRole `json:"role"`
	SessionToken string    `json:"session_token"`
}

type adminRecord struct {
	ID           int64
	Email        string
	PasswordHash string
	Role         AdminRole
}

// AdminSessionClaims represents the verified admin session embedded in the mock token.
type AdminSessionClaims struct {
	AdminID   int64     `json:"admin_id"`
	Email     string    `json:"email"`
	Role      AdminRole `json:"role"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// RegisterAdmin creates a pending access request without issuing credentials.
func RegisterAdmin(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if db == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database is not initialized"})
		}

		var req RegisterAdminRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
		}

		email := strings.ToLower(strings.TrimSpace(req.Email))
		if email == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email is required"})
		}
		if _, err := mail.ParseAddress(email); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid email format"})
		}

		ctx := c.UserContext()
		existingAdmin, err := findAdminByEmail(ctx, db, email)
		if err == nil && existingAdmin != nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "admin already exists for this email"})
		}
		if err != nil && err != sql.ErrNoRows {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to check existing admin", "details": err.Error()})
		}

		if err := insertPendingAdminRequest(ctx, db, email); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create admin", "details": err.Error()})
		}

		return c.Status(fiber.StatusCreated).JSON(registerAdminResponse{
			Message: registrationSubmittedMessage,
		})
	}
}

// RequestApproveAdminOTP issues an OTP challenge to the acting super admin before governance approval.
func RequestApproveAdminOTP(db *sql.DB, mailer services.MailSender) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if db == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database is not initialized"})
		}
		if mailer == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "mail sender is not configured"})
		}

		superAdmin, err := requireSuperAdmin(c.UserContext(), db, c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}

		var req RequestApproveAdminOTPRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
		}

		email := strings.ToLower(strings.TrimSpace(req.Email))
		if email == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email is required"})
		}
		if _, err := mail.ParseAddress(email); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid email format"})
		}

		targetAdmin, err := findAdminByEmail(c.UserContext(), db, email)
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "pending admin request not found"})
		}
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load admin request", "details": err.Error()})
		}
		if targetAdmin.Role != AdminRolePending {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "admin request is not pending approval"})
		}

		expiresAt, err := services.IssueOTPChallenge(
			c.UserContext(),
			superAdmin.ID,
			superAdmin.Email,
			services.OTPPurposeApproveUser,
			services.OTPMetadata{"target_email": email},
			mailer,
			"HIGH PRIORITY: UrbanMemory Governance Approval OTP",
			func(code string, otpExpiresAt time.Time) string {
				return fmt.Sprintf(`UrbanMemory Governance Approval

You requested elevated approval for: %s

Verification code: %s
This code expires at %s UTC.

If you did not initiate this approval, revoke the session immediately.
`, email, code, otpExpiresAt.Format(time.RFC3339))
			},
		)
		if err != nil {
			log.Printf("request approve otp failed: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate approval otp", "details": err.Error()})
		}

		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
			"message":    "Governance OTP dispatched to your administrative inbox.",
			"email":      email,
			"expires_at": expiresAt.Format(time.RFC3339),
		})
	}
}

// ApproveAdmin promotes a pending user request and generates the first temporary password.
func ApproveAdmin(db *sql.DB, mailer services.MailSender) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if db == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database is not initialized"})
		}
		if mailer == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "mail sender is not configured"})
		}

		if _, err := requireSuperAdmin(c.UserContext(), db, c); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}

		var req ApproveAdminRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
		}

		email := strings.ToLower(strings.TrimSpace(req.Email))
		if email == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email is required"})
		}
		if _, err := mail.ParseAddress(email); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid email format"})
		}

		targetAdmin, err := findAdminByEmail(c.UserContext(), db, email)
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "pending admin request not found"})
		}
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load admin request", "details": err.Error()})
		}
		if targetAdmin.Role != AdminRolePending {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "admin request is not pending approval"})
		}

		temporaryPassword, err := generateTemporaryPassword(12)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate temporary password"})
		}

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(temporaryPassword), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to hash temporary password"})
		}

		if err := approvePendingAdminRequest(c.UserContext(), db, email, string(passwordHash)); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to approve admin request", "details": err.Error()})
		}

		subject := "UrbanMemory Account Activation"
		body := buildApprovalEmailBody(email, temporaryPassword)
		if err := mailer.SendEmail(email, subject, body); err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "failed to send activation email", "details": err.Error()})
		}

		return c.JSON(approveAdminResponse{
			Message: "Admin access approved. Temporary credentials have been dispatched.",
			Email:   email,
			Role:    AdminRoleApproved,
		})
	}
}

// LoginAdmin verifies credentials and returns a mock session token.
func LoginAdmin(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if db == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database is not initialized"})
		}

		var req LoginAdminRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
		}

		email := strings.ToLower(strings.TrimSpace(req.Email))
		password := strings.TrimSpace(req.Password)
		if email == "" || password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "username/email and password are required"})
		}
		if email == "super" {
			email = BootstrapSuperAdminIdentifier
		}

		ctx := c.UserContext()
		var (
			adminID     int64
			storedHash  string
			role        string
			storedEmail string
		)
		err := db.QueryRowContext(ctx, `SELECT id, email, password_hash, role FROM admins WHERE email = $1`, email).Scan(&adminID, &storedEmail, &storedHash, &role)
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
		}
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch admin", "details": err.Error()})
		}

		if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
		}

		sessionToken, err := generateMockJWT(adminID, storedEmail, role)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate session token"})
		}

		return c.JSON(loginAdminResponse{
			Message:      "admin authenticated",
			AdminID:      adminID,
			Email:        storedEmail,
			Role:         AdminRole(role),
			SessionToken: sessionToken,
		})
	}
}

func EnsureBootstrapSuperAdmin(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("database is not initialized")
	}

	schemaStatements := []string{
		`CREATE TABLE IF NOT EXISTS admins (
			id BIGSERIAL PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'pending'
				CHECK (role IN ('super_admin', 'approved', 'pending')),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_admins_email ON admins (email)`,
		`CREATE INDEX IF NOT EXISTS idx_admins_role ON admins (role)`,
		`CREATE TABLE IF NOT EXISTS otp_verifications (
			id BIGSERIAL PRIMARY KEY,
			admin_id BIGINT NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
			otp_hash TEXT NOT NULL,
			expires_at TIMESTAMPTZ NOT NULL,
			purpose TEXT NOT NULL DEFAULT 'notary',
			metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
			used BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`ALTER TABLE otp_verifications ADD COLUMN IF NOT EXISTS purpose TEXT NOT NULL DEFAULT 'notary'`,
		`ALTER TABLE otp_verifications ADD COLUMN IF NOT EXISTS metadata JSONB NOT NULL DEFAULT '{}'::jsonb`,
		`CREATE INDEX IF NOT EXISTS idx_otp_verifications_admin_id ON otp_verifications (admin_id)`,
		`CREATE INDEX IF NOT EXISTS idx_otp_verifications_expires_at ON otp_verifications (expires_at)`,
		`CREATE INDEX IF NOT EXISTS idx_otp_verifications_used ON otp_verifications (used)`,
		`CREATE INDEX IF NOT EXISTS idx_otp_verifications_purpose ON otp_verifications (purpose)`,
	}

	for _, statement := range schemaStatements {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("ensure admin auth schema: %w", err)
		}
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(BootstrapSuperAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash bootstrap super admin password: %w", err)
	}

	// Keep the bootstrap admin ID stable across restarts by renaming any old record in place.
	if _, err := db.ExecContext(ctx, `
		UPDATE admins
		SET email = $1,
			password_hash = $2,
			role = $3
		WHERE email = $4
	`, BootstrapSuperAdminIdentifier, string(passwordHash), string(AdminRoleSuperAdmin), "super"); err != nil {
		return fmt.Errorf("rename legacy bootstrap super admin: %w", err)
	}

	// Insert or update the bootstrap super admin with valid email
	if _, err := db.ExecContext(ctx, `
		INSERT INTO admins (email, password_hash, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (email) DO UPDATE
		SET password_hash = EXCLUDED.password_hash,
			role = EXCLUDED.role
	`, BootstrapSuperAdminIdentifier, string(passwordHash), string(AdminRoleSuperAdmin)); err != nil {
		return fmt.Errorf("upsert bootstrap super admin: %w", err)
	}

	return nil
}

func generateTemporaryPassword(length int) (string, error) {
	if length < 8 {
		length = 8
	}
	alphabet := []rune("ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789")
	result := make([]rune, length)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		result[i] = alphabet[n.Int64()]
	}
	return string(result), nil
}

func insertPendingAdminRequest(ctx context.Context, db *sql.DB, email string) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO admins (email, password_hash, role)
		VALUES ($1, $2, $3)
	`, email, lockedAwaitingApprovalHash, string(AdminRolePending))
	return err
}

func approvePendingAdminRequest(ctx context.Context, db *sql.DB, email, passwordHash string) error {
	result, err := db.ExecContext(ctx, `
		UPDATE admins
		SET password_hash = $2,
			role = $3
		WHERE email = $1
		  AND role = $4
	`, email, passwordHash, string(AdminRoleApproved), string(AdminRolePending))
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return fmt.Errorf("pending admin request not updated")
	}

	return nil
}

func findAdminByEmail(ctx context.Context, db *sql.DB, email string) (*adminRecord, error) {
	admin := &adminRecord{}
	err := db.QueryRowContext(ctx, `
		SELECT id, email, password_hash, role
		FROM admins
		WHERE email = $1
	`, email).Scan(&admin.ID, &admin.Email, &admin.PasswordHash, &admin.Role)
	if err != nil {
		return nil, err
	}
	return admin, nil
}

func findAdminByID(ctx context.Context, db *sql.DB, adminID int64) (*adminRecord, error) {
	admin := &adminRecord{}
	err := db.QueryRowContext(ctx, `
		SELECT id, email, password_hash, role
		FROM admins
		WHERE id = $1
	`, adminID).Scan(&admin.ID, &admin.Email, &admin.PasswordHash, &admin.Role)
	if err != nil {
		return nil, err
	}
	return admin, nil
}

func requireActiveAdmin(ctx context.Context, db *sql.DB, c *fiber.Ctx) (*adminRecord, error) {
	token := strings.TrimSpace(parseBearerToken(c.Get("Authorization")))
	if token == "" {
		token = strings.TrimSpace(c.Get("X-Session-Token"))
	}
	if token == "" {
		return nil, fmt.Errorf("expired authentication token")
	}

	claims, err := ValidateAdminSessionToken(token)
	if err != nil {
		return nil, err
	}

	admin, err := findAdminByID(ctx, db, claims.AdminID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("admin not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch admin: %w", err)
	}
	if !strings.EqualFold(admin.Email, claims.Email) {
		return nil, fmt.Errorf("session token email mismatch")
	}
	if admin.Role != AdminRoleApproved && admin.Role != AdminRoleSuperAdmin {
		return nil, fmt.Errorf("admin role not approved")
	}

	return admin, nil
}

func requireSuperAdmin(ctx context.Context, db *sql.DB, c *fiber.Ctx) (*adminRecord, error) {
	admin, err := requireActiveAdmin(ctx, db, c)
	if err != nil {
		return nil, err
	}
	if admin.Role != AdminRoleSuperAdmin {
		return nil, fmt.Errorf("super admin privileges are required")
	}

	return admin, nil
}

func describeOTPError(err error) string {
	switch {
	case errors.Is(err, services.ErrOTPNotFound):
		return "No active verification challenge found."
	case errors.Is(err, services.ErrOTPExpired):
		return "Verification window expired. Request a fresh code."
	case errors.Is(err, services.ErrOTPInvalid):
		return "Invalid verification code."
	default:
		return err.Error()
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

func generateMockJWT(adminID int64, email, role string) (string, error) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(`{"sub":"%d","email":"%s","role":"%s","iat":%d,"exp":%d}`, adminID, email, role, time.Now().UTC().Unix(), time.Now().UTC().Add(12*time.Hour).Unix())))
	secret := strings.TrimSpace(os.Getenv("ADMIN_SESSION_SECRET"))
	if secret == "" {
		secret = "urbanmemory-admin-session-secret"
	}
	h := sha256.Sum256([]byte(header + "." + payload + "." + secret))
	signature := hex.EncodeToString(h[:])
	return header + "." + payload + "." + signature, nil
}

func buildApprovalEmailBody(email string, temporaryPassword string) string {
	return fmt.Sprintf(`UrbanMemory Administrative Access Approved

Hello %s,

Your access request has been approved by a super administrator.

Temporary password: %s

Please sign in and rotate this credential immediately.
If you did not expect this message, contact a system administrator.
`, email, temporaryPassword)
}

// ValidateAdminSessionToken verifies the mock token signature and expiration.
func ValidateAdminSessionToken(token string) (AdminSessionClaims, error) {
	claims := AdminSessionClaims{}
	token = strings.TrimSpace(token)
	if token == "" {
		return claims, fmt.Errorf("session token is required")
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return claims, fmt.Errorf("invalid session token format")
	}

	secret := strings.TrimSpace(os.Getenv("ADMIN_SESSION_SECRET"))
	if secret == "" {
		secret = "urbanmemory-admin-session-secret"
	}

	computed := sha256.Sum256([]byte(parts[0] + "." + parts[1] + "." + secret))
	if subtle.ConstantTimeCompare([]byte(hex.EncodeToString(computed[:])), []byte(strings.TrimSpace(parts[2]))) != 1 {
		return claims, fmt.Errorf("invalid session token signature")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return claims, fmt.Errorf("invalid session token payload: %w", err)
	}

	var raw struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Role  string `json:"role"`
		Iat   int64  `json:"iat"`
		Exp   int64  `json:"exp"`
	}
	if err := json.Unmarshal(payloadBytes, &raw); err != nil {
		return claims, fmt.Errorf("invalid session token payload JSON: %w", err)
	}

	adminID, err := strconv.ParseInt(strings.TrimSpace(raw.Sub), 10, 64)
	if err != nil || adminID <= 0 {
		return claims, fmt.Errorf("invalid session token subject")
	}
	if strings.TrimSpace(raw.Email) == "" {
		return claims, fmt.Errorf("session token missing email")
	}
	role := AdminRole(strings.TrimSpace(raw.Role))
	if role != AdminRoleSuperAdmin && role != AdminRoleApproved && role != AdminRolePending {
		return claims, fmt.Errorf("invalid session token role")
	}

	now := time.Now().UTC()
	issuedAt := time.Unix(raw.Iat, 0).UTC()
	expiresAt := time.Unix(raw.Exp, 0).UTC()
	if raw.Exp <= 0 || now.After(expiresAt) {
		return claims, fmt.Errorf("session token expired")
	}

	claims = AdminSessionClaims{
		AdminID:   adminID,
		Email:     strings.TrimSpace(raw.Email),
		Role:      role,
		IssuedAt:  issuedAt,
		ExpiresAt: expiresAt,
	}
	return claims, nil
}
