// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0
package tracing

import (
	"context"

	"github.com/ultravioletrs/cocos/manager"
	"go.opentelemetry.io/otel/trace"
)

var _ manager.Service = (*tracingMiddleware)(nil)

type tracingMiddleware struct {
	tracer trace.Tracer
	svc    manager.Service
}

// New returns a new auth service with tracing capabilities.
func New(svc manager.Service, tracer trace.Tracer) manager.Service {
	return &tracingMiddleware{tracer, svc}
}

func (tm *tracingMiddleware) Run(ctx context.Context, mc *manager.ComputationRunReq) (string, error) {
	ctx, span := tm.tracer.Start(ctx, "run")
	defer span.End()

	return tm.svc.Run(ctx, mc)
}

func (tm *tracingMiddleware) Stop(ctx context.Context, computationID string) error {
	ctx, span := tm.tracer.Start(ctx, "stop")
	defer span.End()

	return tm.svc.Stop(ctx, computationID)
}

func (tm *tracingMiddleware) FetchBackendInfo(ctx context.Context, computationId string) ([]byte, error) {
	_, span := tm.tracer.Start(ctx, "fetch_backend_info")
	defer span.End()

	return tm.svc.FetchBackendInfo(ctx, computationId)
}

func (tm *tracingMiddleware) ReportBrokenConnection(addr string) {
	tm.svc.ReportBrokenConnection(addr)
}
