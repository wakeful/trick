// Copyright 2025 variHQ OÜ
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
)

func (a *App) assumeRole(ctx context.Context, role string) (*types.Credentials, error) {
	slog.Debug("assuming role", slog.String("role", role))

	assumeRole, err := a.client.AssumeRole(ctx, &sts.AssumeRoleInput{ //nolint:exhaustruct
		RoleArn:         aws.String(role),
		RoleSessionName: aws.String("trick"),
		DurationSeconds: aws.Int32(int32(a.sessionDuration.Seconds())),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to assume role, %w", err)
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(
		credentials.NewStaticCredentialsProvider(
			*assumeRole.Credentials.AccessKeyId,
			*assumeRole.Credentials.SecretAccessKey,
			*assumeRole.Credentials.SessionToken,
		),
	), config.WithRegion(a.region))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %w", err)
	}

	slog.Debug("replacing client", slog.String("role", role))

	a.client = sts.NewFromConfig(cfg)

	return assumeRole.Credentials, nil
}

func (a *App) assumeNextInterestingRole(ctx context.Context) (*types.Credentials, error) {
	var outputCred *types.Credentials

	for {
		role := a.nextRole()

		slog.Info("trying to assume role", slog.String("role", role))

		cred, err := a.assumeRole(ctx, role)
		if err != nil {
			return nil, fmt.Errorf("unable to assume role, %w", err)
		}

		outputCred = cred

		if len(a.usableRoles) == 0 {
			slog.Debug("all roles have meaningful permissions")

			break
		}

		if _, ok := a.usableRoles[role]; ok {
			slog.Debug("found role with meaningful permissions", slog.String("role", role))

			break
		}

		slog.Debug("role is lacking meaningful permissions", slog.String("role", role))
	}

	return outputCred, nil
}
