package subconv

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/bestruirui/bestsub/internal/core/mihomo"
	"github.com/bestruirui/bestsub/internal/database/op"
	"github.com/bestruirui/bestsub/internal/models/setting"
	"github.com/bestruirui/bestsub/internal/utils/log"
)

var ErrSubconvURLNotSet = errors.New("subconv url is not set")

func ConvertData(ctx context.Context, raw string, target string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	subStoreUrl := strings.TrimSpace(op.GetSettingStr(setting.SUBCONV_URL))
	if subStoreUrl == "" {
		log.Errorf("substore url is not set")
		return "", ErrSubconvURLNotSet
	}
	client := mihomo.Default(op.GetSettingBool(setting.SUBCONV_URL_PROXY))
	if client == nil {
		err := errors.New("failed to create http client")
		log.Errorf("%v", err)
		return "", err
	}
	defer client.Release()
	reqBody := struct {
		Data   string `json:"data"`
		Client string `json:"client"`
	}{
		Data:   raw,
		Client: target,
	}
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		log.Errorf("failed to marshal request body: %v", err)
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", subStoreUrl+"/api/proxy/parse", bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		log.Errorf("failed to create request: %v", err)
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("failed to do request: %v", err)
		return "", fmt.Errorf("failed to do request: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("failed to read response body: %v", err)
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		err := fmt.Errorf("unexpected response status: %d", resp.StatusCode)
		log.Errorf("%v", err)
		return "", err
	}
	var respBody struct {
		Data struct {
			Parres string `json:"par_res"`
		}
	}
	err = json.Unmarshal(body, &respBody)
	if err != nil {
		log.Errorf("failed to unmarshal response body: %v body: %s", err, string(body))
		return "", fmt.Errorf("failed to unmarshal response body: %w", err)
	}
	if strings.TrimSpace(respBody.Data.Parres) == "" {
		err := errors.New("subconv response is empty")
		log.Errorf("%v", err)
		return "", err
	}
	return respBody.Data.Parres, nil
}
