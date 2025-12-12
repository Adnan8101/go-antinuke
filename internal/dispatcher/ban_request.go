package dispatcher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go-antinuke-2.0/internal/config"
	"go-antinuke-2.0/internal/database"
	"go-antinuke-2.0/internal/logging"
)

type BanRequestExecutor struct {
	httpPool    *HTTPPool
	rateLimiter *RateLimitMonitor
	token       string
}

func NewBanRequestExecutor(httpPool *HTTPPool, rateLimiter *RateLimitMonitor) *BanRequestExecutor {
	cfg := config.Get()
	return &BanRequestExecutor{
		httpPool:    httpPool,
		rateLimiter: rateLimiter,
		token:       cfg.Bot.Token,
	}
}

func (bre *BanRequestExecutor) ExecuteBan(guildID, userID uint64, reason string) (int64, error) {
	startTime := time.Now()

	if !bre.rateLimiter.CanExecute("ban", guildID) {
		return 0, fmt.Errorf("rate limited")
	}

	url := fmt.Sprintf("https://discord.com/api/v10/guilds/%d/bans/%d", guildID, userID)

	payload := map[string]interface{}{
		"delete_message_seconds": 0,
	}

	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bot %s", bre.token))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Audit-Log-Reason", reason)
	req.Header.Set("Connection", "keep-alive")

	client := bre.httpPool.GetClient()
	requestSentTime := time.Since(startTime)

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	bre.rateLimiter.UpdateFromResponse(resp, "ban", guildID)

	executionTime := time.Since(startTime)
	executionUs := executionTime.Microseconds()
	requestUs := requestSentTime.Microseconds()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		go logging.Info("[🔨 BAN EXECUTED] User: %d | Guild: %d | Prep: %d µs | Total: %d µs | Status: %d",
			userID, guildID, requestUs, executionUs, resp.StatusCode)

		// Add to banned users database
		if db := database.GetDB(); db != nil {
			guildIDStr := strconv.FormatUint(guildID, 10)
			userIDStr := strconv.FormatUint(userID, 10)

			// Check if banned entity is a bot by fetching user info
			isBot := false
			addedBy := ""
			// We don't have session here, so we'll mark isBot as false for now
			// This will be updated in the member join handler if needed

			if err := db.AddBannedUser(guildIDStr, userIDStr, reason, "antinuke-bot", isBot, addedBy); err != nil {
				logging.Warn("Failed to add banned user to database: %v", err)
			} else {
				logging.Info("[DB] Added user %d to banned list for guild %d", userID, guildID)
			}
		}

		return executionUs, nil
	}

	go logging.Error("[❌ BAN FAILED] User: %d | Guild: %d | Time: %d µs | Status: %d",
		userID, guildID, executionUs, resp.StatusCode)
	return 0, fmt.Errorf("ban failed: %d", resp.StatusCode)
}

func (bre *BanRequestExecutor) ExecuteKick(guildID, userID uint64, reason string) error {
	if !bre.rateLimiter.CanExecute("kick", guildID) {
		return fmt.Errorf("rate limited")
	}

	url := fmt.Sprintf("https://discord.com/api/v10/guilds/%d/members/%d", guildID, userID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bot %s", bre.token))
	req.Header.Set("X-Audit-Log-Reason", reason)

	client := bre.httpPool.GetClient()
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bre.rateLimiter.UpdateFromResponse(resp, "kick", guildID)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	return fmt.Errorf("kick failed: %d", resp.StatusCode)
}
