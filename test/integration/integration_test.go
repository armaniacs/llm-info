package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/your-org/llm-info/internal/api"
	"github.com/your-org/llm-info/internal/config"
	"github.com/your-org/llm-info/internal/model"
	"github.com/your-org/llm-info/internal/ui"
)

func TestEndToEndFlow(t *testing.T) {
	// モックAPIサーバーのセットアップ
	mockResponse := api.ModelInfoResponse{
		Models: []api.ModelInfo{
			{
				ID:        "gpt-4",
				MaxTokens: 8192,
				Mode:      "chat",
				InputCost: 0.00003,
			},
			{
				ID:        "claude-3-opus",
				MaxTokens: 200000,
				Mode:      "chat",
				InputCost: 0.000015,
			},
			{
				ID:        "gemini-1.5-pro",
				MaxTokens: 1000000,
				Mode:      "chat",
				InputCost: 0,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// エンドポイントの検証
		if r.URL.Path != "/model/info" {
			t.Errorf("Expected path /model/info, got %s", r.URL.Path)
		}

		// HTTPメソッドの検証
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		// Content-Typeヘッダーの検証
		expectedContentType := "application/json"
		if contentType := r.Header.Get("Content-Type"); contentType != expectedContentType {
			t.Errorf("Expected Content-Type header %q, got %q", expectedContentType, contentType)
		}

		// レスポンスの設定
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(mockResponse); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// 設定の作成
	cfg := config.New(server.URL, "test-api-key", 5*time.Second)

	// APIクライアントの作成
	client := api.NewClient(cfg)

	// API呼び出し
	response, err := client.GetModelInfo()
	if err != nil {
		t.Fatalf("Failed to get model info: %v", err)
	}

	// レスポンスの検証
	if len(response.Models) != 3 {
		t.Errorf("Expected 3 models, got %d", len(response.Models))
	}

	// APIレスポンスをアプリケーションモデルに変換
	models := model.FromAPIResponse(response.Models)

	// 変換の検証
	if len(models) != 3 {
		t.Errorf("Expected 3 converted models, got %d", len(models))
	}

	// 最初のモデルの検証
	expectedFirstModel := model.Model{
		Name:      "gpt-4",
		MaxTokens: 8192,
		Mode:      "chat",
		InputCost: 0.00003,
	}

	if models[0].Name != expectedFirstModel.Name {
		t.Errorf("First model name = %q, expected %q", models[0].Name, expectedFirstModel.Name)
	}
	if models[0].MaxTokens != expectedFirstModel.MaxTokens {
		t.Errorf("First model max tokens = %d, expected %d", models[0].MaxTokens, expectedFirstModel.MaxTokens)
	}
	if models[0].Mode != expectedFirstModel.Mode {
		t.Errorf("First model mode = %q, expected %q", models[0].Mode, expectedFirstModel.Mode)
	}
	if models[0].InputCost != expectedFirstModel.InputCost {
		t.Errorf("First model input cost = %f, expected %f", models[0].InputCost, expectedFirstModel.InputCost)
	}

	// UI表示のテスト（出力の検証は難しいため、エラーが出ないことを確認）
	// 実際のE2Eテストでは出力をキャプチャして検証することも可能
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("RenderTable panicked: %v", r)
		}
	}()
	ui.RenderTable(models)
}

func TestEndToEndFlowWithEmptyResponse(t *testing.T) {
	// 空のレスポンスを返すモックサーバー
	mockResponse := api.ModelInfoResponse{
		Models: []api.ModelInfo{},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// 設定とクライアントの作成
	cfg := config.New(server.URL, "", 5*time.Second)
	client := api.NewClient(cfg)

	// API呼び出し
	response, err := client.GetModelInfo()
	if err != nil {
		t.Fatalf("Failed to get model info: %v", err)
	}

	// 空のレスポンスの検証
	if len(response.Models) != 0 {
		t.Errorf("Expected 0 models, got %d", len(response.Models))
	}

	// 変換と表示
	models := model.FromAPIResponse(response.Models)
	if len(models) != 0 {
		t.Errorf("Expected 0 converted models, got %d", len(models))
	}

	// 空のモデルリストの表示
	ui.RenderTable(models)
}

func TestEndToEndFlowWithError(t *testing.T) {
	// エラーを返すモックサーバー
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal server error"}`))
	}))
	defer server.Close()

	// 設定とクライアントの作成
	cfg := config.New(server.URL, "", 5*time.Second)
	client := api.NewClient(cfg)

	// API呼び出し（エラーが発生することを期待）
	_, err := client.GetModelInfo()
	if err == nil {
		t.Error("Expected error from API call, got nil")
	}

	// エラーメッセージの検証
	expectedError := "API request failed with status 500: {\"error\": \"Internal server error\"}"
	if err.Error() != expectedError {
		t.Errorf("Expected error message %q, got %q", expectedError, err.Error())
	}
}

func TestEndToEndFlowWithAuthentication(t *testing.T) {
	// 認証を要求するモックサーバー
	mockResponse := api.ModelInfoResponse{
		Models: []api.ModelInfo{
			{
				ID:        "authenticated-model",
				MaxTokens: 4096,
				Mode:      "chat",
				InputCost: 0.00001,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 認証ヘッダーの検証
		authHeader := r.Header.Get("Authorization")
		expectedAuth := "Bearer test-api-key"
		if authHeader != expectedAuth {
			t.Errorf("Expected Authorization header %q, got %q", expectedAuth, authHeader)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Unauthorized"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// APIキー付きの設定
	cfg := config.New(server.URL, "test-api-key", 5*time.Second)
	client := api.NewClient(cfg)

	// API呼び出し
	response, err := client.GetModelInfo()
	if err != nil {
		t.Fatalf("Failed to get model info: %v", err)
	}

	// レスポンスの検証
	if len(response.Models) != 1 {
		t.Errorf("Expected 1 model, got %d", len(response.Models))
	}

	// 変換と表示
	models := model.FromAPIResponse(response.Models)
	if len(models) != 1 {
		t.Errorf("Expected 1 converted model, got %d", len(models))
	}

	if models[0].Name != "authenticated-model" {
		t.Errorf("Expected model name 'authenticated-model', got %q", models[0].Name)
	}

	ui.RenderTable(models)
}
