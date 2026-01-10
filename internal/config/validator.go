package config

import (
	"fmt"
	"net/url"

	"github.com/your-org/llm-info/pkg/config"
)

// ValidateConfig は設定値を検証する
func ValidateConfig(cfg *config.Config) error {
	if len(cfg.Gateways) == 0 {
		return fmt.Errorf("at least one gateway must be configured")
	}

	// ゲートウェイ名の重複チェック
	names := make(map[string]bool)
	for _, gw := range cfg.Gateways {
		if names[gw.Name] {
			return fmt.Errorf("duplicate gateway name: %s", gw.Name)
		}
		names[gw.Name] = true

		if err := validateGateway(&gw); err != nil {
			return fmt.Errorf("gateway %s: %w", gw.Name, err)
		}
	}

	// デフォルトゲートウェイの存在チェック
	if cfg.DefaultGateway != "" {
		if !names[cfg.DefaultGateway] {
			return fmt.Errorf("default gateway '%s' not found", cfg.DefaultGateway)
		}
	}

	// グローバル設定の検証
	if err := validateGlobal(&cfg.Global); err != nil {
		return fmt.Errorf("global settings: %w", err)
	}

	return nil
}

// validateGateway は個別のゲートウェイ設定を検証する
func validateGateway(gw *config.Gateway) error {
	if gw.Name == "" {
		return fmt.Errorf("gateway name cannot be empty")
	}

	if gw.URL == "" {
		return fmt.Errorf("gateway URL cannot be empty")
	}

	if _, err := url.Parse(gw.URL); err != nil {
		return fmt.Errorf("invalid URL format")
	}

	if gw.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	return nil
}

// validateGlobal はグローバル設定を検証する
func validateGlobal(global *config.Global) error {
	if global.Timeout <= 0 {
		return fmt.Errorf("global timeout must be positive")
	}

	// 出力形式の妥当性チェック
	validFormats := []string{"table", "json"}
	isValidFormat := false
	for _, format := range validFormats {
		if global.OutputFormat == format {
			isValidFormat = true
			break
		}
	}
	if !isValidFormat {
		return fmt.Errorf("invalid output format: %s (valid formats: %v)", global.OutputFormat, validFormats)
	}

	// ソート項目の妥当性チェック
	validSortBy := []string{"name", "max_tokens", "mode", "input_cost"}
	isValidSortBy := false
	for _, sortBy := range validSortBy {
		if global.SortBy == sortBy {
			isValidSortBy = true
			break
		}
	}
	if !isValidSortBy {
		return fmt.Errorf("invalid sort by: %s (valid options: %v)", global.SortBy, validSortBy)
	}

	return nil
}

// ValidateLegacyConfig は古い形式の設定値を検証する（後方互換性）
func ValidateLegacyConfig(fileConfig *config.FileConfig) error {
	if len(fileConfig.Gateways) == 0 {
		return fmt.Errorf("at least one gateway must be configured")
	}

	// ゲートウェイ名の重複チェック
	names := make(map[string]bool)
	for _, gw := range fileConfig.Gateways {
		if names[gw.Name] {
			return fmt.Errorf("duplicate gateway name: %s", gw.Name)
		}
		names[gw.Name] = true

		if err := validateLegacyGateway(&gw); err != nil {
			return fmt.Errorf("gateway %s: %w", gw.Name, err)
		}
	}

	// デフォルトゲートウェイの存在チェック
	if fileConfig.DefaultGateway != "" {
		if !names[fileConfig.DefaultGateway] {
			return fmt.Errorf("default gateway '%s' not found", fileConfig.DefaultGateway)
		}
	}

	// 共通設定の検証
	if err := validateCommon(&fileConfig.Common); err != nil {
		return fmt.Errorf("common settings: %w", err)
	}

	return nil
}

// validateLegacyGateway は古い形式の個別ゲートウェイ設定を検証する
func validateLegacyGateway(gw *config.GatewayConfig) error {
	if gw.Name == "" {
		return fmt.Errorf("gateway name cannot be empty")
	}

	if gw.URL == "" {
		return fmt.Errorf("gateway URL cannot be empty")
	}

	if _, err := url.Parse(gw.URL); err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if gw.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	return nil
}

// validateCommon は共通設定を検証する
func validateCommon(common *config.CommonConfig) error {
	if common.Timeout <= 0 {
		return fmt.Errorf("common timeout must be positive")
	}

	// 出力形式の妥当性チェック
	validFormats := []string{"table", "json"}
	isValidFormat := false
	for _, format := range validFormats {
		if common.Output.Format == format {
			isValidFormat = true
			break
		}
	}
	if !isValidFormat {
		return fmt.Errorf("invalid output format: %s (valid formats: %v)", common.Output.Format, validFormats)
	}

	return nil
}
