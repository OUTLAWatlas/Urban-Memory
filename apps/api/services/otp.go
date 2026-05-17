package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"sync"
	"time"
)

var (
	opmMu sync.RWMutex
	opmDB *sql.DB

	ErrOTPNotFound = errors.New("otp challenge not found")
	ErrOTPExpired  = errors.New("otp challenge expired")
	ErrOTPInvalid  = errors.New("otp challenge invalid")
)

type OTPPurpose string

const (
	OTPPurposeNotary         OTPPurpose = "notary"
	OTPPurposePasswordChange OTPPurpose = "password_change"
	OTPPurposeApproveUser    OTPPurpose = "approve_user"
)

type OTPMetadata map[string]string

// ConfigureOTPStorage sets the database handle used by GenerateAndStoreOTP.
func ConfigureOTPStorage(db *sql.DB) {
	opmMu.Lock()
	defer opmMu.Unlock()
	opmDB = db
}

// GenerateAndStoreOTP creates a 6-digit OTP challenge for notary validation.
func GenerateAndStoreOTP(ctx context.Context, adminID int64, recipientEmail string, mailer MailSender) (time.Time, error) {
	return IssueOTPChallenge(ctx, adminID, recipientEmail, OTPPurposeNotary, nil, mailer, "HIGH PRIORITY: UrbanMemory OTP Verification Code", buildOTPEmailBody)
}

// IssueOTPChallenge creates a 6-digit OTP, stores the SHA-256 hash with purpose-scoped metadata,
// then delivers the raw code to the admin's registered email address.
func IssueOTPChallenge(
	ctx context.Context,
	adminID int64,
	recipientEmail string,
	purpose OTPPurpose,
	metadata OTPMetadata,
	mailer MailSender,
	subject string,
	bodyBuilder func(code string, expiresAt time.Time) string,
) (time.Time, error) {
	if adminID <= 0 {
		return time.Time{}, fmt.Errorf("adminID must be positive")
	}

	opmMu.RLock()
	db := opmDB
	opmMu.RUnlock()
	if db == nil {
		return time.Time{}, fmt.Errorf("otp storage database is not configured")
	}
	if mailer == nil {
		return time.Time{}, fmt.Errorf("mail sender is not configured")
	}
	if strings.TrimSpace(recipientEmail) == "" {
		return time.Time{}, fmt.Errorf("recipient email is required")
	}
	if strings.TrimSpace(subject) == "" {
		return time.Time{}, fmt.Errorf("email subject is required")
	}
	if bodyBuilder == nil {
		return time.Time{}, fmt.Errorf("email body builder is required")
	}

	rawCode, err := generateNumericOTP(6)
	if err != nil {
		return time.Time{}, fmt.Errorf("generate otp: %w", err)
	}

	now := time.Now().UTC()
	expiresAt := now.Add(5 * time.Minute)
	hash := hashOTP(rawCode)
	if metadata == nil {
		metadata = OTPMetadata{}
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return time.Time{}, fmt.Errorf("marshal otp metadata: %w", err)
	}

	// Insert the OTP record into storage
	log.Printf("[OTP][DEBUG] inserting otp record admin_id=%d purpose=%s email=%s expires_at=%s", adminID, purpose, strings.TrimSpace(recipientEmail), expiresAt.Format(time.RFC3339))
	_, err = db.ExecContext(ctx, `
		INSERT INTO otp_verifications (admin_id, otp_hash, expires_at, used, purpose, metadata)
		VALUES ($1, $2, $3, FALSE, $4, $5::jsonb)
	`, adminID, hash, expiresAt, string(purpose), string(metadataJSON))
	if err != nil {
		log.Printf("[OTP][ERROR] insert otp verification failed admin_id=%d purpose=%s email=%s: %v", adminID, purpose, strings.TrimSpace(recipientEmail), err)
		return time.Time{}, fmt.Errorf("insert otp verification: %w", err)
	}

	// Build email body and send
	body := buildOTPEmailBody(rawCode, expiresAt)
	if bodyBuilder != nil {
		body = bodyBuilder(rawCode, expiresAt)
	}
	log.Printf("[OTP][DEBUG] sending otp email admin_id=%d purpose=%s email=%s", adminID, purpose, strings.TrimSpace(recipientEmail))
	if err := mailer.SendEmail(strings.TrimSpace(recipientEmail), subject, body); err != nil {
		log.Printf("[OTP][ERROR] deliver otp email failed admin_id=%d purpose=%s email=%s: %v", adminID, purpose, strings.TrimSpace(recipientEmail), err)
		return time.Time{}, fmt.Errorf("deliver otp email: %w", err)
	}

	log.Printf("[OTP][ISSUE] admin_id=%d purpose=%s email=%s expires_at=%s delivery=email status=sent code=%s", adminID, purpose, strings.TrimSpace(recipientEmail), expiresAt.Format(time.RFC3339), rawCode)

	return expiresAt, nil
}

// VerifyAndConsumeOTP validates the latest active challenge for a given admin and purpose.
func VerifyAndConsumeOTP(ctx context.Context, adminID int64, code string, purpose OTPPurpose) (OTPMetadata, error) {
	if adminID <= 0 {
		return nil, fmt.Errorf("adminID must be positive")
	}
	if strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("otp code is required")
	}

	opmMu.RLock()
	db := opmDB
	opmMu.RUnlock()
	if db == nil {
		return nil, fmt.Errorf("otp storage database is not configured")
	}

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, fmt.Errorf("begin otp transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var (
		otpID        int64
		storedHash   string
		expiresAt    time.Time
		metadataJSON string
	)
	err = tx.QueryRowContext(ctx, `
		SELECT id, otp_hash, expires_at, COALESCE(metadata::text, '{}')
		FROM otp_verifications
		WHERE admin_id = $1
		  AND used = FALSE
		  AND purpose = $2
		ORDER BY expires_at DESC, id DESC
		LIMIT 1
		FOR UPDATE
	`, adminID, string(purpose)).Scan(&otpID, &storedHash, &expiresAt, &metadataJSON)
	if err == sql.ErrNoRows {
		return nil, ErrOTPNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("load otp challenge: %w", err)
	}

	now := time.Now().UTC()
	if now.After(expiresAt) {
		return nil, ErrOTPExpired
	}

	if subtleCompareHash(storedHash, hashOTP(code)) == false {
		return nil, ErrOTPInvalid
	}

	if _, err := tx.ExecContext(ctx, `UPDATE otp_verifications SET used = TRUE WHERE id = $1`, otpID); err != nil {
		return nil, fmt.Errorf("mark otp as used: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit otp challenge: %w", err)
	}

	metadata := OTPMetadata{}
	if err := json.Unmarshal([]byte(metadataJSON), &metadata); err != nil {
		return nil, fmt.Errorf("decode otp metadata: %w", err)
	}

	return metadata, nil
}

func generateNumericOTP(length int) (string, error) {
	if length != 6 {
		length = 6
	}
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func hashOTP(code string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(code)))
	return hex.EncodeToString(sum[:])
}

func subtleCompareHash(left string, right string) bool {
	return strings.EqualFold(strings.TrimSpace(left), strings.TrimSpace(right))
}

func buildOTPEmailBody(code string, expiresAt time.Time) string {
	return fmt.Sprintf(`UrbanMemory Security Verification

Your one-time verification code is: %s

This code expires at %s UTC.
Do not share this code with anyone.

If you did not request this code, please contact a system administrator immediately.
`,
		code,
		expiresAt.Format(time.RFC3339),
	)
}
