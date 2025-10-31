// Copyright 2025 variHQ OÜ
// SPDX-License-Identifier: BSD-3-Clause

package ui

import (
	_ "embed"
	"fmt"
	"html/template"
	"strconv"
	"strings"
)

// flagsToDiagram creates a Mermaid.js stateDiagram from the provided parameters.
func flagsToDiagram(
	roles []string,
	usableRoles map[string]struct{},
	refreshMinutes int64,
) string {
	if len(roles) == 0 {
		return ""
	}

	var builder strings.Builder

	builder.WriteString("stateDiagram\n")

	extractRoleName := func(arn string) string {
		if strings.Contains(arn, "/") {
			parts := strings.Split(arn, "/")

			return parts[len(parts)-1]
		}

		parts := strings.Split(arn, ":")

		return parts[len(parts)-1]
	}

	for pos, role := range roles {
		roleName := extractRoleName(role)

		builder.WriteString("    r")
		builder.WriteString(strconv.Itoa(pos))
		builder.WriteString(": ")
		builder.WriteString(roleName)
		builder.WriteString("\n")
	}

	builder.WriteString("    [*] --> r0\n")

	for pos, role := range roles {
		nextIdx := (pos + 1) % len(roles)
		_, isUsable := usableRoles[role]

		transitionMsg := fmt.Sprintf("wait %dmin and jump", refreshMinutes)
		if !isUsable && len(usableRoles) > 0 {
			transitionMsg = "lacks permission so we jump to " + extractRoleName(roles[nextIdx])
		}

		builder.WriteString("    r")
		builder.WriteString(strconv.Itoa(pos))
		builder.WriteString(" --> r")
		builder.WriteString(strconv.Itoa(nextIdx))
		builder.WriteString(": ")
		builder.WriteString(transitionMsg)
		builder.WriteString("\n")
	}

	return builder.String()
}

//go:embed templates/diagram.html
var DiagramTemplate string

//go:embed static/mermaid.min.js.gz
var MermaidScript []byte

// RenderDiagramHTML pre-renders the diagram HTML once at startup for optimal performance.
func RenderDiagramHTML(
	roles []string,
	usableRoles map[string]struct{},
	refreshMinutes int64,
) (string, error) {
	diagram := flagsToDiagram(roles, usableRoles, refreshMinutes)

	tmpl, err := template.New("diagram").Parse(DiagramTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse diagram template: %w", err)
	}

	var htmlContent strings.Builder

	err = tmpl.Execute(&htmlContent, diagram)
	if err != nil {
		return "", fmt.Errorf("failed to execute diagram template: %w", err)
	}

	return htmlContent.String(), nil
}
