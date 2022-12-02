package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-shared/util"
)

// ProjectIntegration GET {{ host }}/pipelines/api/v1/projectintegrations/{{projectIntegrationId}}

type ProjectIntegration struct {
	Name                  string           `json:"name"`
	ProjectId             int              `json:"projectId,omitempty"`
	Project               ProjectJSON      `json:"project,omitempty"`
	MasterIntegrationId   int              `json:"masterIntegrationId"`
	MasterIntegrationName string           `json:"masterIntegrationName"`
	FormJSONValues        []FormJSONValues `json:"formJSONValues"`
	Environments          []string         `json:"environments,omitempty"`
	IsInternal            bool             `json:"isInternal,omitempty"`
	ID                    int              `json:"id,omitempty"`
}

type FormJSONValues struct {
	Label     string `json:"label"`
	Value     string `json:"value"`
	Sensitive bool   `json:"-"`
}

func (f FormJSONValues) Id() string {
	return f.Label
}

type ProjectJSON struct {
	Key  string `json:"key,omitempty"`
	Name string `json:"name,omitempty"`
}

const projectIntegrationsUrl = "pipelines/api/v1/projectintegrations"

var baseProjectIntegrationSchema = map[string]*schema.Schema{
	"name": {
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringIsNotEmpty,
		Description:  "The name of the project integration. Should be prefixed with the project key",
	},

	"project_id": {
		Type:         schema.TypeInt,
		Optional:     true,
		Computed:     true,
		ValidateFunc: validation.IntAtLeast(0),
		Description:  "Id of the project.",
	},
	"master_integration_id": {
		Type:         schema.TypeInt,
		Required:     true,
		ValidateFunc: validation.IntAtLeast(0),
		Description:  "The Id of the master integration.",
	},
	"master_integration_name": {
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringIsNotEmpty,
		Description:  "The name of the master integration.",
	},
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
				"is_sensitive": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: "Is the underlying Value sensitive or not",
				},
			},
		},
		Description: "Multiple objects with the values for the integration.",
	},
	"environments": {
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		Description: "In a project, an array of environment names in which this pipeline source will be.",
	},
	"is_internal": {
		Type:        schema.TypeBool,
		Optional:    true,
		Description: "Set this as false to create a Pipelines integration.",
	},
}

func PipelineProjectIntegrationResource() *schema.Resource {

	var projectIntegrationSchemaV1 = util.MergeMaps(
		baseProjectIntegrationSchema,
		map[string]*schema.Schema{
			"project": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "An object containing a project name as an alternative to projectId. The following properties can be set: name, key",
			},
		},
	)

	var projectIntegrationSchemaV2 = util.MergeMaps(
		baseProjectIntegrationSchema,
		map[string]*schema.Schema{
			"project": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
							Description:      "Name of the project",
						},
						"key": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
							Description:      "Key of the project",
						},
					},
				},
				Description: "An object containing a project name as an alternative to projectId.",
			},
		},
	)

	var unpackProject = func(d *util.ResourceData) ProjectJSON {
		var project ProjectJSON
		if v, ok := d.GetOk("project"); ok {
			sets := v.(*schema.Set).List()
			if len(sets) == 0 {
				return project
			}
			value := sets[0].(map[string]interface{})
			project.Key = value["key"].(string)
			project.Name = value["name"].(string)
		}

		return project
	}

	var unpackProjectIntegration = func(data *schema.ResourceData) (ProjectIntegration, error) {
		d := &util.ResourceData{ResourceData: data}

		projectIntegration := ProjectIntegration{
			Name:                  d.GetString("name", false),
			ProjectId:             d.GetInt("project_id", false),
			MasterIntegrationId:   d.GetInt("master_integration_id", false),
			MasterIntegrationName: d.GetString("master_integration_name", false),
			Environments:          d.GetList("environments"),
			IsInternal:            d.GetBool("is_internal", false),
			Project:               unpackProject(d),
			FormJSONValues:        UnpackFormJSONValues(d, "form_json_values"),
		}
		return projectIntegration, nil
	}

	var packProjectIntegration = func(ctx context.Context, d *schema.ResourceData, projectIntegration ProjectIntegration) diag.Diagnostics {
		setValue := util.MkLens(d)

		setValue("name", projectIntegration.Name)
		setValue("project_id", projectIntegration.ProjectId)
		setValue("master_integration_id", projectIntegration.MasterIntegrationId)
		setValue("master_integration_name", projectIntegration.MasterIntegrationName)
		setValue("environments", projectIntegration.Environments)
		setValue("is_internal", projectIntegration.IsInternal)
		errors := PackFormJSONValues(ctx, d, "form_json_values", projectIntegration.FormJSONValues)

		if len(errors) > 0 {
			return diag.Errorf("failed to pack project integration %q", errors)
		}

		return nil
	}

	var readProjectIntegration = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		tflog.Debug(ctx, "readProjectIntegration")
		projectIntegration := ProjectIntegration{}
		resp, err := m.(*resty.Client).R().
			SetResult(&projectIntegration).
			Get(projectIntegrationsUrl + "/" + data.Id())
		tflog.Debug(ctx, fmt.Sprintf("projectIntegration body: %s", string(json.RawMessage(resp.Body()))))
		if err != nil {
			return diag.FromErr(err)
		}
		tflog.Debug(ctx, fmt.Sprintf("projectIntegration Obj: %v", projectIntegration))
		return packProjectIntegration(ctx, data, projectIntegration)
	}

	var createProjectIntegration = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		tflog.Debug(ctx, "createProjectIntegration")
		tflog.Trace(ctx, fmt.Sprintf("%+v\n", data))

		projectIntegration, err := unpackProjectIntegration(data)
		if err != nil {
			return diag.FromErr(err)
		}

		resp, err := m.(*resty.Client).R().SetBody(projectIntegration).Post(projectIntegrationsUrl)
		if err != nil {
			return diag.FromErr(err)
		}
		var result ProjectIntegration
		err = json.Unmarshal(resp.Body(), &result)
		if err != nil {
			return diag.FromErr(err)
		}
		data.SetId(strconv.Itoa(result.ID))

		return readProjectIntegration(ctx, data, m)
	}

	var updateProjectIntegration = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		tflog.Debug(ctx, "updateProjectIntegration")
		tflog.Trace(ctx, fmt.Sprintf("%+v\n", data))

		projectIntegration, err := unpackProjectIntegration(data)
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = m.(*resty.Client).R().
			SetBody(projectIntegration).
			Put(projectIntegrationsUrl + "/" + data.Id())
		if err != nil {
			return diag.FromErr(err)
		}

		return readProjectIntegration(ctx, data, m)
	}

	var deleteProjectIntegration = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		tflog.Debug(ctx, "deleteProjectIntegration")
		tflog.Trace(ctx, fmt.Sprintf("%+v\n", data))

		resp, err := m.(*resty.Client).R().
			Delete(projectIntegrationsUrl + "/" + data.Id())

		if err != nil && resp.StatusCode() == http.StatusNotFound {
			data.SetId("")
			return diag.FromErr(err)
		}

		return nil
	}

	resourceV1 := &schema.Resource{
		Schema: projectIntegrationSchemaV1,
	}

	var resourceStateUpgradeV1 = func(ctx context.Context, rawState map[string]any, meta any) (map[string]any, error) {
		// Convert from a TypeMap (i.e. map[string]any) to TypeSet of 1 item (i.e. []map[string]interface)
		oldProjectState := rawState["project"].(map[string]interface{})
		rawState["project"] = []map[string]interface{}{
			{
				"key":  oldProjectState["key"],
				"name": oldProjectState["name"],
			},
		}

		return rawState, nil
	}

	return &schema.Resource{
		CreateContext: createProjectIntegration,
		ReadContext:   readProjectIntegration,
		UpdateContext: updateProjectIntegration,
		DeleteContext: deleteProjectIntegration,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaVersion: 2,
		Schema:        projectIntegrationSchemaV2,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceV1.CoreConfigSchema().ImpliedType(),
				Upgrade: resourceStateUpgradeV1,
				Version: 1,
			},
		},
		Description: "Provides an Jfrog Pipelines Project Integration resource.",
	}
}

func UnpackFormJSONValues(d *util.ResourceData, key string) []FormJSONValues {
	var formJSONValues []FormJSONValues
	keyValues := d.Get(key).([]interface{})
	for _, keyValue := range keyValues {
		idx := keyValue.(map[string]interface{})
		formJSONValue := FormJSONValues{
			Label:     idx["label"].(string),
			Value:     idx["value"].(string),
			Sensitive: idx["is_sensitive"].(bool),
		}
		formJSONValues = append(formJSONValues, formJSONValue)
	}
	return formJSONValues
}

func PackFormJSONValues(ctx context.Context, d *schema.ResourceData, schemaKey string, formJSONValues []FormJSONValues) []error {
	setValue := util.MkLens(d)
	var keyValues []interface{}
	existingValues := UnpackFormJSONValues(&util.ResourceData{ResourceData: d}, "form_json_values")

	for _, idx := range formJSONValues {
		keyValue := map[string]interface{}{
			"label":        idx.Label,
			"value":        idx.Value,
			"is_sensitive": idx.Sensitive,
		}

		lookup := FindConfigurationById(existingValues, idx.Label)
		// the API will always return the redacted value. Putting this into tf-state will cause a diff every time
		// as it tries to correct "***" -> "secret_val".
		if lookup != nil && lookup.Sensitive {
			if lookup.Value != "" {
				keyValue["value"] = lookup.Value
			}
			//the incoming `idx` value will always be false, the JFrog API has no concept of this field
			keyValue["is_sensitive"] = true
		}

		tflog.Debug(ctx, "packFormJSONValues", keyValue)
		keyValues = append(keyValues, keyValue)
	}
	return setValue(schemaKey, keyValues)
}
