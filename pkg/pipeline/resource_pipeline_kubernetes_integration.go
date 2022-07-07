package pipeline

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-shared/util"
)

func pipelineKubernetesProjectIntegrationResource() *schema.Resource {

	var kubernetesIntegrationSchema = util.MergeSchema(
		projectIntegrationSchema,
		map[string]*schema.Schema{
			"kubeconfig": {
				Type:         schema.TypeString,
				Required:     true,
				Sensitive:    true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
				Description:  "Token for Kubernetes access",
			},
		},
	)

	var unpackKubernetesFormValues = func(data *schema.ResourceData) []FormJSONValues {
		d := &util.ResourceData{data}
		var formJSONValues = []FormJSONValues{
			{
				Label: "kubeconfig",
				Value: d.GetString("kubeconfig", true),
			},
		}
		return formJSONValues
	}

	var readKubernetesProjectIntegration = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		tflog.Debug(ctx, "readKubernetesProjectIntegration")
		_, err := readProjectIntegration(data, m)
		if err != nil {
			return diag.FromErr(err)
		}

		return nil
	}

	var createKubernetesProjectIntegration = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		tflog.Debug(ctx, "createKubernetesProjectIntegration")

		kubernetesFormValues := unpackKubernetesFormValues(data)
		setUniqueIntegrationNameAndId(data, "kubernetesConfig", 86)
		err := createProjectIntegration(data, m, kubernetesFormValues)
		if err != nil {
			return diag.FromErr(err)
		}

		return readKubernetesProjectIntegration(ctx, data, m)
	}

	var updateKubernetesProjectIntegration = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		tflog.Debug(ctx, "updateKubernetesProjectIntegration")

		kubernetesFormValues := unpackKubernetesFormValues(data)
		err := updateProjectIntegration(data, m, kubernetesFormValues)
		if err != nil {
			return diag.FromErr(err)
		}

		return readKubernetesProjectIntegration(ctx, data, m)
	}

	return &schema.Resource{
		SchemaVersion: 1,
		CreateContext: createKubernetesProjectIntegration,
		ReadContext:   readKubernetesProjectIntegration,
		UpdateContext: updateKubernetesProjectIntegration,
		DeleteContext: deleteProjectIntegration,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema:      kubernetesIntegrationSchema,
		Description: "Provides an Jfrog Pipelines Kubernetes Project Integration resource.",
	}
}
