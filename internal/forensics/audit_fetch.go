package forensics

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go-antinuke-2.0/internal/config"
)

type AuditLogFetcher struct {
	token      string
	httpClient *http.Client
}

func NewAuditLogFetcher() *AuditLogFetcher {
	cfg := config.Get()
	return &AuditLogFetcher{
		token: cfg.Bot.Token,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (alf *AuditLogFetcher) FetchRecent(guildID uint64, limit int) ([]AuditLogEntry, error) {
	url := fmt.Sprintf("https://discord.com/api/v10/guilds/%d/audit-logs?limit=%d", guildID, limit)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bot %s", alf.token))

	resp, err := alf.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch failed: %d", resp.StatusCode)
	}

	var result AuditLogResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.AuditLogEntries, nil
}

func (alf *AuditLogFetcher) FetchByAction(guildID uint64, actionType int, limit int) ([]AuditLogEntry, error) {
	url := fmt.Sprintf("https://discord.com/api/v10/guilds/%d/audit-logs?action_type=%d&limit=%d", guildID, actionType, limit)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bot %s", alf.token))

	resp, err := alf.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch failed: %d", resp.StatusCode)
	}

	var result AuditLogResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.AuditLogEntries, nil
}

type AuditLogResponse struct {
	AuditLogEntries []AuditLogEntry `json:"audit_log_entries"`
}

type AuditLogEntry struct {
	ID         string `json:"id"`
	ActionType int    `json:"action_type"`
	UserID     string `json:"user_id"`
	TargetID   string `json:"target_id"`
	Reason     string `json:"reason"`
}
