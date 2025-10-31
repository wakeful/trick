// Copyright 2025 variHQ OÜ
// SPDX-License-Identifier: BSD-3-Clause

package ui_test

import (
	"strings"
	"testing"

	"github.com/wakeful/trick/internal/ui"
)

func TestRenderDiagramHTML(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		region         string
		roles          []string
		usableRoles    map[string]struct{}
		refreshMinutes int64
		wantErr        bool
		wantContains   []string
	}{
		{
			name:   "successfully renders diagram with roles",
			region: "us-west-2",
			roles: []string{
				"arn:aws:iam::123456789012:role/RoleA",
				"arn:aws:iam::123456789012:role/RoleB",
				"arn:aws:iam::123456789012:role/RoleC",
			},
			usableRoles: map[string]struct{}{
				"arn:aws:iam::123456789012:role/RoleA": {},
			},
			refreshMinutes: 10,
			wantErr:        false,
			wantContains: []string{
				"<!doctype html>",
				"10",
				"stateDiagram",
				"RoleA",
				"RoleB",
				"RoleC",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := ui.RenderDiagramHTML(tt.roles, tt.usableRoles, tt.refreshMinutes)

			if (err != nil) != tt.wantErr {
				t.Errorf("renderDiagramHTML() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if tt.wantErr {
				return
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf(
						"renderDiagramHTML() missing expected content:\nwant substring: %q\ngot length: %d",
						want,
						len(result),
					)
				}
			}

			if !strings.HasPrefix(result, "<!doctype html>") {
				t.Error("rendered HTML should start with doctype")
			}

			if !strings.Contains(result, "</html>") {
				t.Error("rendered HTML should end with closing html tag")
			}
		})
	}
}
