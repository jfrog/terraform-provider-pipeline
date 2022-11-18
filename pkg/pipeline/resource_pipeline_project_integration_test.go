package pipeline

import (
	"context"
	"testing"

	"github.com/jfrog/terraform-provider-shared/util"
)

const testFormJsonSchemaKey = "form_json_values"

func TestPackFormJSONValues(t *testing.T) {
	testCases := map[string]struct {
		input         map[string]string
		existingState map[string]string
		result        map[string]string
	}{
		"no_asterisks": {
			input:         map[string]string{"key_a": "not_a_secret", "key_b": "something_else"},
			existingState: map[string]string{"key_a": "not_a_secret", "key_b": "something_else"},
			result:        map[string]string{"key_a": "not_a_secret", "key_b": "something_else"},
		},
		"with_asterisks": {
			input:         map[string]string{"key_a": redactedSecretValue, "key_b": "something_else"},
			existingState: map[string]string{"key_a": "super_secret", "key_b": "something_else"},
			result:        map[string]string{"key_a": "super_secret", "key_b": "something_else"},
		},
		// this technically won't happen, even on resource create; the Input from terraform should have the real value
		// this is just testing code path safety
		"with_asterisks_no_state": {
			input:         map[string]string{"key_a": redactedSecretValue, "key_b": "something_else"},
			existingState: map[string]string{},
			result:        map[string]string{"key_a": redactedSecretValue, "key_b": "something_else"},
		},
	}

	for name, tcase := range testCases {
		t.Run(name, func(t *testing.T) {
			projectSchemaResource := pipelineProjectIntegrationResource()
			schemaData := projectSchemaResource.TestResourceData()
			fJsonState := make([]interface{}, 0)
			for k, v := range tcase.existingState {
				fJsonState = append(fJsonState, map[string]interface{}{
					"label": k,
					"value": v,
				})
			}
			err := schemaData.Set(testFormJsonSchemaKey, fJsonState)
			if err != nil {
				t.Fatalf("error creating test schema %v", err)
			}

			fJsonInput := make([]FormJSONValues, 0)
			for k, v := range tcase.input {
				fJsonInput = append(fJsonInput, FormJSONValues{Label: k, Value: v})
			}

			errs := packFormJSONValues(context.TODO(), schemaData, testFormJsonSchemaKey, fJsonInput)
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
