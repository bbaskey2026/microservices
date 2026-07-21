package health

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestLivenessHandler(t *testing.T) {
	app := fiber.New()
	RegisterHealthRoutes(app, "test-service", nil, nil)

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test health endpoint: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var healthResp HealthResponse
	if err := json.Unmarshal(body, &healthResp); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if !healthResp.Success || healthResp.Status != "UP" || healthResp.Service != "test-service" {
		t.Errorf("Unexpected health response: %+v", healthResp)
	}
}

func TestReadinessHandlerNoDeps(t *testing.T) {
	app := fiber.New()
	RegisterHealthRoutes(app, "test-service", nil, nil)

	req := httptest.NewRequest("GET", "/ready", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test readiness endpoint: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}
}
