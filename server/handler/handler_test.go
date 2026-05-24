package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/example/qianchuan-saas/auth"
	"github.com/example/qianchuan-saas/db"
	"github.com/example/qianchuan-saas/models"
	"github.com/example/qianchuan-saas/router"
)

func setupTestRouter() *gin.Engine {
	auth.InitJWT("test-secret")
	db.Connect("postgres://qianchuan:qianchuan_dev@localhost:5432/qianchuan_test?sslmode=disable")
	db.AutoMigrate(&models.User{}, &models.QianchuanAccount{}, &models.UniAd{})
	return router.Setup(nil)
}

func TestRegisterAndLogin(t *testing.T) {
	r := setupTestRouter()

	// Clean up any existing test user from a previous run
	db.DB.Where("email = ?", "test@example.com").Delete(&models.User{})

	body := map[string]string{
		"name": "测试用户", "email": "test@example.com", "password": "test123456",
	}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("register failed: code=%d body=%s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	loginBody := map[string]string{"email": "test@example.com", "password": "test123456"}
	b, _ = json.Marshal(loginBody)
	req, _ = http.NewRequest("POST", "/api/login", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("login failed: code=%d body=%s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["token"] == nil {
		t.Fatal("login response missing token")
	}
	t.Logf("login OK, token=%s", resp["token"])

	// Clean up test data
	db.DB.Where("email = ?", "test@example.com").Delete(&models.User{})
}
