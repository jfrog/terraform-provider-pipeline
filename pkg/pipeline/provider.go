package pipeline

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-shared/client"
	"github.com/jfrog/terraform-provider-shared/util"
)

var Version = "0.0.1"
var productId = "terraform-provider-pipeline/" + Version

func Provider() *schema.Provider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"url": {
				Type:         schema.TypeString,
				Required:     true,
				DefaultFunc:  schema.MultiEnvDefaultFunc([]string{"PIPELINES_URL", "JFROG_URL"}, "http://localhost:8081"),
				ValidateFunc: validation.IsURLWithHTTPorHTTPS,
				Description:  "URL of Artifactory. This can also be sourced from the `PIPELINES_URL` or `JFROG_URL` environment variable. Default to 'http://localhost:8081' if not set.",
			},
			"access_token": {
				Type:             schema.TypeString,
				Required:         true,
				Sensitive:        true,
				DefaultFunc:      schema.MultiEnvDefaultFunc([]string{"PIPELINES_ACCESS_TOKEN", "JFROG_ACCESS_TOKEN"}, ""),
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
				Description:      "This is a Bearer token that can be given to you by your admin under `Identity and Access`. This can also be sourced from the `PIPELINES_ACCESS_TOKEN` or `JFROG_ACCESS_TOKEN` environment variable. Defauult to empty string if not set.",
			},
			"check_license": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Toggle for pre-flight checking of Artifactory Enterprise license. Default to `true`.",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"pipeline_source":              pipelineSourceResource(),
			"pipeline_project_integration": pipelineProjectIntegrationResource(),
			"pipeline_node_pool":           pipelineNodePoolResource(),
			"pipeline_node":                pipelineNodeResource(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"pipeline_project": projectDataSource(),
		},
	}

	p.ConfigureContextFunc = func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
		terraformVersion := p.TerraformVersion
		if terraformVersion == "" {
			terraformVersion = "0.13+compatible"
		}
		return providerConfigure(ctx, data, terraformVersion)
	}

	return p
}

func providerConfigure(ctx context.Context, d *schema.ResourceData, terraformVersion string) (interface{}, diag.Diagnostics) {
	URL, ok := d.GetOk("url")
	if URL == nil || URL == "" || !ok {
		return nil, diag.Errorf("you must supply a URL")
	}

	restyBase, err := client.Build(URL.(string), Version)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	accessToken := d.Get("access_token").(string)

	restyBase, err = client.AddAuth(restyBase, "", accessToken)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	checkLicense := d.Get("check_license").(bool)
	if checkLicense {
		licenseErr := util.CheckArtifactoryLicense(restyBase, "Enterprise")
		if licenseErr != nil {
			return nil, licenseErr
		}
	}

	featureUsage := fmt.Sprintf("Terraform/%s", terraformVersion)
	util.SendUsage(ctx, restyBase, productId, featureUsage)

	return restyBase, nil
}
