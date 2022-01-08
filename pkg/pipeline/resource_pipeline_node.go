package pipeline

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
		d := &ResourceData{data}

		node := Node{
			FriendlyName:      d.getString("friendly_name"),
			ProjectId:         d.getInt("project_id"),
			NodePoolId:        d.getInt("node_pool_id"),
			IsOnDemand:        d.getBool("is_on_demand"),
			IsAutoInitialized: d.getBool("is_auto_initialized"),
			IPAddress:         d.getString("ip_address"),
			IsSwapEnabled:     d.getBool("is_swap_enabled"),
		}
		node.SystemPropertyBag.Token = d.getString("token")
		return node, nil
	}

	var packNode = func(d *schema.ResourceData, node Node) diag.Diagnostics {
		var errors []error
		setValue := mkLens(d)

		errors = setValue("project_id", node.ProjectId)
		errors = append(errors, setValue("friendly_name", node.FriendlyName)...)
		errors = append(errors, setValue("node_pool_id", node.NodePoolId)...)
		errors = append(errors, setValue("is_on_demand", node.IsOnDemand)...)
		errors = append(errors, setValue("is_auto_initialized", node.IsAutoInitialized)...)
		errors = append(errors, setValue("ip_address", node.IPAddress)...)
		errors = append(errors, setValue("is_swap_enabled", node.IsSwapEnabled)...)
		errors = append(errors, setValue("token", node.SystemPropertyBag.Token)...)
		log.Println("[TRACE] token in object", node.SystemPropertyBag.Token)
		if len(errors) > 0 {
			return diag.Errorf("failed to pack node pool %q", errors)
		}

		return nil
	}

	var readNode = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		log.Printf("[DEBUG] readNode")
		node := Node{}
		_, err := m.(*resty.Client).R().
			SetResult(&node).
			Get(nodesUrl + "/" + data.Id())
		if err != nil {
			return diag.FromErr(err)
		}
		log.Println("[TRACE] from readNode; Node obj", node)
		return packNode(data, node)
	}

	var createNode = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		log.Printf("[DEBUG] createNode")
		log.Printf("[TRACE] %+v\n", data)

		node, err := unpackNode(data)
		if err != nil {
			return diag.FromErr(err)
		}
		log.Printf("[TRACE] node; %+v\n", node)
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
		log.Printf("[DEBUG] updateNode")
		log.Printf("[TRACE] %+v\n", data)

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
		log.Printf("[DEBUG] deleteNode")
		log.Printf("[TRACE] %+v\n", data)

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
