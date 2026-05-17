package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

const defaultPinataEndpoint = "https://api.pinata.cloud/pinning/pinFileToIPFS"

type pinataUploadResponse struct {
	IpfsHash string `json:"IpfsHash"`
	Error    string `json:"error"`
	Message  string `json:"message"`
}

// UploadToIPFS uploads raw GeoJSON bytes to Pinata and returns the resulting CID.
func UploadToIPFS(jsonData []byte) (string, error) {
	if len(jsonData) == 0 {
		return "", fmt.Errorf("jsonData is empty")
	}

	jwt := strings.TrimSpace(os.Getenv("PINATA_JWT"))
	if jwt == "" {
		return "", fmt.Errorf("PINATA_JWT is not set")
	}

	endpoint := strings.TrimSpace(os.Getenv("PINATA_ENDPOINT"))
	if endpoint == "" {
		endpoint = defaultPinataEndpoint
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	filePart, err := writer.CreateFormFile("file", "layer.geojson")
	if err != nil {
		return "", fmt.Errorf("create multipart file field: %w", err)
	}
	if _, err := filePart.Write(jsonData); err != nil {
		return "", fmt.Errorf("write multipart payload: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("finalize multipart payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, body)
	if err != nil {
		return "", fmt.Errorf("build pinata request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 45 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload to pinata: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read pinata response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("pinata returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var parsed pinataUploadResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("parse pinata response: %w", err)
	}
	if strings.TrimSpace(parsed.IpfsHash) == "" {
		if parsed.Error != "" {
			return "", fmt.Errorf("pinata upload failed: %s", parsed.Error)
		}
		if parsed.Message != "" {
			return "", fmt.Errorf("pinata upload failed: %s", parsed.Message)
		}
		return "", fmt.Errorf("pinata upload failed: missing IpfsHash")
	}

	return strings.TrimSpace(parsed.IpfsHash), nil
}
