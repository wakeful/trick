// Copyright 2025 variHQ OÜ
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"container/ring"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
)

func TestApp_assumeRole(t *testing.T) {
	t.Parallel()

	type fields struct {
		client          MockSTSClient
		region          string
		sessionDuration time.Duration
	}

	tests := []struct {
		name    string
		fields  fields
		role    string
		wantErr bool
	}{
		{
			name: "we are allowed to assume role-a",
			fields: fields{
				client: MockSTSClient{
					mockAssumeRoleOutput: map[string]sts.AssumeRoleOutput{
						"arn:aws:iam::0987654321:role/role-a": {
							Credentials: &types.Credentials{
								AccessKeyId:     aws.String("access-key-id"),
								SecretAccessKey: aws.String("secret-access-key"),
								SessionToken:    aws.String("session-token"),
							},
						},
					},
					mockAssumeRoleError: nil,
				},
				region:          "eu-west-1",
				sessionDuration: 42 * time.Second,
			},
			role:    "arn:aws:iam::0987654321:role/role-a",
			wantErr: false,
		},
		{
			name: "we should fail when trying to assume role-b",
			fields: fields{
				client: MockSTSClient{
					mockAssumeRoleOutput: make(map[string]sts.AssumeRoleOutput),
					mockAssumeRoleError:  errors.New("error"),
				},
				region:          "eu-west-1",
				sessionDuration: 42 * time.Second,
			},
			role:    "arn:aws:iam::0987654321:role/role-b",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			a := &App{
				client:          tt.fields.client,
				region:          tt.fields.region,
				sessionDuration: tt.fields.sessionDuration,
			}

			_, err := a.assumeRole(t.Context(), tt.role)
			if (err != nil) != tt.wantErr {
				t.Errorf("assumeRole() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestApp_assumeNextInterestingRole(t *testing.T) {
	t.Parallel()

	roles, _ := setRolePool([]string{
		"arn:aws:iam::0987654321:role/role-a",
		"arn:aws:iam::0987654321:role/role-b",
		"arn:aws:iam::0987654321:role/role-c",
	})

	credentials := &types.Credentials{
		AccessKeyId:     aws.String("access-key-id"),
		SecretAccessKey: aws.String("secret-access-key"),
		SessionToken:    aws.String("session-token"),
	}

	type fields struct {
		client          MockSTSClient
		region          string
		roles           *ring.Ring
		usableRoles     map[string]struct{}
		sessionDuration time.Duration
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "we are allowed to assume role-a",
			fields: fields{
				client: MockSTSClient{
					mockAssumeRoleOutput: map[string]sts.AssumeRoleOutput{
						"arn:aws:iam::0987654321:role/role-a": {
							Credentials: credentials,
						},
					},
					mockAssumeRoleError: nil,
				},
				region:          "eu-west-1",
				roles:           roles,
				usableRoles:     make(map[string]struct{}),
				sessionDuration: 42 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "we should fail when trying to assume role-b",
			fields: fields{
				client: MockSTSClient{
					mockAssumeRoleOutput: make(map[string]sts.AssumeRoleOutput),
					mockAssumeRoleError:  errors.New("error"),
				},
				region:          "eu-west-1",
				roles:           roles,
				usableRoles:     make(map[string]struct{}),
				sessionDuration: 42 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "role-a is in the usableRoles map and can be assumed successfully",
			fields: fields{
				client: MockSTSClient{
					mockAssumeRoleOutput: map[string]sts.AssumeRoleOutput{
						"arn:aws:iam::0987654321:role/role-a": {
							Credentials: credentials,
						},
					},
					mockAssumeRoleError: nil,
				},
				region: "eu-west-1",
				roles:  roles,
				usableRoles: map[string]struct{}{
					"arn:aws:iam::0987654321:role/role-a": {},
				},
				sessionDuration: 42 * time.Second,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			a := &App{
				client:          tt.fields.client,
				region:          tt.fields.region,
				roles:           tt.fields.roles,
				usableRoles:     tt.fields.usableRoles,
				sessionDuration: tt.fields.sessionDuration,
			}

			_, err := a.assumeNextInterestingRole(t.Context())
			if (err != nil) != tt.wantErr {
				t.Errorf("assumeNextInterestingRole() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
