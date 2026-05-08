// Package openai_compat provides OpenAI protocol compatibility helpers for
// upstreams with different endpoint support.
package openai_compat

type AccountResponsesSupport int

const (
	ResponsesSupportUnknown AccountResponsesSupport = iota
	ResponsesSupportYes
	ResponsesSupportNo
)

const ExtraKeyResponsesSupported = "openai_responses_supported"

func ResolveResponsesSupport(extra map[string]any) AccountResponsesSupport {
	if extra == nil {
		return ResponsesSupportUnknown
	}
	v, ok := extra[ExtraKeyResponsesSupported]
	if !ok {
		return ResponsesSupportUnknown
	}
	supported, ok := v.(bool)
	if !ok {
		return ResponsesSupportUnknown
	}
	if supported {
		return ResponsesSupportYes
	}
	return ResponsesSupportNo
}

func ShouldUseResponsesAPI(extra map[string]any) bool {
	return ResolveResponsesSupport(extra) != ResponsesSupportNo
}
