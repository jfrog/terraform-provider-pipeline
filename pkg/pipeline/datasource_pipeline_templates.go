package pipeline

import (
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type PipelineTemplate struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	Namespace        string `json:"namespace"`
	SyntaxVersion    string `json:"syntaxVersion"`
	TemplateSourceId int    `json:"templateSourceId"`
	LatestSha        string `json:"latestSha"`
}

const pipelineTemplatesUrl = "pipelines/api/v1/templates"

func pipelineTemplatesDataSource() *schema.Resource {
	return &schema.Resource{
		Read: dataSourcePipelineTemplatesRead,

		Schema: map[string]*schema.Schema{
			"templates": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"namespace": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"syntax_version": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"template_source_id": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},
						"latest_sha": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourcePipelineTemplatesRead(d *schema.ResourceData, m interface{}) error {
	templates := []PipelineTemplate{}
	_, err := m.(*resty.Client).R().
		SetResult(&templates).
		Get(pipelineTemplatesUrl)
	if err != nil {
		return err
	}

	pipelineTemplates := make([]map[string]interface{}, 0)

	for _, v := range templates {
		template := make(map[string]interface{})

		template["id"] = v.ID
		template["name"] = v.Name
		template["namespace"] = v.Namespace
		template["syntax_version"] = v.SyntaxVersion
		template["template_source_id"] = v.TemplateSourceId
		template["latest_sha"] = v.LatestSha

		pipelineTemplates = append(pipelineTemplates, template)
	}

	if err := d.Set("templates", pipelineTemplates); err != nil {
		return err
	}

	// always run
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return nil
}
