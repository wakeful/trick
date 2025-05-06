// Copyright 2025 variHQ OÜ
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
)

type MockCmdExecutor struct {
	executeFunc func(command string, args ...string) ([]byte, error)
}

func (m MockCmdExecutor) Execute(name string, arg ...string) ([]byte, error) {
	return m.executeFunc(name, arg...)
}

var _ CmdExecutor = (*MockCmdExecutor)(nil)

func TestProfileWriter_writeAWSProfile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		profileName string
		credentials *types.Credentials
		executeFunc func(command string, args ...string) ([]byte, error)
		region      string
		wantErr     bool
	}{
		{
			name:        "no credentials",
			profileName: "testing-profile",
			credentials: nil,
			executeFunc: func(_ string, _ ...string) ([]byte, error) { return nil, nil },
			region:      "eu-west-1",
			wantErr:     true,
		},
		{
			name:        "provided credentials",
			profileName: "testing-profile",
			credentials: &types.Credentials{
				AccessKeyId:     aws.String("access-key-id"),
				SecretAccessKey: aws.String("secret-access-key"),
				SessionToken:    aws.String("session-token"),
			},
			executeFunc: func(_ string, _ ...string) ([]byte, error) { return nil, nil },
			region:      "eu-west-1",
			wantErr:     false,
		},
		{
			name:        "provided credentials but failed on write",
			profileName: "testing-profile",
			credentials: &types.Credentials{
				AccessKeyId:     aws.String("access-key-id"),
				SecretAccessKey: aws.String("secret-access-key"),
				SessionToken:    aws.String("session-token"),
			},
			executeFunc: func(_ string, _ ...string) ([]byte, error) {
				return nil, errors.New("failed to write profile")
			},
			region:  "eu-west-1",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := &ProfileWriter{
				cmdExecutor: &MockCmdExecutor{
					executeFunc: tt.executeFunc,
				},
				profileName: tt.profileName,
			}
			if err := p.writeAWSProfile(tt.credentials, tt.region); (err != nil) != tt.wantErr {
				t.Errorf("writeAWSProfile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestApp_run(t *testing.T) {
	t.Parallel()

	pool, err := setRolePool([]string{"arn:aws:iam::0987654321:role/role-a", "arn:aws:iam::0987654321:role/role-b"})
	if err != nil {
		t.Fatalf("setRolePool failed: %v", err)
	}

	tests := []struct {
		name             string
		profileWriteFunc func(_ string, _ ...string) ([]byte, error)
		runDuration      time.Duration
		useCancel        bool
	}{
		{
			name:             "runs successfully",
			profileWriteFunc: func(_ string, _ ...string) ([]byte, error) { return nil, nil },
			runDuration:      300 * time.Millisecond,
			useCancel:        false,
		},
		{
			name:             "fails when writing profiles",
			profileWriteFunc: func(_ string, _ ...string) ([]byte, error) { return nil, errors.New("failed to write profile") },
			runDuration:      300 * time.Millisecond,
			useCancel:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			credentials := &types.Credentials{
				AccessKeyId:     aws.String("access-key-id"),
				SecretAccessKey: aws.String("secret-access-key"),
				SessionToken:    aws.String("session-token"),
			}
			a := &App{
				client: MockSTSClient{
					mockAssumeRoleOutput: map[string]sts.AssumeRoleOutput{
						"arn:aws:iam::0987654321:role/role-a": {
							Credentials: credentials,
						},
						"arn:aws:iam::0987654321:role/role-b": {
							Credentials: credentials,
						},
					},
					mockAssumeRoleError: nil,
				},

				profileWriter: &ProfileWriter{
					cmdExecutor: &MockCmdExecutor{
						executeFunc: tt.profileWriteFunc,
					},
					profileName: "testing-profile",
				},
				region:      "eu-west-1",
				roles:       pool,
				usableRoles: make(map[string]struct{}),
			}

			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			done := make(chan struct{})

			go func() {
				a.run(ctx, time.NewTicker(tt.runDuration))
				close(done)
			}()

			if tt.useCancel {
				time.Sleep(100 * time.Millisecond)
				cancel()
			} else {
				time.Sleep(tt.runDuration)
				cancel()
			}

			select {
			case <-done:
				break
			case <-time.After(100 * time.Millisecond):
				t.Fatal("run() timed out")
			}
		})
	}
}
