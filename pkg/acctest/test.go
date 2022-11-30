package acctest

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jfrog/terraform-provider-pipeline/pkg/pipeline"
	"github.com/jfrog/terraform-provider-shared/client"
	"github.com/jfrog/terraform-provider-shared/test"
)

// Provider PreCheck(t) must be called before using this provider instance.
var Provider *schema.Provider
var ProviderFactories map[string]func() (*schema.Provider, error)

// testAccProviderConfigure ensures Provider is only configured once
//
// The PreCheck(t) function is invoked for every test and this prevents
// extraneous reconfiguration to the same values each time. However, this does
// not prevent reconfiguration that may happen should the address of
// Provider be errantly reused in ProviderFactories.
var testAccProviderConfigure sync.Once

func init() {
	Provider = pipeline.Provider()

	ProviderFactories = map[string]func() (*schema.Provider, error){
		"pipeline": func() (*schema.Provider, error) { return pipeline.Provider(), nil },
	}
}

// PreCheck This function should be present in every acceptance test.
func PreCheck(t *testing.T) {
	// Since we are outside the scope of the Terraform configuration we must
	// call Configure() to properly initialize the provider configuration.
	testAccProviderConfigure.Do(func() {
		restyClient := GetTestResty(t)

		artifactoryUrl := GetArtifactoryUrl(t)
		// Set custom base URL so repos that relies on it will work
		// https://www.jfrog.com/confluence/display/JFROG/Artifactory+REST+API#ArtifactoryRESTAPI-UpdateCustomURLBase
		_, err := restyClient.R().
			SetBody(artifactoryUrl).
			SetHeader("Content-Type", "text/plain").
			Put("/artifactory/api/system/configuration/baseUrl")
		if err != nil {
			t.Fatalf("Failed to set custom base URL: %v", err)
		}

		configErr := Provider.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
		if configErr != nil {
			t.Fatalf("Failed to configure provider %v", configErr)
		}
	})
}

func GetArtifactoryUrl(t *testing.T) string {
	return test.GetEnvVarWithFallback(t, "PIPELINES_URL", "JFROG_URL")
}

func GetTestResty(t *testing.T) *resty.Client {
	artifactoryUrl := GetArtifactoryUrl(t)
	restyClient, err := client.Build(artifactoryUrl, "")
	if err != nil {
		t.Fatal(err)
	}

	accessToken := test.GetEnvVarWithFallback(t, "PIPELINES_ACCESS_TOKEN", "JFROG_ACCESS_TOKEN")
	restyClient, err = client.AddAuth(restyClient, "", accessToken)
	if err != nil {
		t.Fatal(err)
	}
	return restyClient
}

func CreateProject(t *testing.T, projectKey string) {
	type AdminPrivileges struct {
		ManageMembers   bool `json:"manage_members"`
		ManageResources bool `json:"manage_resources"`
		IndexResources  bool `json:"index_resources"`
	}

	type Project struct {
		Key             string          `json:"project_key"`
		DisplayName     string          `json:"display_name"`
		Description     string          `json:"description"`
		AdminPrivileges AdminPrivileges `json:"admin_privileges"`
	}

	restyClient := GetTestResty(t)

	project := Project{
		Key:         projectKey,
		DisplayName: projectKey,
		Description: fmt.Sprintf("%s description", projectKey),
		AdminPrivileges: AdminPrivileges{
			ManageMembers:   true,
			ManageResources: true,
			IndexResources:  true,
		},
	}

	_, err := restyClient.R().
		SetBody(project).
		Post("/access/api/v1/projects")
	if err != nil {
		t.Fatal(err)
	}
}

func DeleteProject(t *testing.T, projectKey string) {
	restyClient := GetTestResty(t)
	_, err := restyClient.R().Delete("/access/api/v1/projects/" + projectKey)
	if err != nil {
		t.Fatal(err)
	}
}

func CreateProjectIntegration(t *testing.T, name string, projectKey string) int {
	restyClient := GetTestResty(t)

	project := pipeline.ProjectIntegration{
		Name: name,
		Project: pipeline.ProjectJSON{
			Name: projectKey,
			Key:  projectKey,
		},
		MasterIntegrationId:   78,
		MasterIntegrationName: "slackKey",
		FormJSONValues: []pipeline.FormJSONValues{
			pipeline.FormJSONValues{
				Label: "url",
				Value: "http://foo.bar",
			},
		},
	}

	_, err := restyClient.R().
		SetBody(project).
		SetResult(&project).
		Post("/pipelines/api/v1/projectIntegrations")
	if err != nil {
		t.Fatal(err)
	}

	return project.ID
}

func DeleteProjectIntegration(t *testing.T, id int) {
	restyClient := GetTestResty(t)
	_, err := restyClient.R().Delete(fmt.Sprintf("/pipelines/api/v1/projectIntegrations/%d", id))
	if err != nil {
		t.Fatal(err)
	}
}
