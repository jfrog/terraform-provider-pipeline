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

type SystemPropertyBag struct {
	Token string `json:"token"`
}

type Node struct {
	//Project                   string          `json:"project"`
	FriendlyName      string            `json:"friendlyName"`
	ProjectId         int               `json:"projectId"`
	NodePoolId        int               `json:"nodePoolId"`
	IsOnDemand        bool              `json:"isOnDemand"`
	IsAutoInitialized bool              `json:"isAutoInitialized"`
	IPAddress         string            `json:"IPAddress,omitempty"`
	IsSwapEnabled     bool              `json:"isSwapEnabled,omitempty"`
	SystemPropertyBag SystemPropertyBag `json:"systemPropertyBag,omitempty"`
	ID                int               `json:"id,omitempty"`
}

const nodesUrl = "pipelines/api/v1/nodes"

func pipelineNodeResource() *schema.Resource {

	var nodeSchema = map[string]*schema.Schema{
		"friendly_name": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringIsNotEmpty,
			Description:  "The name of the node. Should be prefixed with the project key",
		},
		"project_id": {
			Type:         schema.TypeInt,
			Required:     true,
			ValidateFunc: validation.IntAtLeast(0),
			Description:  "Id of the project where the node will live.",
		},
		"node_pool_id": {
			Type:         schema.TypeInt,
			Required:     true,
			ValidateFunc: validation.IntAtLeast(0),
			Description:  "Id of the node pool where the node will live.",
		},
		"is_on_demand": {
			Type:        schema.TypeBool,
			Required:    true,
			Description: "Set to true for dynamic node pool. Set to false for static node pool.",
		},
		"is_auto_initialized": {
			Type:        schema.TypeBool,
			Required:    true,
			Description: "Determine auto or manual initialization.",
		},
		"ip_address": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringIsNotEmpty,
			Description:  "TSet the architecture. This is currently limited to x86_64.",
		},
		"is_swap_enabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Enable/disable the use of swap space to increase the amount of virtual memory available to the node. ",
		},
		"token": {
			Type:      schema.TypeString,
			Computed:  true,
			Sensitive: true,
		},
	}

	var unpackNode = func(data *schema.ResourceData) (Node, error) {
		d := &util.ResourceData{data}

		node := Node{
			FriendlyName:      d.GetString("friendly_name", false),
			ProjectId:         d.GetInt("project_id", false),
			NodePoolId:        d.GetInt("node_pool_id", false),
			IsOnDemand:        d.GetBool("is_on_demand", false),
			IsAutoInitialized: d.GetBool("is_auto_initialized", false),
			IPAddress:         d.GetString("ip_address", false),
			IsSwapEnabled:     d.GetBool("is_swap_enabled", false),
		}
		node.SystemPropertyBag.Token = d.GetString("token", false)
		return node, nil
	}

	var packNode = func(ctx context.Context, d *schema.ResourceData, node Node) diag.Diagnostics {
		var errors []error
		setValue := util.MkLens(d)

		errors = setValue("project_id", node.ProjectId)
		errors = append(errors, setValue("friendly_name", node.FriendlyName)...)
		errors = append(errors, setValue("node_pool_id", node.NodePoolId)...)
		errors = append(errors, setValue("is_on_demand", node.IsOnDemand)...)
		errors = append(errors, setValue("is_auto_initialized", node.IsAutoInitialized)...)
		errors = append(errors, setValue("ip_address", node.IPAddress)...)
		errors = append(errors, setValue("is_swap_enabled", node.IsSwapEnabled)...)
		errors = append(errors, setValue("token", node.SystemPropertyBag.Token)...)
		tflog.Trace(ctx, fmt.Sprintf("token in object", node.SystemPropertyBag.Token))
		if len(errors) > 0 {
			return diag.Errorf("failed to pack node pool %q", errors)
		}

		return nil
	}

	var readNode = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		tflog.Debug(ctx, "readNode")
		node := Node{}
		_, err := m.(*resty.Client).R().
			SetResult(&node).
			Get(nodesUrl + "/" + data.Id())
		if err != nil {
			return diag.FromErr(err)
		}
		tflog.Trace(ctx, fmt.Sprintf("from readNode; Node obj %v", node))
		return packNode(ctx, data, node)
	}

	var createNode = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		tflog.Debug(ctx, "createNode")
		tflog.Trace(ctx, fmt.Sprintf("%+v\n", data))

		node, err := unpackNode(data)
		if err != nil {
			return diag.FromErr(err)
		}
		tflog.Trace(ctx, fmt.Sprintf("node; %+v\n", node))
		resp, err := m.(*resty.Client).R().SetBody(node).Post(nodesUrl)
		if err != nil {
			return diag.FromErr(err)
		}
		var result Node
		err = json.Unmarshal(resp.Body(), &result)
		if err != nil {
			return diag.FromErr(err)
		}
		data.SetId(strconv.Itoa(result.ID))

		return readNode(ctx, data, m)
	}

	var updateNode = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		tflog.Debug(ctx, "updateNode")
		tflog.Trace(ctx, fmt.Sprintf("%+v\n", data))

		node, err := unpackNode(data)
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = m.(*resty.Client).R().
			SetBody(node).
			Put(nodesUrl + "/" + data.Id())
		if err != nil {
			return diag.FromErr(err)
		}

		return readNode(ctx, data, m)
	}

	var deleteNode = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		tflog.Debug(ctx, "deleteNode")
		tflog.Trace(ctx, fmt.Sprintf("%+v\n", data))

		resp, err := m.(*resty.Client).R().
			Delete(nodesUrl + "/" + data.Id())

		if err != nil && resp.StatusCode() == http.StatusNotFound {
			data.SetId("")
			return diag.FromErr(err)
		}

		return nil
	}

	return &schema.Resource{
		SchemaVersion: 1,
		CreateContext: createNode,
		ReadContext:   readNode,
		UpdateContext: updateNode,
		DeleteContext: deleteNode,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema:      nodeSchema,
		Description: "Provides an Jfrog Pipelines Node resource.",
	}
}
