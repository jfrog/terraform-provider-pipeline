package pipeline

import (
	"context"
	"testing"

	"github.com/jfrog/terraform-provider-shared/util"
)

const testFormJsonSchemaKey = "form_json_values"
const redactedSecretValue = "********"

func TestPackFormJSONValues(t *testing.T) {
	testCases := map[string]struct {
		input         []FormJSONValues
		existingState []FormJSONValues
		result        map[string]string
	}{
		"no_asterisks": {
			input: []FormJSONValues{{
				Label:     "key_a",
				Value:     "not_a_secret",
				Sensitive: false,
			}, {
				Label:     "key_b",
				Value:     "something_else",
				Sensitive: false,
			},
			},
			existingState: []FormJSONValues{{
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
			input: []FormJSONValues{{
				Label:     "key_a",
				Value:     redactedSecretValue,
				Sensitive: true,
			}, {
				Label:     "key_b",
				Value:     "something_else",
				Sensitive: false,
			},
			},
			existingState: []FormJSONValues{{
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
			input: []FormJSONValues{{
				Label:     "key_a",
				Value:     redactedSecretValue,
				Sensitive: false,
			}, {
				Label:     "key_b",
				Value:     "something_else",
				Sensitive: false,
			},
			},
			existingState: []FormJSONValues{},
			result:        map[string]string{"key_a": redactedSecretValue, "key_b": "something_else"},
		},
	}

	for name, tcase := range testCases {
		t.Run(name, func(t *testing.T) {
			projectSchemaResource := pipelineProjectIntegrationResource()
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

			errs := packFormJSONValues(context.TODO(), schemaData, testFormJsonSchemaKey, tcase.input)
			for _, err2 := range errs {
				t.Errorf("error bubbled from packFormJSONValues: %v", err2)
			}

			resultValues := unpackFormJSONValues(&util.ResourceData{ResourceData: schemaData}, testFormJsonSchemaKey)
			for _, value := range resultValues {
				k := value.Label
				if tcase.result[k] != value.Value {
					t.Errorf("key %s returned %s; expected %s", k, value.Value, tcase.result[k])
				}
			}
		})
	}
}
