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

// Project GET {{ host }}/access/api/v1/projects/{{prjKey}}/
//GET {{ host }}/artifactory/api/repositories/?prjKey={{prjKey}}
type NodePool struct {
	//Project                   string          `json:"project"`
	Name                   string   `json:"name"`
	ProjectId              int      `json:"projectId"`
	NumberOfNodes          int      `json:"numberOfNodes,omitempty"`
	IsOnDemand             bool     `json:"isOnDemand"`
	Architecture           string   `json:"architecture"`
	OperatingSystem        string   `json:"operatingSystem"`
	NodeIdleIntervalInMins int      `json:"nodeIdleIntervalInMins"`
	Environments           []string `json:"environments,omitempty"`
	ID                     int      `json:"id,omitempty"`
}

const nodePoolsUrl = "pipelines/api/v1/nodePools"

func pipelineNodePoolResource() *schema.Resource {

	var nodePoolSchema = map[string]*schema.Schema{
		"name": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			Description:  "The name of the node pool. Should be prefixed with the project key",
		},
		"project_id": {
			Type:         schema.TypeInt,
			Required:     true,
			ValidateFunc: validation.IntAtLeast(0),
			Description:  "Id of the project where the node pool will live.",
		},
		"number_of_nodes": {
			Type:         schema.TypeInt,
			Optional:     true,
			ValidateFunc: validation.IntAtLeast(0),
			Description:  "Max number of nodes available in the pool.",
		},
		"is_on_demand": {
			Type:        schema.TypeBool,
			Required:    true,
			Description: "Set to true for dynamic node pool. Set to false for static node pool.",
		},
		"architecture": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			Description:  "Set the architecture. This is currently limited to x86_64.",
		},
		"operating_system": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			Description:  "Operating systems supported for the selected architecture.",
		},
		"node_idle_interval_in_mins": {
			Type:         schema.TypeInt,
			Optional:     true,
			ValidateFunc: validation.IntAtLeast(0),
			Description:  "Number of minutes a node can be idle before it is destroyed.",
		},
		"environments": {
			Type:     schema.TypeSet,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Description: "In a project, an array of environment names in which this pipeline source will be.",
		},
	}

	var unpackNodePool = func(data *schema.ResourceData) (NodePool, error) {
		d := &util.ResourceData{data}

		nodePool := NodePool{
			ProjectId:              d.GetInt("project_id", false),
			Name:                   d.GetString("name", false),
			NumberOfNodes:          d.GetInt("number_of_nodes", false),
			IsOnDemand:             d.GetBool("is_on_demand", false),
			Architecture:           d.GetString("architecture", false),
			OperatingSystem:        d.GetString("operating_system", false),
			NodeIdleIntervalInMins: d.GetInt("node_idle_interval_in_mins", false),
			Environments:           d.GetList("environments"),
		}
		return nodePool, nil
	}

	var packNodePool = func(d *schema.ResourceData, nodePool NodePool) diag.Diagnostics {
		var errors []error
		setValue := util.MkLens(d)

		errors = setValue("project_id", nodePool.ProjectId)
		errors = append(errors, setValue("name", nodePool.Name)...)
		errors = append(errors, setValue("number_of_nodes", nodePool.NumberOfNodes)...)
		errors = append(errors, setValue("is_on_demand", nodePool.IsOnDemand)...)
		errors = append(errors, setValue("architecture", nodePool.Architecture)...)
		errors = append(errors, setValue("operating_system", nodePool.OperatingSystem)...)
		errors = append(errors, setValue("node_idle_interval_in_mins", nodePool.NodeIdleIntervalInMins)...)
		errors = append(errors, setValue("environments", nodePool.Environments)...)

		if len(errors) > 0 {
			return diag.Errorf("failed to pack node pool %q", errors)
		}

		return nil
	}

	var readNodePool = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		tflog.Debug(ctx, "readNodePool")

		// This one returns an array of nodepools because the api doesn't seem to provide a GET
		// for a single node pool id. instead it's a query value on id.

		nodePools := []NodePool{}
		resp, err := m.(*resty.Client).R().
			SetResult(&nodePools).
			Get(nodePoolsUrl + "?nodePoolIds=" + data.Id())
		tflog.Trace(ctx, fmt.Sprintf("%v", resp))
		if err != nil {
			return diag.FromErr(err)
		}
		return packNodePool(data, nodePools[0])
	}

	var createNodePool = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		tflog.Debug(ctx, "createNodePool")
		tflog.Trace(ctx, fmt.Sprintf("%+v\n", data))

		nodePool, err := unpackNodePool(data)
		if err != nil {
			return diag.FromErr(err)
		}

		resp, err := m.(*resty.Client).R().SetBody(nodePool).Post(nodePoolsUrl)
		tflog.Debug(ctx, fmt.Sprintf("%v", resp))
		if err != nil {
			return diag.FromErr(err)
		}
		var result NodePool
		err = json.Unmarshal(resp.Body(), &result)
		if err != nil {
			return diag.FromErr(err)
		}
		data.SetId(strconv.Itoa(result.ID))

		return readNodePool(ctx, data, m)
	}

	var updateNodePool = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		tflog.Debug(ctx, "updateNodePool")
		tflog.Trace(ctx, fmt.Sprintf("%+v\n", data))

		nodePool, err := unpackNodePool(data)
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = m.(*resty.Client).R().
			SetBody(nodePool).
			Put(nodePoolsUrl + "/" + data.Id())
		if err != nil {
			return diag.FromErr(err)
		}

		return readNodePool(ctx, data, m)
	}

	var deleteNodePool = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		tflog.Debug(ctx, "deleteNodePool")
		tflog.Trace(ctx, fmt.Sprintf("%+v\n", data))

		resp, err := m.(*resty.Client).R().
			Delete(nodePoolsUrl + "/" + data.Id())

		if err != nil && resp.StatusCode() == http.StatusNotFound {
			data.SetId("")
			return diag.FromErr(err)
		}

		return nil
	}

	return &schema.Resource{
		SchemaVersion: 1,
		CreateContext: createNodePool,
		ReadContext:   readNodePool,
		UpdateContext: updateNodePool,
		DeleteContext: deleteNodePool,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema:      nodePoolSchema,
		Description: "Provides an Jfrog Pipelines Node Pool resource.",
	}
}
