package spreadsheet

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"log/slog"

	"github.com/Ssnakerss/modmas/internal/middleware"
	"github.com/Ssnakerss/modmas/internal/types"
	"github.com/go-chi/chi/v5"
)

// TestIsValidFieldType tests the isValidFieldType helper function
func TestIsValidFieldType(t *testing.T) {
	tests := []struct {
		fieldType string
		expected  bool
	}{
		{"text", true},
		{"integer", true},
		{"decimal", true},
		{"boolean", true},
		{"date", true},
		{"datetime", true},
		{"select", true},
		{"multi_select", true},
		{"email", true},
		{"url", true},
		{"phone", true},
		{"attachment", true},
		{"invalid", false},
		{"", false},
		{"TEXT", false},
	}

	for _, tt := range tests {
		t.Run(tt.fieldType, func(t *testing.T) {
			result := isValidFieldType(tt.fieldType)
			if result != tt.expected {
				t.Errorf("isValidFieldType(%q) = %v, expected %v", tt.fieldType, result, tt.expected)
			}
		})
	}
}

// TestValidateFieldOptions tests the validateFieldOptions helper function
func TestValidateFieldOptions(t *testing.T) {
	tests := []struct {
		name      string
		fieldType string
		options   map[string]interface{}
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid select field with choices",
			fieldType: "select",
			options: map[string]interface{}{
				"choices": []interface{}{
					map[string]interface{}{"value": "opt1", "label": "Option 1"},
					map[string]interface{}{"value": "opt2", "label": "Option 2"},
				},
			},
			expectErr: false,
		},
		{
			name:      "valid multi_select field with choices",
			fieldType: "multi_select",
			options: map[string]interface{}{
				"choices": []interface{}{
					map[string]interface{}{"value": "opt1", "label": "Option 1"},
				},
			},
			expectErr: false,
		},
		{
			name:      "select field missing choices",
			fieldType: "select",
			options:   map[string]interface{}{},
			expectErr: true,
			errMsg:    "select field must have 'choices' option",
		},
		{
			name:      "choices not an array",
			fieldType: "select",
			options: map[string]interface{}{
				"choices": "not an array",
			},
			expectErr: true,
			errMsg:    "'choices' must be a non-empty array",
		},
		{
			name:      "empty choices array",
			fieldType: "select",
			options: map[string]interface{}{
				"choices": []interface{}{},
			},
			expectErr: true,
			errMsg:    "'choices' must be a non-empty array",
		},
		{
			name:      "choice not an object",
			fieldType: "select",
			options: map[string]interface{}{
				"choices": []interface{}{"not an object"},
			},
			expectErr: true,
			errMsg:    "choice[0] must be an object",
		},
		{
			name:      "choice missing value",
			fieldType: "select",
			options: map[string]interface{}{
				"choices": []interface{}{
					map[string]interface{}{"label": "Option 1"},
				},
			},
			expectErr: true,
			errMsg:    "choice[0] must have a 'value' string field",
		},
		{
			name:      "choice missing label",
			fieldType: "select",
			options: map[string]interface{}{
				"choices": []interface{}{
					map[string]interface{}{"value": "opt1"},
				},
			},
			expectErr: true,
			errMsg:    "choice[0] must have a 'label' string field",
		},
		{
			name:      "non-select field with options (should pass)",
			fieldType: "text",
			options: map[string]interface{}{
				"some_option": "value",
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFieldOptions(tt.fieldType, tt.options)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if err.Error() != tt.errMsg {
					t.Errorf("Expected error message %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

// TestHandler_Create_Validation tests validation paths in Create handler
func TestHandler_Create_Validation(t *testing.T) {
	// Create a minimal handler with nil service (validation happens before service calls)
	handler := &Handler{
		service: nil,
		logger:  slog.Default(),
	}

	tests := []struct {
		name           string
		requestBody    interface{}
		userID         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "missing name",
			requestBody: map[string]interface{}{
				"workspace_id": "workspace-123",
				"fields":       []interface{}{},
			},
			userID:         "user-123",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"message":"name is required"`,
		},
		{
			name: "missing workspace_id",
			requestBody: map[string]interface{}{
				"name":   "Test Spreadsheet",
				"fields": []interface{}{},
			},
			userID:         "user-123",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"message":"workspace_id is required"`,
		},
		{
			name: "field missing name",
			requestBody: types.CreateSpreadsheetInput{
				WorkspaceID: "workspace-123",
				Name:        "Test Spreadsheet",
				Fields: []types.CreateFieldInput{
					{
						Name:      "",
						FieldType: "text",
					},
				},
			},
			userID:         "user-123",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"message":"field[0]: name is required"`,
		},
		{
			name: "invalid field type",
			requestBody: types.CreateSpreadsheetInput{
				WorkspaceID: "workspace-123",
				Name:        "Test Spreadsheet",
				Fields: []types.CreateFieldInput{
					{
						Name:      "Field 1",
						FieldType: "invalid_type",
					},
				},
			},
			userID:         "user-123",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"message":"field[0]: invalid field type 'invalid_type'"`,
		},
		{
			name:           "invalid JSON body",
			requestBody:    "{invalid json",
			userID:         "user-123",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"message":"invalid request body"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody []byte
			switch v := tt.requestBody.(type) {
			case string:
				reqBody = []byte(v)
			default:
				reqBody, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(http.MethodPost, "/spreadsheets", bytes.NewReader(reqBody))
			// Set user ID in context
			ctx := req.Context()
			ctx = context.WithValue(ctx, middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			// This will panic when trying to call service, but validation should return first
			// We need to catch panic for tests that pass validation
			defer func() {
				if r := recover(); r != nil {
					// Expected panic for tests that pass validation
					if tt.expectedStatus == http.StatusCreated {
						t.Errorf("Handler panicked: %v", r)
					}
				}
			}()

			handler.Create(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedBody != "" && !bytes.Contains(rr.Body.Bytes(), []byte(tt.expectedBody)) {
				t.Errorf("Expected body to contain %q, got %q", tt.expectedBody, rr.Body.String())
			}
		})
	}
}

// TestHandler_Update_Validation tests validation paths in Update handler
func TestHandler_Update_Validation(t *testing.T) {
	handler := &Handler{
		service: nil,
		logger:  slog.Default(),
	}

	tests := []struct {
		name           string
		spreadsheetID  string
		requestBody    interface{}
		expectedStatus int
		expectedBody   string
	}{
		{
			name:          "missing name",
			spreadsheetID: "spreadsheet-123",
			requestBody: map[string]interface{}{
				"description": "Updated Description",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"message":"name is required"`,
		},
		{
			name:           "invalid JSON",
			spreadsheetID:  "spreadsheet-123",
			requestBody:    "{invalid json",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"message":"invalid request body"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody []byte
			switch v := tt.requestBody.(type) {
			case string:
				reqBody = []byte(v)
			default:
				reqBody, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(http.MethodPut, "/spreadsheets/"+tt.spreadsheetID, bytes.NewReader(reqBody))

			// Setup chi router context with URL param
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.spreadsheetID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r != nil {
					// Expected panic for tests that pass validation
					if tt.expectedStatus == http.StatusOK {
						t.Errorf("Handler panicked: %v", r)
					}
				}
			}()

			handler.Update(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedBody != "" && !bytes.Contains(rr.Body.Bytes(), []byte(tt.expectedBody)) {
				t.Errorf("Expected body to contain %q, got %q", tt.expectedBody, rr.Body.String())
			}
		})
	}
}

// TestNewHandler tests the NewHandler constructor
func TestNewHandler(t *testing.T) {
	service := &Service{}
	logger := slog.Default()

	handler := NewHandler(service, logger)

	if handler == nil {
		t.Error("Expected handler to be created, got nil")
	}
	if handler.service != service {
		t.Error("Handler service not set correctly")
	}
	if handler.logger != logger {
		t.Error("Handler logger not set correctly")
	}
}

// Note: Full integration tests would require creating a real Service with all its dependencies.
// This is beyond the scope of unit tests for the handler alone.
// For comprehensive testing, consider:
// 1. Creating integration tests with a test database
// 2. Using interfaces for dependencies to enable mocking
// 3. Testing the service layer separately from the handler layer
