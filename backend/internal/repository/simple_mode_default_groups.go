package repository

import (
	"context"
	"fmt"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/group"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type defaultGroupSpec struct {
	name                        string
	platform                    string
	supportedModelScopes        []string
	allowMessagesDispatch       bool
	messagesDispatchModelConfig service.OpenAIMessagesDispatchModelConfig
}

var defaultGroupSpecs = []defaultGroupSpec{
	{
		name:                  "GPT 系列-余额",
		platform:              service.PlatformOpenAI,
		supportedModelScopes:  []string{"claude", "gemini_text", "gemini_image"},
		allowMessagesDispatch: true,
		messagesDispatchModelConfig: service.OpenAIMessagesDispatchModelConfig{
			OpusMappedModel:   "gpt-5.5",
			SonnetMappedModel: "gpt-5.3-codex",
			HaikuMappedModel:  "gpt-5.4-mini",
		},
	},
	{
		name:                 "Claude 系列-余额",
		platform:             service.PlatformAnthropic,
		supportedModelScopes: []string{"claude", "gemini_text", "gemini_image"},
	},
	{
		name:                 "Gemini 系列-余额",
		platform:             service.PlatformGemini,
		supportedModelScopes: []string{"claude", "gemini_text", "gemini_image"},
	},
	{
		name:                 "Antigravity 系列-余额",
		platform:             service.PlatformAntigravity,
		supportedModelScopes: []string{"gemini_text", "gemini_image", "claude"},
	},
}

func ensureDefaultGroups(ctx context.Context, client *dbent.Client) error {
	if client == nil {
		return fmt.Errorf("nil ent client")
	}

	for _, spec := range defaultGroupSpecs {
		if err := createDefaultGroupIfNotExists(ctx, client, spec); err != nil {
			return err
		}
	}

	return nil
}

func createDefaultGroupIfNotExists(ctx context.Context, client *dbent.Client, spec defaultGroupSpec) error {
	exists, err := client.Group.Query().
		Where(group.NameEQ(spec.name), group.DeletedAtIsNil()).
		Exist(ctx)
	if err != nil {
		return fmt.Errorf("check default group exists %s: %w", spec.name, err)
	}
	if exists {
		return nil
	}

	_, err = client.Group.Create().
		SetName(spec.name).
		SetDescription("").
		SetPlatform(spec.platform).
		SetStatus(service.StatusActive).
		SetSubscriptionType(service.SubscriptionTypeStandard).
		SetRateMultiplier(1.0).
		SetIsExclusive(false).
		SetSortOrder(0).
		SetDailyLimitUsd(0).
		SetWeeklyLimitUsd(0).
		SetMonthlyLimitUsd(0).
		SetDefaultValidityDays(0).
		SetMcpXMLInject(true).
		SetSupportedModelScopes(spec.supportedModelScopes).
		SetAllowMessagesDispatch(spec.allowMessagesDispatch).
		SetMessagesDispatchModelConfig(spec.messagesDispatchModelConfig).
		SetRpmLimit(0).
		Save(ctx)
	if err != nil {
		if dbent.IsConstraintError(err) {
			// Concurrent server startups may race on creation; treat as success.
			return nil
		}
		return fmt.Errorf("create default group %s: %w", spec.name, err)
	}
	return nil
}
