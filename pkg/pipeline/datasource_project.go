package pipeline

import (
	"strconv"

	"github.com/go-resty/resty/v2"
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
		Read: dataSourceProjectRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The name of the project",
			},
		},
	}
}

func dataSourceProjectRead(d *schema.ResourceData, m interface{}) error {
	var projects []Project
	_, err := m.(*resty.Client).R().
		SetResult(&projects).
		SetPathParam("projectName", d.Get("name").(string)).
		Get(projectsUrl)
	if err != nil {
		return err
	}
	return packProject(projects[0], d)
}

func packProject(project Project, d *schema.ResourceData) error {
	d.SetId(strconv.Itoa(project.Id))
	return nil
}
