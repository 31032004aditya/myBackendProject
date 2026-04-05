package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/internal/handler"
	"backend/internal/middleware"
	"backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func setupTestRouter() *gin.Engine {
	// Use Gin in test mode so we don't get verbose output
	gin.SetMode(gin.TestMode)

	r := gin.Default()

	// Instantiate fresh in-memory repos for each test run to ensure isolation
	userRepo := repository.NewUserRepository()
	recordRepo := repository.NewRecordRepository()

	authHandler := handler.NewAuthHandler(userRepo)
	recordHandler := handler.NewRecordHandler(recordRepo)

	// Routing
	r.POST("/api/auth/register", authHandler.Register)
	r.POST("/api/auth/login", authHandler.Login)

	api := r.Group("/api")
	api.Use(middleware.AuthRequired())
	{
		dashboard := api.Group("/dashboard")
		dashboard.Use(middleware.RoleRequired("analyst", "admin"))
		{
			dashboard.GET("/summary", recordHandler.GetSummary)
		}

		records := api.Group("/records")
		{
			records.POST("", middleware.RoleRequired("admin"), recordHandler.Create)
		}
	}
	return r
}

func TestCoreProjectRequirements(t *testing.T) {
	router := setupTestRouter()

	var adminToken string

	t.Run("1. Register User (First user gets Admin)", func(t *testing.T) {
		body := []byte(`{"username":"admin_test","password":"Password123"}`)
		req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("FAIL: Expected 201 Created for registration, got %d. Body: %s", w.Code, w.Body.String())
		}
		
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		if response["role"] != "admin" {
			t.Fatalf("FAIL: Expected first registered user to be 'admin', got '%v'", response["role"])
		}
		t.Log("PASS: Registered admin user securely")
	})

	t.Run("2. Login and get JWT Token", func(t *testing.T) {
		body := []byte(`{"username":"admin_test","password":"Password123"}`)
		req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("FAIL: Expected 200 OK for login, got %d", w.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		
		token, exists := response["token"].(string)
		if !exists || token == "" {
			t.Fatalf("FAIL: Expected JWT token in response")
		}
		adminToken = token
		t.Log("PASS: User logged in and retrieved active JWT token")
	})

	t.Run("3. RBAC/Auth Enforcement (Reject un-authenticated)", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/dashboard/summary", nil)
		w := httptest.NewRecorder()
		// No Bearer token sent!
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("FAIL: Expected 401 Unauthorized for unprotected access, got %d", w.Code)
		}
		t.Log("PASS: Security middleware blocked unauthorized access")
	})

	t.Run("4. Create Financial Records (Admin Access)", func(t *testing.T) {
		// Income Entry: 2000
		income := []byte(`{"amount": 2000.0, "type": "income", "category": "Salary"}`)
		req1, _ := http.NewRequest("POST", "/api/records", bytes.NewBuffer(income))
		req1.Header.Set("Authorization", "Bearer "+adminToken)
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)

		if w1.Code != http.StatusCreated {
			t.Fatalf("FAIL: Expected 201 Created for income, got %d: %s", w1.Code, w1.Body.String())
		}

		// Expense Entry: 500
		expense := []byte(`{"amount": 500.0, "type": "expense", "category": "Food"}`)
		req2, _ := http.NewRequest("POST", "/api/records", bytes.NewBuffer(expense))
		req2.Header.Set("Authorization", "Bearer "+adminToken)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		if w2.Code != http.StatusCreated {
			t.Fatalf("FAIL: Expected 201 Created for expense, got %d", w2.Code)
		}
		
		t.Log("PASS: Processed active financial records securely")
	})

	t.Run("5. Dashboard Aggregation Validation", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/dashboard/summary", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		
		// Give the system a fraction of a millisecond just in case (though in-memory is instant)
		time.Sleep(10 * time.Millisecond)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("FAIL: Expected 200 OK for summary fetch, got %d", w.Code)
		}

		var summary map[string]float64
		json.Unmarshal(w.Body.Bytes(), &summary)

		if summary["netBalance"] != 1500.0 {
			t.Fatalf("FAIL: Expected net balance of 1500, got %v", summary["netBalance"])
		}
		if summary["totalIncome"] != 2000.0 {
			t.Fatalf("FAIL: Expected total income of 2000, got %v", summary["totalIncome"])
		}
		if summary["totalExpense"] != 500.0 {
			t.Fatalf("FAIL: Expected total expense of 500, got %v", summary["totalExpense"])
		}
		
		t.Log("PASS: Dashboard calculations returned absolutely accurately")
	})
}
