package pipeline

import (
	"context"
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type Project struct {
	Name string `json:"name"`
	Id   int    `json:"id"`
}

const projectsUrl = "pipelines/api/v1/projects?names={projectName}"

func projectDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceProjectRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The name of the project",
			},
		},

		Description: "Gets the project that has an associated Pipelines object, such as an integration, pipeline source or node pool.",
	}
}

func dataSourceProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var projects []Project
	_, err := m.(*resty.Client).R().
		SetResult(&projects).
		SetPathParam("projectName", d.Get("name").(string)).
		Get(projectsUrl)
	if err != nil {
		return diag.FromErr(err)
	}
	return packProjectData(projects[0], d)
}

func packProjectData(project Project, d *schema.ResourceData) diag.Diagnostics {
	d.SetId(strconv.Itoa(project.Id))
	return nil
}
