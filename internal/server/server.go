// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0
package server

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ultravioletrs/cocos/pkg/clients/grpc"
)

//go:generate mockery --name Server --output ./mocks --filename server.go
type Server interface {
	Start() error
	Stop() error
}

type AgentConfig struct {
	grpc.BaseConfig
	CertFile     string `env:"SERVER_CERT"        envDefault:""`
	KeyFile      string `env:"SERVER_KEY"         envDefault:""`
	ServerCAFile string `env:"SERVER_CA_CERTS"    envDefault:""`
	ClientCAFile string `env:"CLIENT_CA_CERTS"    envDefault:""`
	AttestedTLS  bool   `env:"ATTESTED_TLS"       envDefault:"false"`
}

type BaseServer struct {
	Ctx      context.Context
	Cancel   context.CancelFunc
	Name     string
	Address  string
	Config   AgentConfig
	Logger   *slog.Logger
	Protocol string
}

func stopAllServer(servers ...Server) error {
	var errs []error
	for _, server := range servers {
		if err := server.Stop(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("encountered errors while stopping servers: %v", errs)
	}

	return nil
}

func StopHandler(ctx context.Context, cancel context.CancelFunc, logger *slog.Logger, svcName string, servers ...Server) error {
	var err error
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGABRT)
	select {
	case sig := <-c:
		defer cancel()
		err = stopAllServer(servers...)
		if err != nil {
			logger.Error(fmt.Sprintf("%s service error during shutdown: %v", svcName, err))
		}
		logger.Info(fmt.Sprintf("%s service shutdown by signal: %s", svcName, sig))
		return err
	case <-ctx.Done():
		return nil
	}
}
