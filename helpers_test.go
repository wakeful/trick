// Copyright 2025 variHQ OÜ
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func TestApp_nextRole(t *testing.T) {
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
	tests := []struct {
		name    string
		client  ServiceSTS
		ctx     context.Context
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
			ctx:     t.Context(),
			want:    "arn::mock::0987654321::role-a",
			wantErr: false,
		},
		{
			name: "get caller identity error",
			client: &MockSTSClient{
				mockGetCallerIdentityOutput: sts.GetCallerIdentityOutput{},
				mockGetCallerIdentityError:  errors.New("error"),
			},
			ctx:     t.Context(),
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &App{
				client: tt.client,
			}

			got, err := a.getID(tt.ctx)
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
