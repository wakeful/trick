// Copyright 2025 variHQ OÜ
// SPDX-License-Identifier: BSD-3-Clause

package parser_test

import (
	"reflect"
	"testing"

	"github.com/wakeful/trick/internal/parser"
)

func TestParseFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		path    string
		want    *parser.Config
		wantErr bool
	}{
		{
			name:    "should fail when file does not exist",
			path:    "./file-does-not-exist.hcl",
			want:    nil,
			wantErr: true,
		},
		{
			name: "",
			path: "./example.config.hcl",
			want: &parser.Config{
				SelectProfile: "simple",
				Profiles: []*parser.Profile{
					{
						Name:   "simple",
						Region: "eu-west-1",
						Chain: &parser.Chain{
							TTL: 5,
							UseRoles: []*parser.UseRoles{
								{
									ARN:  "arn::42::role-a",
									Skip: false,
								},
								{
									ARN:  "arn::42::role-b",
									Skip: false,
								},
								{
									ARN:  "arn::42::role-c",
									Skip: false,
								},
							},
						},
					},
					{
						Name:   "complex",
						Region: "eu-west-1",
						Chain: &parser.Chain{
							TTL: 15,
							UseRoles: []*parser.UseRoles{
								{
									ARN:  "arn::42::role-a",
									Skip: false,
								},
								{
									ARN:  "arn::42::role-b",
									Skip: true,
								},
								{
									ARN:  "arn::42::role-c",
									Skip: true,
								},
								{
									ARN:  "arn::42::role-d",
									Skip: false,
								},
							},
						},
					},
					{
						Name:   "with_defaults",
						Region: "eu-west-1",
						Chain: &parser.Chain{
							TTL: 12,
							UseRoles: []*parser.UseRoles{
								{
									ARN:  "arn::42::role-a",
									Skip: false,
								},
								{
									ARN:  "arn::42::role-b",
									Skip: true,
								},
								{
									ARN:  "arn::42::role-c",
									Skip: true,
								},
								{
									ARN:  "arn::42::role-d",
									Skip: false,
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parser.ParseFile(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFile() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseFile() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_ToFlags(t *testing.T) {
	t.Parallel()

	profiles := []*parser.Profile{
		{
			Name:   "simple",
			Region: "eu-west-1",
			Chain: &parser.Chain{
				TTL: 15,
				UseRoles: []*parser.UseRoles{
					{
						ARN:  "arn::0987654321::role-a",
						Skip: false,
					},
					{
						ARN:  "arn::0987654321::role-b",
						Skip: true,
					},
					{
						ARN:  "arn::0987654321::role-c",
						Skip: true,
					},
					{
						ARN:  "arn::0987654321::role-d",
						Skip: false,
					},
				},
			},
		},
		{
			Name:   "empty-chain",
			Region: "eu-west-1",
			Chain:  nil,
		},
	}

	tests := []struct {
		name          string
		SelectProfile string
		Profiles      []*parser.Profile
		wantTTL       int64
		wantRoles     []string
		wantUsable    []string
		wantErr       bool
	}{
		{
			name:    "should fail when no profiles is selected",
			wantErr: true,
		},
		{
			name:          "simple config with usable roles",
			SelectProfile: "simple",
			Profiles:      profiles,
			wantTTL:       15,
			wantRoles: []string{
				"arn::0987654321::role-a",
				"arn::0987654321::role-b",
				"arn::0987654321::role-c",
				"arn::0987654321::role-d",
			},
			wantUsable: []string{"arn::0987654321::role-b", "arn::0987654321::role-c"},
			wantErr:    false,
		},
		{
			name:          "empty chain should return empty roles and usable roles",
			SelectProfile: "empty-chain",
			Profiles:      profiles,
			wantTTL:       0,
			wantRoles:     []string{},
			wantUsable:    []string{},
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := &parser.Config{
				SelectProfile: tt.SelectProfile,
				Profiles:      tt.Profiles,
			}

			ttl, roles, usableRoles, err := c.ToFlags()
			if (err != nil) != tt.wantErr {
				t.Errorf("ToFlags() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(ttl, tt.wantTTL) {
				t.Errorf("ToFlags() TTL  = %v, want %v", ttl, tt.wantTTL)
			}

			if !reflect.DeepEqual(roles, tt.wantRoles) {
				t.Errorf("ToFlags() roles  = %v, want %v", roles, tt.wantRoles)
			}

			if !reflect.DeepEqual(usableRoles, tt.wantUsable) {
				t.Errorf("ToFlags() usable roles = %v, want %v", usableRoles, tt.wantUsable)
			}
		})
	}
}
