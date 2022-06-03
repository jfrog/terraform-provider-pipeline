package pipeline

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func pipelineGithubProjectIntegrationResource() *schema.Resource {

	var githubIntegrationSchema = mergeSchema(
		projectIntegrationSchema,
		map[string]*schema.Schema{
			"token": {
				Type:         schema.TypeString,
				Required:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "Token for Github access",
			},

			"url": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "URL to Github instance",
			},
			"master_integration_id": {
				Type:        schema.TypeInt,
				Default:     20,
				Optional:    true,
				Description: "The Id of the master integration.",
			},
			"master_integration_name": {
				Type:        schema.TypeString,
				Default:     "github",
				Optional:    true,
				Description: "The name of the master integration.",
			},
		},
	)

	var unpackGithubFormValues = func(data *schema.ResourceData) []FormJSONValues {
		d := &ResourceData{data}
		var formJSONValues = []FormJSONValues{
			{
				Label: "token",
				Value: d.getString("token"),
			},
			{
				Label: "url",
				Value: d.getString("url"),
			},
		}
		return formJSONValues
	}

	var packGithubFormValues = func(d *schema.ResourceData, formJSONValues []FormJSONValues) []error {
		setValue := mkLens(d)
		var errors []error
		for _, idx := range formJSONValues {
			if idx.Label == "url" {
				errors = append(errors, setValue("url", idx.Value)...)
			}
		}
		return errors
	}

	var readGithubProjectIntegration = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		log.Printf("[DEBUG] readGithubProjectIntegration")
		formJsonValues, err := readProjectIntegration(data, m)
		if err != nil {
			return diag.FromErr(err)
		}
		errors := packGithubFormValues(data, formJsonValues)
		if len(errors) > 0 {
			return diag.Errorf("failed to pack github project integration %q", errors)
		}

		return nil
	}

	var createGithubProjectIntegration = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		log.Printf("[DEBUG] createGithubProjectIntegration")

		githubFormValues := unpackGithubFormValues(data)
		err := createProjectIntegration(data, m, githubFormValues)
		if err != nil {
			return diag.FromErr(err)
		}

		return readGithubProjectIntegration(ctx, data, m)
	}

	var updateGithubProjectIntegration = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		log.Printf("[DEBUG] updateGithubProjectIntegration")

		githubFormValues := unpackGithubFormValues(data)
		err := updateProjectIntegration(data, m, githubFormValues)
		if err != nil {
			return diag.FromErr(err)
		}

		return readGithubProjectIntegration(ctx, data, m)
	}

	return &schema.Resource{
		SchemaVersion: 1,
		CreateContext: createGithubProjectIntegration,
		ReadContext:   readGithubProjectIntegration,
		UpdateContext: updateGithubProjectIntegration,
		DeleteContext: deleteProjectIntegration,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema:      githubIntegrationSchema,
		Description: "Provides an Jfrog Pipelines Github Project Integration resource.",
	}
}
