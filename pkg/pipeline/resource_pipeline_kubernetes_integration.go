package pipeline

import (
	"context"
	"log"

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
			"master_integration_id": {
				Type:        schema.TypeInt,
				Default:     86,
				Optional:    true,
				Description: "The Id of the master integration.",
			},
			"master_integration_name": {
				Type:        schema.TypeString,
				Default:     "kubernetesConfig",
				Optional:    true,
				Description: "The name of the master integration.",
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
		log.Printf("[DEBUG] readKubernetesProjectIntegration")
		_, err := readProjectIntegration(data, m)
		if err != nil {
			return diag.FromErr(err)
		}

		return nil
	}

	var createKubernetesProjectIntegration = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		log.Printf("[DEBUG] createKubernetesProjectIntegration")

		kubernetesFormValues := unpackKubernetesFormValues(data)
		err := createProjectIntegration(data, m, kubernetesFormValues)
		if err != nil {
			return diag.FromErr(err)
		}

		return readKubernetesProjectIntegration(ctx, data, m)
	}

	var updateKubernetesProjectIntegration = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		log.Printf("[DEBUG] updateKubernetesProjectIntegration")

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
