package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	defaultNodeURL         = "http://127.0.0.1:8545"
	defaultContractAddress = "0x5FbDB2315678afecb367f032d93F642f64180aa3"
	// Hardhat account #0 private key. Use ETH_PRIVATE_KEY in non-local environments.
	defaultHardhatPrivateKey = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
)

var ErrBlockchainDisabled = errors.New("blockchain integration disabled")

type BlockchainService struct {
	client          *ethclient.Client
	ledger          *UrbanLedger
	contractAddress common.Address
	privateKey      *ecdsa.PrivateKey
	chainID         *big.Int
	sourceRef       string
}

type HashVerificationResult struct {
	LayerType       string
	Year            uint16
	ExpectedHash    string
	OnChainHash     string
	OnChainIPFSCID  string
	OnChainSource   string
	OnChainUnixTime uint64
	Match           bool
}

func NewBlockchainService(ctx context.Context) (*BlockchainService, error) {
	if !isBlockchainEnabled() {
		return nil, ErrBlockchainDisabled
	}

	nodeURL := getEnv("ETH_NODE_URL", defaultNodeURL)
	contractAddrHex := getEnv("LEDGER_CONTRACT_ADDRESS", defaultContractAddress)
	sourceRef := getEnv("LEDGER_SOURCE_REF", "UrbanMemory API")

	if !common.IsHexAddress(contractAddrHex) {
		return nil, fmt.Errorf("invalid LEDGER_CONTRACT_ADDRESS: %s", contractAddrHex)
	}
	contractAddress := common.HexToAddress(contractAddrHex)

	client, err := ethclient.DialContext(ctx, nodeURL)
	if err != nil {
		return nil, fmt.Errorf("connect ethereum node: %w", err)
	}

	chainID, err := client.ChainID(ctx)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("read chain id: %w", err)
	}

	privHex := strings.TrimPrefix(getEnv("ETH_PRIVATE_KEY", defaultHardhatPrivateKey), "0x")
	privateKey, err := crypto.HexToECDSA(privHex)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("parse ETH_PRIVATE_KEY: %w", err)
	}

	ledger, err := NewUrbanLedger(contractAddress, client)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("bind UrbanLedger contract: %w", err)
	}

	return &BlockchainService{
		client:          client,
		ledger:          ledger,
		contractAddress: contractAddress,
		privateKey:      privateKey,
		chainID:         chainID,
		sourceRef:       sourceRef,
	}, nil
}

func (s *BlockchainService) Close() {
	if s.client != nil {
		s.client.Close()
	}
}

func GenerateSHA256Hash(geoJSONData []byte) string {
	h := sha256.Sum256(geoJSONData)
	return hex.EncodeToString(h[:])
}

func (s *BlockchainService) CommitHashToLedger(ctx context.Context, layerType string, year uint16, sha256Hash string, ipfsCID string) (common.Hash, error) {
	if s == nil || s.client == nil || s.ledger == nil {
		return common.Hash{}, errors.New("blockchain service is not initialized")
	}
	if strings.TrimSpace(layerType) == "" {
		return common.Hash{}, errors.New("layerType is required")
	}
	if strings.TrimSpace(sha256Hash) == "" {
		return common.Hash{}, errors.New("sha256Hash is required")
	}
	if strings.TrimSpace(ipfsCID) == "" {
		return common.Hash{}, errors.New("ipfsCID is required")
	}

	fromAddress := crypto.PubkeyToAddress(s.privateKey.PublicKey)
	nonce, err := s.client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return common.Hash{}, fmt.Errorf("read pending nonce: %w", err)
	}

	transactor, err := bind.NewKeyedTransactorWithChainID(s.privateKey, s.chainID)
	if err != nil {
		return common.Hash{}, fmt.Errorf("create keyed transactor: %w", err)
	}

	transactor.Context = ctx
	transactor.Nonce = big.NewInt(int64(nonce))

	if gasPrice, gpErr := s.client.SuggestGasPrice(ctx); gpErr == nil {
		transactor.GasPrice = gasPrice
	}

	tx, err := s.ledger.CommitLayerHash(transactor, layerType, year, sha256Hash, ipfsCID, s.sourceRef)
	if err != nil {
		return common.Hash{}, fmt.Errorf("commitLayerHash tx submission failed: %w", err)
	}

	receipt, err := bind.WaitMined(ctx, s.client, tx)
	if err != nil {
		return tx.Hash(), fmt.Errorf("wait for tx mined: %w", err)
	}
	if receipt == nil {
		return tx.Hash(), errors.New("nil transaction receipt")
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return tx.Hash(), fmt.Errorf("transaction reverted: tx=%s status=%d", tx.Hash().Hex(), receipt.Status)
	}

	return tx.Hash(), nil
}

func (s *BlockchainService) VerifyHashOnChain(ctx context.Context, layerType string, year uint16, currentGeoJSON []byte) (HashVerificationResult, error) {
	result := HashVerificationResult{LayerType: layerType, Year: year}

	if s == nil || s.ledger == nil {
		return result, errors.New("blockchain service is not initialized")
	}
	if strings.TrimSpace(layerType) == "" {
		return result, errors.New("layerType is required")
	}
	if len(currentGeoJSON) == 0 {
		return result, errors.New("currentGeoJSON is empty")
	}

	expectedHash := GenerateSHA256Hash(currentGeoJSON)
	result.ExpectedHash = expectedHash

	onChainHash, onChainTs, onChainSource, onChainCID, err := s.ledger.VerifyLayer(&bind.CallOpts{Context: ctx}, layerType, year)
	if err != nil {
		if isUnnotarizedLayerError(err) {
			log.Printf("⚠️ Blockchain Notice: Requested layer/year is currently unnotarized. Proceeding with pipeline.")
			result.Match = false
			return result, nil
		}
		return result, fmt.Errorf("verifyLayer call failed: %w", err)
	}

	result.OnChainHash = onChainHash
	result.OnChainSource = onChainSource
	result.OnChainIPFSCID = onChainCID
	if onChainTs != nil {
		result.OnChainUnixTime = onChainTs.Uint64()
	}

	result.Match = strings.EqualFold(strings.TrimSpace(onChainHash), strings.TrimSpace(expectedHash))
	return result, nil
}

func isUnnotarizedLayerError(err error) bool {
	if err == nil {
		return false
	}

	errText := err.Error()
	return strings.Contains(errText, "No cryptographic record found") ||
		strings.Contains(errText, "VM Exception while processing transaction: reverted")
}

func BuildLayerGeoJSONHash(ctx context.Context, db *sql.DB, cityName, layerType string, year uint16) (string, []byte, error) {
	if db == nil {
		return "", nil, errors.New("nil database handle")
	}

	query := `
		SELECT COALESCE(
			jsonb_agg(ST_AsGeoJSON(geom)::jsonb ORDER BY id)::text,
			'[]'
		)
		FROM urban_artifacts
		WHERE city_name ILIKE $1
		  AND layer_type = $2
		  AND valid_from <= make_date($3, 12, 31)
		  AND (valid_to IS NULL OR valid_to >= make_date($3, 1, 1));
	`

	var payload string
	if err := db.QueryRowContext(ctx, query, cityName, layerType, int(year)).Scan(&payload); err != nil {
		return "", nil, fmt.Errorf("query layer payload: %w", err)
	}

	data := []byte(payload)
	return GenerateSHA256Hash(data), data, nil
}

func parseYearToUint16(year int) (uint16, error) {
	if year < 0 || year > int(^uint16(0)) {
		return 0, fmt.Errorf("year out of uint16 range: %d", year)
	}
	return uint16(year), nil
}

func getUint16Env(key string, fallback uint16) (uint16, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseUint(raw, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}
	return uint16(parsed), nil
}

func isBlockchainEnabled() bool {
	raw := strings.TrimSpace(strings.ToLower(os.Getenv("BLOCKCHAIN_ENABLED")))
	switch raw {
	case "", "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return true
	}
}

func isBlockchainUnavailableError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrBlockchainDisabled) {
		return true
	}

	message := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(message, "connectex") ||
		strings.Contains(message, "connect refused") ||
		strings.Contains(message, "actively refused") ||
		strings.Contains(message, "read chain id") ||
		strings.Contains(message, "connect ethereum node")
}

func blockchainUnavailableMessage(err error) string {
	if errors.Is(err, ErrBlockchainDisabled) {
		return "blockchain integration disabled"
	}
	return "blockchain unavailable"
}
