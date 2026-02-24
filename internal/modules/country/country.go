package country

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/evilCYH/NodeHub/internal/modules/country/channel"
)

func GetCode(ctx context.Context, client *http.Client) string {
	for _, channel := range channel.Channels {
		ctx, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		request, err := http.NewRequestWithContext(ctx, "GET", channel.Url(), nil)
		if err != nil {
			continue
		}
		channel.Header(request)
		response, err := client.Do(request)
		if err != nil {
			continue
		}
		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		if err != nil {
			continue
		}
		country := channel.CountryCode(body)
		if country != "" {
			return country
		}
		body = nil
	}
	return ""
}
