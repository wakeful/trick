// Copyright 2025 variHQ OÜ
// SPDX-License-Identifier: BSD-3-Clause

// Package main provides functionality for assuming AWS roles and managing credentials.
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wakeful/trick/internal/parser"
)

const (
	// cleanupWaitDuration provides a short delay before termination to ensure proper cleaned up.
	cleanupWaitDuration = 100 * time.Millisecond
	defaultProfileName  = "trick-jump-credentials"
	defaultRefreshTime  = 12
)

var version = "dev"

//nolint:funlen
func main() {
	config := flag.String("config", "", "path to config file")
	refresh := flag.Int64("refresh", defaultRefreshTime, "refresh IAM every n minutes")
	region := flag.String("region", "eu-west-1", "AWS region used for IAM communication")
	showVersion := flag.Bool("version", false, "show version")
	verbose := flag.Bool("verbose", false, "verbose log output")

	var (
		roleVars    StringSlice
		useRoleVars StringSlice
	)

	flag.Var(
		&roleVars,
		"role",
		"AWS role ARN to assume (can be specified multiple times, at least 2 required)",
	)
	flag.Var(
		&useRoleVars,
		"use",
		"AWS role ARN with meaningful permissions to prioritize (must exist in -role list)",
	)

	flag.Parse()

	slog.SetDefault(getLogger(os.Stderr, verbose))

	if *showVersion {
		slog.Info(
			"trick",
			slog.String("repo", "https://github.com/wakeful/trick"),
			slog.String("version", version),
		)

		return
	}

	if *config != "" {
		slog.Debug("loading config file", slog.String("path", *config))

		cfgFile, err := parser.ParseFile(*config)
		if err != nil {
			slog.Error("failed to parse config file", slog.String("error", err.Error()))

			return
		}

		*refresh, roleVars, useRoleVars, err = cfgFile.ToFlags()
		if err != nil {
			slog.Error("failed to convert config to flags", slog.String("error", err.Error()))

			return
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	signalCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer stop()

	slog.Info("starting app")

	ticker := time.NewTicker(time.Minute * time.Duration(*refresh))

	if *refresh < 1 {
		slog.Warn("refresh interval too low, setting to 1 minute")

		*refresh = 1
		ticker = time.NewTicker(time.Minute)
	}

	defer ticker.Stop()

	go func() {
		<-signalCtx.Done()
		slog.Info("signal received, shutting down...")
		cancel()
	}()

	app, err := NewApp(ctx, *region, roleVars, useRoleVars)
	if err != nil {
		slog.Error("failed to initialize app", slog.String("error", err.Error()))

		return
	}

	app.run(ctx, ticker)

	slog.Info("cleaning up resources...")
	time.Sleep(cleanupWaitDuration)
	slog.Info("application terminated gracefully")
}
