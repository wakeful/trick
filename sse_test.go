// Copyright 2025 variHQ OÜ
// SPDX-License-Identifier: BSD-3-Clause

package main

import "testing"

func TestMainHandler(t *testing.T) {
	t.Parallel()

	preRenderedHTML := "<html><body>Test Content</body></html>"
	handler := mainHandler(preRenderedHTML)

	if handler == nil {
		t.Fatal("diagramHandler returned nil")
	}
}
