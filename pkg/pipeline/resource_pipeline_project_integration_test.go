package pipeline_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jfrog/terraform-provider-pipeline/pkg/acctest"
	"github.com/jfrog/terraform-provider-pipeline/pkg/pipeline"
	"github.com/jfrog/terraform-provider-shared/test"
	"github.com/jfrog/terraform-provider-shared/util"
)

const testFormJsonSchemaKey = "form_json_values"
const redactedSecretValue = "********"

func TestPackFormJSONValues(t *testing.T) {
	testCases := map[string]struct {
		input         []pipeline.FormJSONValues
		existingState []pipeline.FormJSONValues
		result        map[string]string
	}{
		"no_asterisks": {
			input: []pipeline.FormJSONValues{{
				Label:     "key_a",
				Value:     "not_a_secret",
				Sensitive: false,
			}, {
				Label:     "key_b",
				Value:     "something_else",
				Sensitive: false,
			},
			},
			existingState: []pipeline.FormJSONValues{{
				Label:     "key_a",
				Value:     "not_a_secret",
				Sensitive: false,
			}, {
				Label:     "key_b",
				Value:     "something_else",
				Sensitive: false,
			},
			},
			result: map[string]string{"key_a": "not_a_secret", "key_b": "something_else"},
		},
		"with_asterisks": {
			input: []pipeline.FormJSONValues{{
				Label:     "key_a",
				Value:     redactedSecretValue,
				Sensitive: true,
			}, {
				Label:     "key_b",
				Value:     "something_else",
				Sensitive: false,
			},
			},
			existingState: []pipeline.FormJSONValues{{
				Label:     "key_a",
				Value:     "super_secret",
				Sensitive: true,
			}, {
				Label:     "key_b",
				Value:     "something_else",
				Sensitive: false,
			},
			},
			result: map[string]string{"key_a": "super_secret", "key_b": "something_else"},
		},
		// this technically won't happen, even on resource create; the Input from terraform should have the real value
		// this is just testing code path safety
		"with_asterisks_no_state": {
			input: []pipeline.FormJSONValues{{
				Label:     "key_a",
				Value:     redactedSecretValue,
				Sensitive: false,
			}, {
				Label:     "key_b",
				Value:     "something_else",
				Sensitive: false,
			},
			},
			existingState: []pipeline.FormJSONValues{},
			result:        map[string]string{"key_a": redactedSecretValue, "key_b": "something_else"},
		},
	}

	for name, tcase := range testCases {
		t.Run(name, func(t *testing.T) {
			projectSchemaResource := pipeline.PipelineProjectIntegrationResource()
			schemaData := projectSchemaResource.TestResourceData()
			fJsonState := make([]interface{}, 0)
			for _, idx := range tcase.existingState {
				fJsonState = append(fJsonState, map[string]interface{}{
					"label":        idx.Label,
					"value":        idx.Value,
					"is_sensitive": idx.Sensitive,
				})
			}
			err := schemaData.Set(testFormJsonSchemaKey, fJsonState)
			if err != nil {
				t.Fatalf("error creating test schema %v", err)
			}

			errs := pipeline.PackFormJSONValues(context.TODO(), schemaData, testFormJsonSchemaKey, tcase.input)
			for _, err2 := range errs {
				t.Errorf("error bubbled from packFormJSONValues: %v", err2)
			}

			resultValues := pipeline.UnpackFormJSONValues(&util.ResourceData{ResourceData: schemaData}, testFormJsonSchemaKey)
			for _, value := range resultValues {
				k := value.Label
				if tcase.result[k] != value.Value {
					t.Errorf("key %s returned %s; expected %s", k, value.Value, tcase.result[k])
				}
			}
		})
	}
}

func TestAccProjectIntegration_withProject(t *testing.T) {
	projectKey := fmt.Sprintf("t%d", test.RandomInt())
	_, fqrn, name := test.MkNames(projectKey, "pipeline_project_integration")

	config := util.ExecuteTemplate("TestDatasourceProjectConfig", `
		resource "pipeline_project_integration" "{{ .name }}" {
			name = "{{ .name }}"
			project {
				name = "{{ .projectKey }}"
				key = "{{ .projectKey }}"
			}
			master_integration_id   = 78
			master_integration_name = "slackKey"
			environments            = ["DEV"]

			form_json_values {
				label = "url"
				value = "http://foo.bar"
			}
		}
	`, map[string]interface{}{
		"name":       name,
		"projectKey": projectKey,
	})

	updatedConfig := util.ExecuteTemplate("TestDatasourceProjectConfig", `
		resource "pipeline_project_integration" "{{ .name }}" {
			name = "{{ .name }}"
			project {
				name = "{{ .projectKey }}"
				key = "{{ .projectKey }}"
			}
			master_integration_id   = 78
			master_integration_name = "slackKey"
			environments            = ["PROD"]

			form_json_values {
				label = "url"
				value = "http://fizz.buzz"
			}
		}
	`, map[string]interface{}{
		"name":       name,
		"projectKey": projectKey,
	})

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.CreateProject(t, projectKey)
		},
		CheckDestroy: func(*terraform.State) error {
			acctest.DeleteProject(t, projectKey)
			return nil
		},
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", name),
					resource.TestCheckResourceAttrSet(fqrn, "project_id"),
					resource.TestCheckResourceAttr(fqrn, "master_integration_id", "78"),
					resource.TestCheckResourceAttr(fqrn, "master_integration_name", "slackKey"),
					resource.TestCheckResourceAttr(fqrn, "environments.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "environments.0", "DEV"),
					resource.TestCheckResourceAttr(fqrn, "form_json_values.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(fqrn, "form_json_values.*", map[string]string{
						"label": "url",
						"value": "http://foo.bar",
					}),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", name),
					resource.TestCheckResourceAttrSet(fqrn, "project_id"),
					resource.TestCheckResourceAttr(fqrn, "master_integration_id", "78"),
					resource.TestCheckResourceAttr(fqrn, "master_integration_name", "slackKey"),
					resource.TestCheckResourceAttr(fqrn, "environments.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "environments.0", "PROD"),
					resource.TestCheckResourceAttr(fqrn, "form_json_values.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(fqrn, "form_json_values.*", map[string]string{
						"label": "url",
						"value": "http://fizz.buzz",
					}),
				),
			},
		},
	})
}

func TestAccProjectIntegration_withProjectId(t *testing.T) {
	var integrationId int

	projectKey := fmt.Sprintf("t%d", test.RandomInt())
	integrationName := fmt.Sprintf("int%d", test.RandomInt())
	_, fqrn, name := test.MkNames(projectKey, "pipeline_project_integration")

	config := util.ExecuteTemplate("TestDatasourceProjectConfig", `
		data "pipeline_project" "{{ .projectKey }}" {
			name = "{{ .projectKey }}"
		}

		resource "pipeline_project_integration" "{{ .name }}" {
			name = "{{ .name }}"
			project_id = data.pipeline_project.{{ .projectKey }}.id
			master_integration_id   = 78
			master_integration_name = "slackKey"
			environments            = ["DEV"]

			form_json_values {
				label = "url"
				value = "http://foo.bar"
			}
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
					resource.TestCheckResourceAttr(fqrn, "name", name),
					resource.TestCheckResourceAttrSet(fqrn, "project_id"),
					resource.TestCheckResourceAttr(fqrn, "project.#", "0"),
					resource.TestCheckResourceAttr(fqrn, "master_integration_id", "78"),
					resource.TestCheckResourceAttr(fqrn, "master_integration_name", "slackKey"),
					resource.TestCheckResourceAttr(fqrn, "environments.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "environments.0", "DEV"),
					resource.TestCheckResourceAttr(fqrn, "form_json_values.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(fqrn, "form_json_values.*", map[string]string{
						"label": "url",
						"value": "http://foo.bar",
					}),
				),
			},
		},
	})
}
