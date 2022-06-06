package pipeline

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-shared/util"
)

func pipelineArtifactoryProjectIntegrationResource() *schema.Resource {

	var artifactoryIntegrationSchema = util.MergeSchema(
		projectIntegrationSchema,
		map[string]*schema.Schema{
			"apikey": {
				Type:         schema.TypeString,
				Required:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "ApiKey for Artifactory access",
			},
			"user": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "User for Artifactory access",
			},
			"url": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "URL to Artifactory instance",
			},
			"master_integration_id": {
				Type:        schema.TypeInt,
				Default:     98,
				Optional:    true,
				Description: "The Id of the master integration.",
			},
			"master_integration_name": {
				Type:        schema.TypeString,
				Default:     "artifactory",
				Optional:    true,
				Description: "The name of the master integration.",
			},
		},
	)

	var unpackArtifactoryFormValues = func(data *schema.ResourceData) []FormJSONValues {
		d := &util.ResourceData{data}
		var formJSONValues = []FormJSONValues{
			{
				Label: "apikey",
				Value: d.GetString("apikey", true),
			},
			{
				Label: "url",
				Value: d.GetString("url", false),
			},
			{
				Label: "user",
				Value: d.GetString("user", false),
			},
		}
		return formJSONValues
	}

	var packArtifactoryFormValues = func(d *schema.ResourceData, formJSONValues []FormJSONValues) []error {
		setValue := util.MkLens(d)
		var errors []error
		for _, idx := range formJSONValues {
			if idx.Label == "url" {
				errors = append(errors, setValue("url", idx.Value)...)
			}
			if idx.Label == "user" {
				errors = append(errors, setValue("user", idx.Value)...)
			}

		}
		return errors
	}

	var readArtifactoryProjectIntegration = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		log.Printf("[DEBUG] readArtifactoryProjectIntegration")
		formJsonValues, err := readProjectIntegration(data, m)
		if err != nil {
			return diag.FromErr(err)
		}
		errors := packArtifactoryFormValues(data, formJsonValues)
		if len(errors) > 0 {
			return diag.Errorf("failed to pack artifactory project integration %q", errors)
		}

		return nil
	}

	var createArtifactoryProjectIntegration = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		log.Printf("[DEBUG] createArtifactoryProjectIntegration")

		artifactoryFormValues := unpackArtifactoryFormValues(data)
		err := createProjectIntegration(data, m, artifactoryFormValues)
		if err != nil {
			return diag.FromErr(err)
		}

		return readArtifactoryProjectIntegration(ctx, data, m)
	}

	var updateArtifactoryProjectIntegration = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		log.Printf("[DEBUG] updateArtifactoryProjectIntegration")

		ArtifactoryFormValues := unpackArtifactoryFormValues(data)
		err := updateProjectIntegration(data, m, ArtifactoryFormValues)
		if err != nil {
			return diag.FromErr(err)
		}

		return readArtifactoryProjectIntegration(ctx, data, m)
	}

	return &schema.Resource{
		SchemaVersion: 1,
		CreateContext: createArtifactoryProjectIntegration,
		ReadContext:   readArtifactoryProjectIntegration,
		UpdateContext: updateArtifactoryProjectIntegration,
		DeleteContext: deleteProjectIntegration,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema:      artifactoryIntegrationSchema,
		Description: "Provides an Jfrog Pipelines Artifactory Project Integration resource.",
	}
}
