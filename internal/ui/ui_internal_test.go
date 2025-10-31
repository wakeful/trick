// Copyright 2025 variHQ OÜ
// SPDX-License-Identifier: BSD-3-Clause

package ui

import (
	"strings"
	"testing"
)

func Test_flagsToDiagram(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		roles          []string
		usableRoles    map[string]struct{}
		refreshMinutes int64
		wantContains   []string
	}{
		{
			name: "simple scenario with three roles",
			roles: []string{
				"arn::42::role-a",
				"arn::42::role-b",
				"arn::42::role-c",
			},
			usableRoles:    map[string]struct{}{},
			refreshMinutes: 12,
			wantContains: []string{
				"stateDiagram",
				"r0: role-a",
				"r1: role-b",
				"r2: role-c",
				"[*] --> r0",
				"r0 --> r1: wait 12min and jump",
				"r1 --> r2: wait 12min and jump",
				"r2 --> r0: wait 12min and jump",
			},
		},
		{
			name: "complex scenario with usable roles",
			roles: []string{
				"arn::42::role-a",
				"arn::42::role-b",
				"arn::42::role-c",
				"arn::42::role-d",
			},
			usableRoles: map[string]struct{}{
				"arn::42::role-a": {},
				"arn::42::role-d": {},
			},
			refreshMinutes: 12,
			wantContains: []string{
				"stateDiagram",
				"r0: role-a",
				"r1: role-b",
				"r2: role-c",
				"r3: role-d",
				"[*] --> r0",
				"r0 --> r1: wait 12min and jump",
				"r1 --> r2: lacks permission so we jump to role-c",
				"r2 --> r3: lacks permission so we jump to role-d",
				"r3 --> r0: wait 12min and jump",
			},
		},
		{
			name: "roles with full ARN format",
			roles: []string{
				"arn:aws:iam::123456789012:role/MyRole",
				"arn:aws:iam::123456789012:role/AnotherRole",
			},
			usableRoles:    map[string]struct{}{},
			refreshMinutes: 15,
			wantContains: []string{
				"stateDiagram",
				"r0: MyRole",
				"r1: AnotherRole",
				"[*] --> r0",
				"r0 --> r1: wait 15min and jump",
				"r1 --> r0: wait 15min and jump",
			},
		},
		{
			name:           "empty roles returns empty string",
			roles:          []string{},
			usableRoles:    map[string]struct{}{},
			refreshMinutes: 12,
			wantContains:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := flagsToDiagram(tt.roles, tt.usableRoles, tt.refreshMinutes)

			if len(tt.roles) == 0 {
				if got != "" {
					t.Errorf(
						"generateMermaidDiagram() with empty roles should return empty string, got %q",
						got,
					)
				}

				return
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf(
						"generateMermaidDiagram() missing expected content:\nwant substring: %q\ngot: %q",
						want,
						got,
					)
				}
			}
		})
	}
}
