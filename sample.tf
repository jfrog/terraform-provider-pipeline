terraform {
  required_providers {
    pipeline = {
      source  = "registry.terraform.io/jfrog/pipeline"
      version = "0.1.0"
    }
  }
}

variable "artifactory_url" {
  type = string
  default = "http://localhost:8081"
}

provider "pipeline" {
  url           = "${var.artifactory_url}"
}

data "pipeline_project" "my-project" {
  name = "my-project"
}

resource "pipeline_project_integration" "my-project-integration" {
  name                    = "my-project-integration"
  project_id              = 0
  project                 = ["my-project"]
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

resource "pipeline_node" "my-node" {
  friendly_name       = "my-node"
  project_id          = 0
  node_pool_id        = 0
  is_on_demand        = true
  is_auto_initialized = true
  ip_address          = "10.0.0.1"
  is_swap_enabled     = true
}

resource "pipeline_node_pool" "my-node-pool" {
  name                       = "my-node-pool"
  project_id                 = 0
  number_of_nodes            = 1
  is_on_demand               = true
  architecture               = "x86_64"
  operating_system           = "Ubuntu_18.04"
  node_idle_interval_in_mins = 20
  environments               = ["DEV"]
}
