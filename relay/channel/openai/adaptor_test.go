package openai

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
)

func TestNormalizeOpenRouterExtraBodyPromotesGoogleImageConfig(t *testing.T) {
	request := &dto.GeneralOpenAIRequest{
		Model: "google/gemini-2.5-flash-image-preview",
		ExtraBody: []byte(`{
			"google": {
				"image_config": {
					"aspect_ratio": "16:9",
					"image_size": "2K"
				}
			}
		}`),
	}

	if err := normalizeOpenRouterExtraBody(request); err != nil {
		t.Fatalf("normalizeOpenRouterExtraBody returned error: %v", err)
	}

	var normalized map[string]any
	if err := common.Unmarshal(request.ExtraBody, &normalized); err != nil {
		t.Fatalf("failed to parse normalized extra_body: %v", err)
	}

	if _, exists := normalized["google"]; exists {
		t.Fatalf("expected google key removed when it only contains image_config")
	}

	imageConfig, ok := normalized["image_config"].(map[string]any)
	if !ok {
		t.Fatalf("expected image_config object, got %T", normalized["image_config"])
	}
	if imageConfig["aspect_ratio"] != "16:9" {
		t.Fatalf("expected aspect_ratio=16:9, got %#v", imageConfig["aspect_ratio"])
	}
	if imageConfig["image_size"] != "2K" {
		t.Fatalf("expected image_size=2K, got %#v", imageConfig["image_size"])
	}
}

func TestNormalizeOpenRouterExtraBodyKeepsGoogleIfHasOtherFields(t *testing.T) {
	request := &dto.GeneralOpenAIRequest{
		ExtraBody: []byte(`{
			"google": {
				"image_config": {"aspect_ratio": "1:1"},
				"safety": {"level": "low"}
			}
		}`),
	}

	if err := normalizeOpenRouterExtraBody(request); err != nil {
		t.Fatalf("normalizeOpenRouterExtraBody returned error: %v", err)
	}

	var normalized map[string]any
	if err := common.Unmarshal(request.ExtraBody, &normalized); err != nil {
		t.Fatalf("failed to parse normalized extra_body: %v", err)
	}

	google, ok := normalized["google"].(map[string]any)
	if !ok {
		t.Fatalf("expected google object to remain")
	}
	if _, hasImageConfig := google["image_config"]; hasImageConfig {
		t.Fatalf("expected google.image_config removed after promotion")
	}
	if _, hasSafety := google["safety"]; !hasSafety {
		t.Fatalf("expected google.safety preserved")
	}
	if _, ok := normalized["image_config"]; !ok {
		t.Fatalf("expected top-level image_config present")
	}
}

func TestNormalizeOpenRouterExtraBodyInvalidJSON(t *testing.T) {
	request := &dto.GeneralOpenAIRequest{ExtraBody: []byte(`{"google":`)}
	if err := normalizeOpenRouterExtraBody(request); err == nil {
		t.Fatal("expected error for invalid extra_body json")
	}
}

func TestConvertOpenAIRequestOpenRouterMergesNormalizedExtraBodyToTopLevel(t *testing.T) {
	adaptor := &Adaptor{}
	request := &dto.GeneralOpenAIRequest{
		Model: "google/gemini-2.5-flash-image-preview",
		ExtraBody: []byte(`{
			"google": {
				"image_config": {
					"aspect_ratio": "16:9",
					"image_size": "2K"
				}
			}
		}`),
	}
	info := &relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{ChannelType: constant.ChannelTypeOpenRouter},
	}

	converted, err := adaptor.ConvertOpenAIRequest(nil, info, request)
	if err != nil {
		t.Fatalf("ConvertOpenAIRequest returned error: %v", err)
	}

	payload, ok := converted.(map[string]any)
	if !ok {
		t.Fatalf("expected merged payload map for openrouter, got %T", converted)
	}
	if _, exists := payload["extra_body"]; exists {
		t.Fatalf("expected extra_body removed after merge")
	}
	imageConfig, ok := payload["image_config"].(map[string]any)
	if !ok {
		t.Fatalf("expected top-level image_config, got %T", payload["image_config"])
	}
	if imageConfig["aspect_ratio"] != "16:9" || imageConfig["image_size"] != "2K" {
		t.Fatalf("unexpected image_config value: %#v", imageConfig)
	}
}
