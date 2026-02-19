package checker

import "context"

func getNodeName(raw map[string]any) string {
	if name, ok := raw["name"].(string); ok && name != "" {
		return name
	}
	if server, ok := raw["server"].(string); ok && server != "" {
		return server
	}
	return "unknown"
}

func getCheckID(ctx context.Context) uint16 {
	if ctx == nil {
		return 0
	}
	if val := ctx.Value("check_id"); val != nil {
		if id, ok := val.(uint16); ok {
			return id
		}
	}
	return 0
}
