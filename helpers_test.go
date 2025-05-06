// Copyright 2025 variHQ OÜ
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"bytes"
	"errors"
	"log/slog"
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func TestApp_nextRole(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		given   []string
		want    []string
		wantErr bool
	}{
		{
			name: "next 5 roles",
			given: []string{
				"role-a",
				"role-b",
				"role-c",
			},
			want: []string{
				"role-a",
				"role-b",
				"role-c",
				"role-a",
				"role-b",
			},
			wantErr: false,
		},
		{
			name:    "we need more than two roles",
			given:   []string{},
			want:    []string{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pool, err := setRolePool(tt.given)
			if (err != nil) != tt.wantErr {
				t.Errorf("setRolePool() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			a := &App{
				roles: pool,
			}

			var got []string
			for range len(tt.want) {
				got = append(got, a.nextRole())
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("App().nextRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApp_getID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		client  ServiceSTS
		want    string
		wantErr bool
	}{
		{
			name: "get caller identity",
			client: &MockSTSClient{
				mockGetCallerIdentityOutput: sts.GetCallerIdentityOutput{
					Arn: aws.String("arn::mock::0987654321::role-a"),
				},
				mockGetCallerIdentityError: nil,
			},
			want:    "arn::mock::0987654321::role-a",
			wantErr: false,
		},
		{
			name: "get caller identity error",
			client: &MockSTSClient{
				mockGetCallerIdentityOutput: sts.GetCallerIdentityOutput{},
				mockGetCallerIdentityError:  errors.New("error"),
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			a := &App{
				client: tt.client,
			}

			got, err := a.getID(t.Context())
			if (err != nil) != tt.wantErr {
				t.Errorf("getID() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if got != tt.want {
				t.Errorf("getID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getLogger(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		verbose          *bool
		testMessage      string
		useLevel         slog.Level
		expectInOutput   bool
		testOutputSearch string
	}{
		{
			name:             "returns configured logger with verbose=false",
			verbose:          aws.Bool(false),
			testMessage:      "debug message",
			useLevel:         slog.LevelDebug,
			expectInOutput:   false,
			testOutputSearch: "debug message",
		},
		{
			name:             "enables debug logging when verbose=true",
			verbose:          aws.Bool(true),
			testMessage:      "debug message",
			useLevel:         slog.LevelDebug,
			expectInOutput:   true,
			testOutputSearch: "debug message",
		},
		{
			name:             "info messages are always logged",
			verbose:          aws.Bool(false),
			testMessage:      "info message",
			useLevel:         slog.LevelInfo,
			expectInOutput:   true,
			testOutputSearch: "info message",
		},
		{
			name:             "handles nil verbose flag",
			verbose:          nil,
			testMessage:      "info message",
			useLevel:         slog.LevelInfo,
			expectInOutput:   true,
			testOutputSearch: "info message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			logger := getLogger(&buf, tt.verbose)

			if logger == nil {
				t.Fatal("Expected logger to not be nil")
			}

			switch tt.useLevel {
			case slog.LevelDebug:
				logger.Debug(tt.testMessage)
			case slog.LevelInfo:
				logger.Info(tt.testMessage)
			case slog.LevelWarn:
				logger.Warn(tt.testMessage)
			case slog.LevelError:
				logger.Error(tt.testMessage)
			}

			logOutput := buf.String()
			containsMessage := strings.Contains(logOutput, tt.testOutputSearch)

			if containsMessage != tt.expectInOutput {
				if tt.expectInOutput {
					t.Errorf("Expected log output to contain '%s', but it did not. Got: %s",
						tt.testOutputSearch, logOutput)
				} else {
					t.Errorf("Expected log output to NOT contain '%s', but it did. Got: %s",
						tt.testOutputSearch, logOutput)
				}
			}
		})
	}
}
