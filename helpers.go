// Copyright 2025 variHQ OÜ
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"container/ring"
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// getID retrieves the AWS identity ARN of the caller using STS.
// It returns the ARN as a string or an error if the identity cannot be retrieved.
func (a *App) getID(ctx context.Context) (string, error) {
	identity, err := a.client.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", fmt.Errorf("unable to get identity, %w", err)
	}

	return *identity.Arn, nil
}

// nextRole advances to the next role in the role pool and returns it.
// If the current value in the role pool is not a string, it returns an empty string.
func (a *App) nextRole() string {
	value, ok := a.roles.Value.(string)
	if !ok {
		return ""
	}

	a.roles = a.roles.Next()

	slog.Debug("next role", slog.String("role", value))

	return value
}

// setRolePool initializes a circular role pool with the provided roles.
// It requires at least two roles to work properly and returns an error otherwise.
// Returns the initialized role pool and nil error on success.
func setRolePool(roles []string) (*ring.Ring, error) {
	const minRoles = 2
	if len(roles) < minRoles {
		return nil, ErrMinRoles
	}

	rolesPool := ring.New(len(roles))
	for i := range rolesPool.Len() {
		rolesPool.Value = roles[i]
		rolesPool = rolesPool.Next()
	}

	return rolesPool, nil
}

// getLogger creates and returns a slog.Logger configured with the specified output and log level.
// The verbosity is set to debug if verbose is true; otherwise, it defaults to info level.
func getLogger(output io.Writer, verbose *bool) *slog.Logger {
	logLevel := slog.LevelInfo
	if verbose != nil && *verbose {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(
		slog.NewTextHandler(output, &slog.HandlerOptions{ //nolint:exhaustruct
			AddSource: false,
			Level:     logLevel,
		}),
	)

	return logger
}
