package factory

import (
	"embed"
	"fmt"

	"github.com/zeroSal/went-web/security"
)

func SecurityFactory(
	embedFS embed.FS,
) (*security.Security, error) {
	routes, err := security.LoadRoutesConfigFromEmbed(embedFS, "config/routes.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to load routes config: %w", err)
	}

	securityConfig, err := security.LoadSecurityConfigFromEmbed(
		embedFS,
		"config/security.yaml",
	)

	if err != nil {
		return nil, fmt.Errorf("failed to load security config: %w", err)
	}

	sec, err := security.NewSecurityFromConfig(securityConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create security: %w", err)
	}

	sec.SetRoutes(routes)

	return sec, nil
}