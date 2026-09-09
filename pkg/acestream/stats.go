package acestream

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// StreamStatus represents the status of a stream
type StreamStatus string

// Stream statuses reported by acestream-engine.
const (
	StatusPrebuf  StreamStatus = "prebuf"
	StatusDL      StreamStatus = "dl"
	StatusCheck   StreamStatus = "check"
	StatusWait    StreamStatus = "wait"
	StatusIdle    StreamStatus = "idle"
	StatusLoading StreamStatus = "loading"
	StatusError   StreamStatus = "error"
)

// StreamStats represents stream statistics
type StreamStats struct {
	Status        StreamStatus `json:"status"`
	Progress      float64      `json:"progress"`
	DownloadSpeed int64        `json:"download_speed"`
	UploadSpeed   int64        `json:"upload_speed"`
	Peers         int          `json:"peers"`
}

// GetStats gets the statistics for a stream
func (c *Client) GetStats(ctx context.Context, contentID string) (*StreamStats, error) {
	body, err := c.getBody(ctx, "/ace/getstream", url.Values{
		"content_id": {contentID},
		"method":     {"get_stats"},
	})
	if err != nil {
		return nil, fmt.Errorf("error getting statistics: %w", err)
	}

	var stats StreamStats
	if err := json.Unmarshal(body, &stats); err != nil {
		// The engine does not always answer JSON — fall back to query string
		return parseStats(string(body))
	}

	return &stats, nil
}

// parseStats parses statistics served as a query string instead of JSON.
func parseStats(body string) (*StreamStats, error) {
	values, err := url.ParseQuery(body)
	if err != nil {
		return nil, fmt.Errorf("error parsing statistics: %w", err)
	}

	stats := &StreamStats{
		Status: StreamStatus(values.Get("status")),
	}

	if progress := values.Get("progress"); progress != "" {
		stats.Progress, _ = strconv.ParseFloat(progress, 64)
	}

	if speed := values.Get("download_speed"); speed != "" {
		stats.DownloadSpeed, _ = strconv.ParseInt(speed, 10, 64)
	}

	if speed := values.Get("upload_speed"); speed != "" {
		stats.UploadSpeed, _ = strconv.ParseInt(speed, 10, 64)
	}

	if peers := values.Get("peers"); peers != "" {
		stats.Peers, _ = strconv.Atoi(peers)
	}

	return stats, nil
}
