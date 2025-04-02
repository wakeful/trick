// Copyright 2025 variHQ OÜ
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"container/ring"
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func (a *App) getID(ctx context.Context) (string, error) {
	identity, err := a.client.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", fmt.Errorf("unable to get identity, %w", err)
	}

	return *identity.Arn, nil
}

func (a *App) nextRole() string {
	value := fmt.Sprintf("%s", a.roles.Value)
	a.roles = a.roles.Next()

	slog.Debug("next role", slog.String("role", value))

	return value
}

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
