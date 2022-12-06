resource "pipeline_project_integration" "my-project-integration" {
  name       = "my-project-integration"
  project_id = 0
  project {
    key = "myproj"
    name = "my-project"
  }
  master_integration_id   = 0
  master_integration_name = "my-master-integration"
  environments            = ["DEV"]
  is_internal             = false

  form_json_values {
    label = "label-1"
    value = "value-1"
  }

  form_json_values {
    label = "label-2"
    value = "value-2"
  }
}
