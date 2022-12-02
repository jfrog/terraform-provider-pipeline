package pipeline_test

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jfrog/terraform-provider-pipeline/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/test"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccDatasourceProject(t *testing.T) {
	var integrationId int

	rand.Seed(time.Now().UnixNano())
	projectKey := fmt.Sprintf("t%d", test.RandomInt())
	integrationName := fmt.Sprintf("int%d", test.RandomInt())
	_, fqrn, name := test.MkNames(projectKey, "pipeline_project")
	dataSourceName := fmt.Sprintf("data.%s", fqrn)

	config := util.ExecuteTemplate("TestDatasourceProjectConfig", `
		data "pipeline_project" "{{ .name }}" {
			name = "{{ .projectKey }}"
		}
	`, map[string]interface{}{
		"name":       name,
		"projectKey": projectKey,
	})

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.CreateProject(t, projectKey)
			integrationId = acctest.CreateProjectIntegration(t, integrationName, projectKey)
		},
		CheckDestroy: func(*terraform.State) error {
			acctest.DeleteProjectIntegration(t, integrationId)
			acctest.DeleteProject(t, projectKey)
			return nil
		},
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", projectKey),
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
				),
			},
		},
	})
}

func TestDatasourceProject_notFound(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	projectKey := fmt.Sprintf("t%d", test.RandomInt())
	_, _, name := test.MkNames(projectKey, "pipeline_project")

	config := util.ExecuteTemplate("TestDatasourceProjectConfig", `
		data "pipeline_project" "{{ .name }}" {
			name = "{{ .projectKey }}"
		}
	`, map[string]interface{}{
		"name":       name,
		"projectKey": projectKey,
	})

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
		},
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(fmt.Sprintf(".*no project found with name '%s'", projectKey)),
			},
		},
	})
}
