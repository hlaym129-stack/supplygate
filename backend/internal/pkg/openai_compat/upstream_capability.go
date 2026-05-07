// Package openai_compat provides OpenAI upstream capability checks.
package openai_compat

// AccountResponsesSupport describes whether an OpenAI APIKey account's
// upstream supports the Responses API.
type AccountResponsesSupport int

const (
	// ResponsesSupportUnknown means the account has not been probed yet.
	ResponsesSupportUnknown AccountResponsesSupport = iota
	// ResponsesSupportYes means the upstream supports /v1/responses.
	ResponsesSupportYes
	// ResponsesSupportNo means the upstream does not support /v1/responses.
	ResponsesSupportNo
)

// ExtraKeyResponsesSupported is the accounts.extra JSON key for the probe
// result. Values are bool: true=supported, false=unsupported, missing=unknown.
const ExtraKeyResponsesSupported = "openai_responses_supported"

// ResolveResponsesSupport reads the probed Responses API capability from an
// account extra map. Missing or malformed values stay unknown.
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

// ShouldUseResponsesAPI returns false only when a probe explicitly confirmed
// the upstream cannot handle /v1/responses. Unknown preserves existing behavior.
func ShouldUseResponsesAPI(extra map[string]any) bool {
	return ResolveResponsesSupport(extra) != ResponsesSupportNo
}
