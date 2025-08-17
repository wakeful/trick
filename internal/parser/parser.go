// Copyright 2025 variHQ OÜ
// SPDX-License-Identifier: BSD-3-Clause

package parser

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/zclconf/go-cty/cty"
)

func ParseFile(path string) (*Config, error) {
	content, err := os.ReadFile(path) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	conf, err := decode(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	setDefault(conf)

	return conf, nil
}

func (c *Config) ToFlags() (int64, []string, []string, error) {
	if c.SelectProfile == "" {
		return 0, nil, nil, ErrProfileNotSelected
	}

	rolesARNs := make([]string, 0)
	useRoles := make([]string, 0)

	var ttl int64 = 0

	for _, profile := range c.Profiles {
		if profile.Name != c.SelectProfile {
			slog.Debug(
				"skipping profile parsing",
				slog.String("profile", profile.Name),
				slog.String("reason", "different profile selected"),
			)

			continue
		}

		if profile.Chain == nil {
			slog.Debug(
				"skipping profile parsing",
				slog.String("profile", profile.Name),
				slog.String("reason", "no chain defined"),
			)

			continue
		}

		ttl = profile.Chain.TTL

		for _, role := range profile.Chain.UseRoles {
			rolesARNs = append(rolesARNs, role.ARN)
			if role.Skip {
				useRoles = append(useRoles, role.ARN)
			}
		}
	}

	return ttl, rolesARNs, useRoles, nil
}

func setDefault(config *Config) {
	for _, profile := range config.Profiles {
		if profile.Region == "" {
			slog.Debug(
				"setting default region",
				slog.String("profile", profile.Name),
				slog.String("region", "eu-west-1"),
			)
			profile.Region = "eu-west-1"
		}

		if profile.Chain == nil {
			continue
		}

		if profile.Chain.TTL == 0 {
			slog.Debug(
				"setting default ttl",
				slog.String("profile", profile.Name),
				slog.Int("ttl", defaultTLL),
			)
			profile.Chain.TTL = defaultTLL
		}
	}
}

func decode(fileContent []byte) (*Config, error) {
	parser := hclparse.NewParser()

	fileHCL, diag := parser.ParseHCL(fileContent, "co.hcl")
	if diag.HasErrors() {
		return nil, fmt.Errorf("failed to parse config: %w", diag)
	}

	if fileHCL == nil {
		return nil, ErrParseHCL
	}

	content, diag := fileHCL.Body.Content(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{Name: "select_profile", Required: true},
		},
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "profile", LabelNames: []string{"name"}},
		},
	})
	if diag.HasErrors() {
		return nil, fmt.Errorf("failed to parse config: %w", diag)
	}

	if content == nil {
		return nil, ErrSchemaConfig
	}

	profileAttrs := make(map[string]cty.Value)

	for _, block := range content.Blocks {
		name := block.Labels[0]
		profileAttrs[name] = cty.StringVal(name)
	}

	ctx := &hcl.EvalContext{
		Variables: map[string]cty.Value{
			"profile": cty.ObjectVal(profileAttrs),
		},
		Functions: nil,
	}

	config := &Config{} //nolint:exhaustruct

	errDecode := hclsimple.Decode("config.hcl", fileContent, ctx, config)
	if errDecode != nil {
		return nil, fmt.Errorf("failed to decode config: %w", errDecode)
	}

	return config, nil
}
