package main

import (
	"github.com/PatchSimple/terraform-provider-pipeline/pkg/pipeline"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

// Run the docs generation tool, check its repository for more information on how it works and how docs
// can be customized.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: pipeline.Provider,
	})
}
