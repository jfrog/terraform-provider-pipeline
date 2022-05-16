resource "pipeline_source" "my-pipeline-source" {
  name                   = "my-pipeline-source"
  project_id             = 0
  project_integration_id = 0
  repository_full_name   = "myrepo/docker-sample"
  file_filter            = "pipelines.yml"
  is_multi_branch        = false
  branch                 = "main"
  branch_exclude_pattern = "debug"
  branch_include_pattern = "features"
  environments           = ["DEV"]
  template_id            = 0
}
