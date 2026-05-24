// Package response 将工具结果编码为对外约定的 JSON 字符串。
package response

import (
	"encoding/json"
	"fmt"
)

// StorePayload soul_store 返回体。
type StorePayload struct {
	Accepted   string `json:"accepted"`
	JobID      string `json:"job_id"`
	Skipped    string `json:"skipped"`
	SkipReason string `json:"skip_reason,omitempty"`
	Message    string `json:"message"`
	Phase      string `json:"phase"`
}

// RetrievePayload soul_retrieve 返回体。
type RetrievePayload struct {
	Hints      string `json:"hints"`
	Skipped    string `json:"skipped"`
	SkipReason string `json:"skip_reason,omitempty"`
	Phase      string `json:"phase"`
}

// PhaseStub 当前阶段标识。
func PhaseStub() string { return "0-stub" }

// FormatStore 返回 JSON 字符串。
func FormatStore(p StorePayload) string {
	if p.Phase == "" {
		p.Phase = PhaseStub()
	}
	b, err := json.Marshal(p)
	if err != nil {
		return fmt.Sprintf(`{"accepted":"false","job_id":"","skipped":"false","message":"encode error: %s","phase":"%s"}`, err.Error(), PhaseStub())
	}
	return string(b)
}

// FormatRetrieve 返回 JSON 字符串。
func FormatRetrieve(p RetrievePayload) string {
	if p.Phase == "" {
		p.Phase = PhaseStub()
	}
	b, err := json.Marshal(p)
	if err != nil {
		return fmt.Sprintf(`{"hints":"","skipped":"false","message":"encode error: %s","phase":"%s"}`, err.Error(), PhaseStub())
	}
	return string(b)
}
