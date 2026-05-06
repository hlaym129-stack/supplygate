package service

import (
	"context"
	"testing"
	"time"
)

type updateCacheStub struct {
	data string
	err  error
}

func (s *updateCacheStub) GetUpdateInfo(_ context.Context) (string, error) {
	return s.data, s.err
}

func (s *updateCacheStub) SetUpdateInfo(_ context.Context, data string, _ time.Duration) error {
	s.data = data
	return nil
}

type githubReleaseClientStub struct {
	repo string
}

func (s *githubReleaseClientStub) FetchLatestRelease(_ context.Context, repo string) (*GitHubRelease, error) {
	s.repo = repo
	return &GitHubRelease{
		TagName:     "v1.2.4",
		Name:        "v1.2.4",
		PublishedAt: "2026-05-06T00:00:00Z",
		HTMLURL:     "https://github.com/" + repo + "/releases/tag/v1.2.4",
	}, nil
}

func (s *githubReleaseClientStub) DownloadFile(_ context.Context, _, _ string, _ int64) error {
	return nil
}

func (s *githubReleaseClientStub) FetchChecksumFile(_ context.Context, _ string) ([]byte, error) {
	return nil, nil
}

func TestUpdateService_DefaultGitHubRepo(t *testing.T) {
	client := &githubReleaseClientStub{}
	svc := NewUpdateService(&updateCacheStub{}, client, "1.2.3", "release", "")

	info, err := svc.CheckUpdate(context.Background(), true)
	if err != nil {
		t.Fatalf("CheckUpdate() error = %v", err)
	}
	if client.repo != defaultUpdateGitHubRepo {
		t.Fatalf("repo = %q, want %q", client.repo, defaultUpdateGitHubRepo)
	}
	if !info.HasUpdate {
		t.Fatalf("HasUpdate = false, want true")
	}
	if info.ReleaseInfo == nil || info.ReleaseInfo.HTMLURL != "https://github.com/hlaym129-stack/supplygate/releases/tag/v1.2.4" {
		t.Fatalf("unexpected release info: %+v", info.ReleaseInfo)
	}
}

func TestUpdateService_CustomGitHubRepo(t *testing.T) {
	client := &githubReleaseClientStub{}
	svc := NewUpdateService(&updateCacheStub{}, client, "1.2.3", "release", "owner/custom")

	if _, err := svc.CheckUpdate(context.Background(), true); err != nil {
		t.Fatalf("CheckUpdate() error = %v", err)
	}
	if client.repo != "owner/custom" {
		t.Fatalf("repo = %q, want owner/custom", client.repo)
	}
}

func TestUpdateService_IgnoresCacheFromDifferentGitHubRepo(t *testing.T) {
	cache := &updateCacheStub{
		data: `{"latest":"9.9.9","release_info":{"html_url":"https://github.com/Wei-Shaw/sub2api/releases/tag/v9.9.9"},"github_repo":"Wei-Shaw/sub2api","timestamp":1778000000}`,
	}
	client := &githubReleaseClientStub{}
	svc := NewUpdateService(cache, client, "1.2.3", "release", "")

	info, err := svc.CheckUpdate(context.Background(), false)
	if err != nil {
		t.Fatalf("CheckUpdate() error = %v", err)
	}
	if client.repo != defaultUpdateGitHubRepo {
		t.Fatalf("repo = %q, want %q", client.repo, defaultUpdateGitHubRepo)
	}
	if info.ReleaseInfo == nil || info.ReleaseInfo.HTMLURL != "https://github.com/hlaym129-stack/supplygate/releases/tag/v1.2.4" {
		t.Fatalf("unexpected release info: %+v", info.ReleaseInfo)
	}
}
