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
				Description:  "The name of the project. Note: this is *not* the project key.",
			},
		},

		Description: "Gets the project that has an associated Pipelines object, such as an integration, pipeline source or node pool.",
	}
}

func dataSourceProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	projectName := d.Get("name").(string)
	var projects []Project
	_, err := m.(*resty.Client).R().
		SetResult(&projects).
		SetPathParam("projectName", projectName).
		Get(projectsUrl)
	if err != nil {
		return diag.FromErr(err)
	}

	if len(projects) == 0 {
		return diag.Errorf("no project found with name '%s'", projectName)
	}

	return packProject(projects[0], d)
}

func packProject(project Project, d *schema.ResourceData) diag.Diagnostics {
	d.SetId(strconv.Itoa(project.Id))
	return nil
}
