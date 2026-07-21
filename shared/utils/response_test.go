package utils_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"banking-app/shared/utils"

	"github.com/gofiber/fiber/v2"
)

func TestSuccessResponse(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return utils.SuccessResponse(c, fiber.StatusOK, "Success message", fiber.Map{"key": "value"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test endpoint: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var res utils.Response
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !res.Success {
		t.Errorf("Expected Success to be true, got false")
	}
	if res.Message != "Success message" {
		t.Errorf("Expected message 'Success message', got '%s'", res.Message)
	}
}

func TestErrorResponse(t *testing.T) {
	app := fiber.New()
	app.Get("/test-err", func(c *fiber.Ctx) error {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Error message", "details")
	})

	req := httptest.NewRequest("GET", "/test-err", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test endpoint: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}

	var res utils.Response
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if res.Success {
		t.Errorf("Expected Success to be false, got true")
	}
	if res.Message != "Error message" {
		t.Errorf("Expected message 'Error message', got '%s'", res.Message)
	}
}
