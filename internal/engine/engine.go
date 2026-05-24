// Package engine Soul 人格/议题引擎接口。Phase-0：StubEngine，核心整理逻辑未实现。
package engine

import "context"

// StoreInput soul_store 入参（协议层 string）。
type StoreInput struct {
	Content       string
	Source        string
	Kind          string
	CorrelationID string
}

// RetrieveInput soul_retrieve 入参。
type RetrieveInput struct {
	Context   string
	QueryHint string
}

// Engine 对外 store/retrieve。
type Engine interface {
	Store(ctx context.Context, in StoreInput) string
	Retrieve(ctx context.Context, in RetrieveInput) string
}
