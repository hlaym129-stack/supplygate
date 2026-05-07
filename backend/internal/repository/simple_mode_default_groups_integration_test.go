//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/ent/group"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestEnsureDefaultGroups_CreatesMissingDefaults(t *testing.T) {
	ctx := context.Background()
	tx := testEntTx(t)
	client := tx.Client()

	seedCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	require.NoError(t, ensureDefaultGroups(seedCtx, client))

	assertGroup := func(name, platform string, allowMessagesDispatch bool) {
		g, err := client.Group.Query().Where(group.NameEQ(name), group.DeletedAtIsNil()).Only(seedCtx)
		require.NoError(t, err)
		require.Equal(t, platform, g.Platform)
		require.Equal(t, service.StatusActive, g.Status)
		require.Equal(t, service.SubscriptionTypeStandard, g.SubscriptionType)
		require.Equal(t, 1.0, g.RateMultiplier)
		require.False(t, g.IsExclusive)
		require.Equal(t, 0, g.SortOrder)
		require.Equal(t, 0, g.DefaultValidityDays)
		require.NotNil(t, g.DailyLimitUsd)
		require.NotNil(t, g.WeeklyLimitUsd)
		require.NotNil(t, g.MonthlyLimitUsd)
		require.Equal(t, 0.0, *g.DailyLimitUsd)
		require.Equal(t, 0.0, *g.WeeklyLimitUsd)
		require.Equal(t, 0.0, *g.MonthlyLimitUsd)
		require.True(t, g.McpXMLInject)
		require.Equal(t, allowMessagesDispatch, g.AllowMessagesDispatch)
		require.Equal(t, 0, g.RpmLimit)
	}

	assertGroup("GPT 系列-余额", service.PlatformOpenAI, true)
	assertGroup("Claude 系列-余额", service.PlatformAnthropic, false)
	assertGroup("Gemini 系列-余额", service.PlatformGemini, false)
	assertGroup("Antigravity 系列-余额", service.PlatformAntigravity, false)

	openAIGroup, err := client.Group.Query().Where(group.NameEQ("GPT 系列-余额"), group.DeletedAtIsNil()).Only(seedCtx)
	require.NoError(t, err)
	require.Equal(t, service.OpenAIMessagesDispatchModelConfig{
		OpusMappedModel:   "gpt-5.5",
		SonnetMappedModel: "gpt-5.3-codex",
		HaikuMappedModel:  "gpt-5.4-mini",
	}, openAIGroup.MessagesDispatchModelConfig)
}

func TestEnsureDefaultGroups_IgnoresSoftDeletedGroups(t *testing.T) {
	ctx := context.Background()
	tx := testEntTx(t)
	client := tx.Client()

	seedCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	g, err := client.Group.Create().
		SetName("Claude 系列-余额").
		SetPlatform(service.PlatformAnthropic).
		SetStatus(service.StatusActive).
		SetSubscriptionType(service.SubscriptionTypeStandard).
		SetRateMultiplier(1.0).
		SetIsExclusive(false).
		Save(seedCtx)
	require.NoError(t, err)

	_, err = client.Group.Delete().Where(group.IDEQ(g.ID)).Exec(seedCtx)
	require.NoError(t, err)

	require.NoError(t, ensureDefaultGroups(seedCtx, client))

	count, err := client.Group.Query().Where(group.NameEQ("Claude 系列-余额"), group.DeletedAtIsNil()).Count(seedCtx)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestEnsureDefaultGroups_DoesNotOverwriteExistingDefaults(t *testing.T) {
	ctx := context.Background()
	tx := testEntTx(t)
	client := tx.Client()

	seedCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	g, err := client.Group.Create().
		SetName("GPT 系列-余额").
		SetDescription("custom").
		SetPlatform(service.PlatformOpenAI).
		SetStatus(service.StatusActive).
		SetSubscriptionType(service.SubscriptionTypeStandard).
		SetRateMultiplier(2.0).
		SetIsExclusive(true).
		SetAllowMessagesDispatch(false).
		Save(seedCtx)
	require.NoError(t, err)

	require.NoError(t, ensureDefaultGroups(seedCtx, client))

	got, err := client.Group.Query().Where(group.IDEQ(g.ID)).Only(seedCtx)
	require.NoError(t, err)
	require.NotNil(t, got.Description)
	require.Equal(t, "custom", *got.Description)
	require.Equal(t, 2.0, got.RateMultiplier)
	require.True(t, got.IsExclusive)
	require.False(t, got.AllowMessagesDispatch)
}
