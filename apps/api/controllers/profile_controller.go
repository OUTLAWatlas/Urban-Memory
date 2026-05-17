package controllers

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/urban-memory/api/services"
	"golang.org/x/crypto/bcrypt"
)

type RequestPasswordChangeRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_new_password"`
}

type ConfirmPasswordChangeRequest struct {
	OTPCode string `json:"otp_code"`
}

type PendingAdminSummary struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Role      AdminRole `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

func RequestPasswordChange(db *sql.DB, mailer services.MailSender) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if db == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database is not initialized"})
		}
		if mailer == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "mail sender is not configured"})
		}

		admin, err := requireActiveAdmin(c.UserContext(), db, c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}

		var req RequestPasswordChangeRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
		}

		currentPassword := strings.TrimSpace(req.CurrentPassword)
		newPassword := strings.TrimSpace(req.NewPassword)
		confirmPassword := strings.TrimSpace(req.ConfirmPassword)

		if currentPassword == "" || newPassword == "" || confirmPassword == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "All password fields are required."})
		}
		if newPassword != confirmPassword {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "New password confirmation does not match."})
		}
		if len(newPassword) < 8 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "New password must be at least 8 characters long."})
		}
		if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(currentPassword)); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Mismatched current password"})
		}

		pendingPasswordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to hash pending password"})
		}

		expiresAt, err := services.IssueOTPChallenge(
			c.UserContext(),
			admin.ID,
			admin.Email,
			services.OTPPurposePasswordChange,
			services.OTPMetadata{"pending_password_hash": string(pendingPasswordHash)},
			mailer,
			"HIGH PRIORITY: UrbanMemory Password Rotation OTP",
			func(code string, otpExpiresAt time.Time) string {
				return fmt.Sprintf(`UrbanMemory Password Rotation

Your account password rotation request is pending.

Verification code: %s
Expires at: %s UTC

If you did not initiate this request, terminate your session immediately.
`, code, otpExpiresAt.Format(time.RFC3339))
			},
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to issue password rotation challenge", "details": err.Error()})
		}

		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
			"message":    "Password rotation OTP sent to your registered inbox.",
			"expires_at": expiresAt.Format(time.RFC3339),
		})
	}
}

func ConfirmPasswordChange(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if db == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database is not initialized"})
		}

		admin, err := requireActiveAdmin(c.UserContext(), db, c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}

		var req ConfirmPasswordChangeRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
		}

		otpCode := strings.TrimSpace(req.OTPCode)
		if len(otpCode) != 6 || strings.Trim(otpCode, "0123456789") != "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "otp_code must be a 6-digit numeric string"})
		}

		metadata, err := services.VerifyAndConsumeOTP(c.UserContext(), admin.ID, otpCode, services.OTPPurposePasswordChange)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": describeOTPError(err)})
		}

		pendingPasswordHash := strings.TrimSpace(metadata["pending_password_hash"])
		if pendingPasswordHash == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "verification challenge missing pending password state"})
		}

		if _, err := db.ExecContext(c.UserContext(), `
			UPDATE admins
			SET password_hash = $2
			WHERE id = $1
		`, admin.ID, pendingPasswordHash); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to commit password rotation", "details": err.Error()})
		}

		return c.JSON(fiber.Map{
			"message": "Password updated successfully.",
		})
	}
}

func ListPendingAdmins(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if db == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database is not initialized"})
		}

		if _, err := requireSuperAdmin(c.UserContext(), db, c); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}

		rows, err := db.QueryContext(c.UserContext(), `
			SELECT id, email, role, created_at
			FROM admins
			WHERE role = $1
			ORDER BY created_at ASC, id ASC
		`, string(AdminRolePending))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load pending admins", "details": err.Error()})
		}
		defer rows.Close()

		pendingAdmins := make([]PendingAdminSummary, 0)
		for rows.Next() {
			var row PendingAdminSummary
			if err := rows.Scan(&row.ID, &row.Email, &row.Role, &row.CreatedAt); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to scan pending admin", "details": err.Error()})
			}
			pendingAdmins = append(pendingAdmins, row)
		}
		if err := rows.Err(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to iterate pending admins", "details": err.Error()})
		}

		return c.JSON(fiber.Map{
			"pending_admins": pendingAdmins,
		})
	}
}
