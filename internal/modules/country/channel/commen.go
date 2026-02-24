package channel

import (
	"net/http"

	"github.com/evilCYH/NodeHub/internal/utils/ua"
)

type Common struct {
	CountryCode string `json:"country_code"`
}

func UserAgent(req *http.Request) {
	ua.SetHeader(req)
}
