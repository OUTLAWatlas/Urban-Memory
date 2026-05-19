package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
)

func TestLayersTrustMetadataShape(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	oldNewSvc := newBlockchainServiceFn
	oldVerify := verifyHashOnChainFn
	defer func() {
		newBlockchainServiceFn = oldNewSvc
		verifyHashOnChainFn = oldVerify
		appDB = nil
	}()
	appDB = db

	newBlockchainServiceFn = func(ctx context.Context) (*BlockchainService, error) {
		return &BlockchainService{}, nil
	}
	verifyHashOnChainFn = func(svc *BlockchainService, ctx context.Context, layerType string, year uint16, payload []byte) (HashVerificationResult, error) {
		expected := GenerateSHA256Hash(payload)
		return HashVerificationResult{
			LayerType:    layerType,
			Year:         year,
			OnChainHash:  expected,
			ExpectedHash: expected,
			Match:        true,
		}, nil
	}

	rows := sqlmock.NewRows([]string{"id", "city_name", "layer_type", "valid_from", "valid_to", "source_ref", "geojson"}).
		AddRow(1, "Mumbai", "slum_boundary", time.Date(2012, time.January, 1, 0, 0, 0, 0, time.UTC), nil, "BMC", `{"type":"Polygon","coordinates":[]}`)
	mock.ExpectQuery("SELECT(.|\\n)*FROM urban_artifacts").WithArgs("mumbai", "slum_boundary", 2012).WillReturnRows(rows)

	app := fiber.New()
	app.Get("/api/v1/:city/layers", GetVerifiedLayer())

	req := httptest.NewRequest("GET", "/api/v1/mumbai/layers?year=2012&layer_type=slum_boundary", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app test: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 got %d", resp.StatusCode)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	dataObj, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("missing data object")
	}
	if _, ok := dataObj["features"].([]any); !ok {
		t.Fatalf("missing data.features")
	}

	verification, ok := body["verification"].(map[string]any)
	if !ok {
		t.Fatalf("missing verification object")
	}
	if verification["is_verified"] != true {
		t.Fatalf("expected verification.is_verified true got %v", verification["is_verified"])
	}
	if verification["on_chain_hash"] == "" {
		t.Fatalf("expected verification.on_chain_hash to be populated")
	}
	if verification["status_message"] != "Match Found" {
		t.Fatalf("expected verification.status_message Match Found got %v", verification["status_message"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestAdminAuthMiddlewareRejectsMissingKey(t *testing.T) {
	if err := os.Setenv("ADMIN_API_KEY", "secret"); err != nil {
		t.Fatalf("set env: %v", err)
	}
	defer os.Unsetenv("ADMIN_API_KEY")

	app := fiber.New()
	app.Post("/api/admin/seal-history", adminAuthRequired(), NotarizeLayer())

	req := httptest.NewRequest("POST", "/api/admin/seal-history", bytes.NewBufferString(`{"layer_type":"slum_boundary","year":2012}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app test: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401 got %d", resp.StatusCode)
	}
}

func TestAdminSealHistorySuccess(t *testing.T) {
	if err := os.Setenv("ADMIN_API_KEY", "secret"); err != nil {
		t.Fatalf("set env: %v", err)
	}
	defer os.Unsetenv("ADMIN_API_KEY")

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	oldNewSvc := newBlockchainServiceFn
	oldCommit := commitHashToLedgerFn
	defer func() {
		newBlockchainServiceFn = oldNewSvc
		commitHashToLedgerFn = oldCommit
		appDB = nil
	}()
	appDB = db

	rows := sqlmock.NewRows([]string{"id", "city_name", "layer_type", "valid_from", "valid_to", "source_ref", "geojson"}).
		AddRow(1, "Mumbai", "slum_boundary", time.Date(2012, time.January, 1, 0, 0, 0, 0, time.UTC), nil, "BMC", `{"type":"Polygon","coordinates":[]}`)
	mock.ExpectQuery("SELECT(.|\\n)*FROM urban_artifacts").WithArgs("Mumbai", "slum_boundary", 2012).WillReturnRows(rows)

	newBlockchainServiceFn = func(ctx context.Context) (*BlockchainService, error) {
		return &BlockchainService{}, nil
	}
	commitHashToLedgerFn = func(svc *BlockchainService, ctx context.Context, layerType string, year uint16, sha256Hash string, ipfsCID string) (common.Hash, error) {
		return common.HexToHash("0x5678"), nil
	}

	app := fiber.New()
	app.Post("/api/admin/seal-history", adminAuthRequired(), NotarizeLayer())

	req := httptest.NewRequest("POST", "/api/admin/seal-history", bytes.NewBufferString(`{"layer_type":"slum_boundary","year":2012,"city":"Mumbai"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Key", "secret")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app test: %v", err)
	}
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("expected 201 got %d", resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestAdminSealHistoryNoData(t *testing.T) {
	if err := os.Setenv("ADMIN_API_KEY", "secret"); err != nil {
		t.Fatalf("set env: %v", err)
	}
	defer os.Unsetenv("ADMIN_API_KEY")

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	oldNewSvc := newBlockchainServiceFn
	defer func() {
		newBlockchainServiceFn = oldNewSvc
		appDB = nil
	}()
	appDB = db

	rows := sqlmock.NewRows([]string{"id", "city_name", "layer_type", "valid_from", "valid_to", "source_ref", "geojson"})
	mock.ExpectQuery("SELECT(.|\\n)*FROM urban_artifacts").WithArgs("Mumbai", "slum_boundary", 2012).WillReturnRows(rows)

	newBlockchainServiceFn = func(ctx context.Context) (*BlockchainService, error) {
		return nil, errors.New("should not be called")
	}

	app := fiber.New()
	app.Post("/api/admin/seal-history", adminAuthRequired(), NotarizeLayer())

	req := httptest.NewRequest("POST", "/api/admin/seal-history", bytes.NewBufferString(`{"layer_type":"slum_boundary","year":2012,"city":"Mumbai"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Key", "secret")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app test: %v", err)
	}
	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected 404 got %d", resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestLedgerStatusUnavailable(t *testing.T) {
	oldNewSvc := newBlockchainServiceFn
	defer func() {
		newBlockchainServiceFn = oldNewSvc
	}()

	newBlockchainServiceFn = func(ctx context.Context) (*BlockchainService, error) {
		return nil, errors.New("dial tcp 127.0.0.1:8545: connect refused")
	}

	app := fiber.New()
	app.Get("/api/v1/ledger/status", handleLedgerStatus())

	req := httptest.NewRequest("GET", "/api/v1/ledger/status", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app test: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 got %d", resp.StatusCode)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["connected"] != false {
		t.Fatalf("expected connected false got %v", body["connected"])
	}
}
