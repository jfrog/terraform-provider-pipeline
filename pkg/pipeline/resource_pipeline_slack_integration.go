package pipeline

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-shared/util"
)

func pipelineSlackProjectIntegrationResource() *schema.Resource {

	var slackIntegrationSchema = util.MergeSchema(
		projectIntegrationSchema,
		map[string]*schema.Schema{
			"url": {
				Type:         schema.TypeString,
				Required:     true,
				Sensitive:    true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsURLWithScheme([]string{"slack"})),
				Description:  "url for Slack access",
			},
		},
	)

	var unpackSlackFormValues = func(data *schema.ResourceData) []FormJSONValues {
		d := &util.ResourceData{data}
		var formJSONValues = []FormJSONValues{
			{
				Label: "url",
				Value: d.GetString("url", true),
			},
		}
		return formJSONValues
	}

	var readSlackProjectIntegration = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		log.Printf("[DEBUG] readSlackProjectIntegration")
		_, err := readProjectIntegration(data, m)
		if err != nil {
			return diag.FromErr(err)
		}

		return nil
	}

	var createSlackProjectIntegration = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		log.Printf("[DEBUG] createSlackProjectIntegration")

		slackFormValues := unpackSlackFormValues(data)
		setUniqueIntegrationNameAndId(data, "slackKey", 78)
		err := createProjectIntegration(data, m, slackFormValues)
		if err != nil {
			return diag.FromErr(err)
		}

		return readSlackProjectIntegration(ctx, data, m)
	}

	var updateSlackProjectIntegration = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		log.Printf("[DEBUG] updateSlackProjectIntegration")

		slackFormValues := unpackSlackFormValues(data)
		err := updateProjectIntegration(data, m, slackFormValues)
		if err != nil {
			return diag.FromErr(err)
		}

		return readSlackProjectIntegration(ctx, data, m)
	}

	return &schema.Resource{
		SchemaVersion: 1,
		CreateContext: createSlackProjectIntegration,
		ReadContext:   readSlackProjectIntegration,
		UpdateContext: updateSlackProjectIntegration,
		DeleteContext: deleteProjectIntegration,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema:      slackIntegrationSchema,
		Description: "Provides an Jfrog Pipelines Slack Project Integration resource.",
	}
}
