package pipeline

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-shared/util"
)

func pipelineGithubProjectIntegrationResource() *schema.Resource {

	var githubIntegrationSchema = util.MergeSchema(
		projectIntegrationSchema,
		map[string]*schema.Schema{
			"token": {
				Type:         schema.TypeString,
				Required:     true,
				Sensitive:    true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
				Description:  "Token for Github access",
			},

			"url": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
				Description:  "URL to Github instance",
			},
		},
	)

	var unpackGithubFormValues = func(data *schema.ResourceData) []FormJSONValues {
		d := &util.ResourceData{data}
		var formJSONValues = []FormJSONValues{
			{
				Label: "token",
				Value: d.GetString("token", true),
			},
			{
				Label: "url",
				Value: d.GetString("url", false),
			},
		}
		return formJSONValues
	}

	var packGithubFormValues = func(d *schema.ResourceData, formJSONValues []FormJSONValues) []error {
		setValue := util.MkLens(d)
		var errors []error
		for _, jsonValue := range formJSONValues {
			if jsonValue.Label == "url" {
				errors = setValue("url", jsonValue.Value)
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
		setUniqueIntegrationNameAndId(data, "github", 20)
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
