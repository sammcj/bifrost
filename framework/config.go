package framework

import "github.com/maximhq/bifrost/framework/pricing"

// FrameworkConfig represents the configuration for the framework.
type FrameworkConfig struct {
	Pricing *pricing.Config `json:"pricing,omitempty"`
}
