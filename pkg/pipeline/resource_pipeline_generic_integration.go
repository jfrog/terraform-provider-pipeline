package pipeline

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-shared/util"
)


func pipelineGenericIntegrationResource() *schema.Resource {
	var genericIntegrationSchema = util.MergeSchema(
		projectIntegrationSchema,
		map[string]*schema.Schema{
			"form_json_values": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"label": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Key or label of the input property.",
						},
						"value": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Value of the input property.",
						},
					},
				},
				Description: "Multiple objects with the values for the integration.",
			},
		},
	)

	var unpackFormValues = func(d *schema.ResourceData) []FormJSONValues {
		var formJSONValues []FormJSONValues
		keyValues := d.Get("form_json_values").([]interface{})
		for _, keyValue := range keyValues {
			jsonValue := keyValue.(map[string]interface{})
			formJSONValue := FormJSONValues{
				Label: jsonValue["label"].(string),
				Value: jsonValue["value"].(string),
			}
			formJSONValues = append(formJSONValues, formJSONValue)
		}
		return formJSONValues
	}

	var packFormValues = func(d *schema.ResourceData, formJSONValues []FormJSONValues) []error {
		setValue := util.MkLens(d)
		var keyValues []interface{}
		for _, jsonValue := range formJSONValues {
			keyValue := map[string]interface{}{
				"label": jsonValue.Label,
				"value": jsonValue.Value,
			}
			keyValues = append(keyValues, keyValue)
		}
		errors := setValue("form_json_values", keyValues)
		return errors
	}

	var readGenericProjectIntegration = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		tflog.Debug(ctx, "readGithubProjectIntegration")
		formJsonValues, err := readProjectIntegration(data, m)
		if err != nil {
			return diag.FromErr(err)
		}
		errors := packFormValues(data, formJsonValues)
		if len(errors) > 0 {
			return diag.Errorf("failed to pack github project integration %q", errors)
		}

		return nil
	}

	var createGenericProjectIntegration = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		tflog.Debug(ctx, "createGithubProjectIntegration")

		formValues := unpackFormValues(data)
		err := createProjectIntegration(data, m, formValues)
		if err != nil {
			return diag.FromErr(err)
		}

		return readGenericProjectIntegration(ctx, data, m)
	}

	var updateGenericProjectIntegration = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		tflog.Debug(ctx, "updateGithubProjectIntegration")

		formValues := unpackFormValues(data)
		err := updateProjectIntegration(data, m, formValues)
		if err != nil {
			return diag.FromErr(err)
		}

		return readGenericProjectIntegration(ctx, data, m)
	}

	return &schema.Resource{
		SchemaVersion: 1,
		CreateContext: createGenericProjectIntegration,
		ReadContext:   readGenericProjectIntegration,
		UpdateContext: updateGenericProjectIntegration,
		DeleteContext: deleteProjectIntegration,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema:      genericIntegrationSchema,
		Description: "Provides a generic Jfrog Pipelines Project Integration resource.",
	}
}