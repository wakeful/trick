// Copyright 2025 variHQ OÜ
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sts/types"
)

func (a *App) run(ctx context.Context, ticker *time.Ticker) {
	if err := a.tick(ctx); err != nil {
		slog.Error("initial tick failed", slog.String("error", err.Error()))

		return
	}

	for {
		select {
		case <-ticker.C:
			if err := a.tick(ctx); err != nil {
				slog.Error("tick failed", slog.String("error", err.Error()))

				return
			}

			slog.Info("credentials refresh")

		case <-ctx.Done():
			ticker.Stop()

			return
		}
	}
}

func (a *App) tick(ctx context.Context) error {
	credentials, err := a.assumeNextInterestingRole(ctx)
	if err != nil {
		return fmt.Errorf("unable to assume role on tick: %w", err)
	}

	if errWrite := writeAWSProfile(credentials, a.region); errWrite != nil {
		return fmt.Errorf("unable to write AWS credentials: %w", errWrite)
	}

	return nil
}

func writeAWSProfile(credentials *types.Credentials, region string) error {
	const profileName = "trick-jump-credentials"
	commands := []struct {
		args []string
		desc string
	}{
		{
			args: []string{"configure", "set", "aws_access_key_id", *credentials.AccessKeyId, "--profile", profileName},
			desc: "setting access key",
		},
		{
			args: []string{"configure", "set", "aws_secret_access_key", *credentials.SecretAccessKey, "--profile", profileName},
			desc: "setting secret key",
		},
		{
			args: []string{"configure", "set", "aws_session_token", *credentials.SessionToken, "--profile", profileName},
			desc: "setting secret key",
		},
		{
			args: []string{"configure", "set", "region", region, "--profile", profileName},
			desc: "setting region",
		},
	}

	for _, cmd := range commands {
		_, err := exec.Command("aws", cmd.args...).CombinedOutput()
		if err != nil {
			slog.Error(cmd.desc, slog.String("error", err.Error()))

			return fmt.Errorf("failed to execute aws command %q: %w", strings.Join(cmd.args, " "), err)
		}
	}

	slog.Debug("aws credentials updated", slog.String("profile", profileName))

	return nil
}
