terraform {
  required_providers {
    pipeline = {
      source  = "registry.terraform.io/jfrog/pipeline"
      version = "1.0.2"
    }
    project = {
      source  = "registry.terraform.io/jfrog/project"
      version = "1.1.3"
    }
    artifactory = {
      source  = "registry.terraform.io/jfrog/artifactory"
      version = "6.10.1"
    }
  }
}

variable "artifactory_url" {
  type = string
  default = "http://localhost:8081"
}

provider "pipeline" {
  url = var.artifactory_url
}

provider "project" {
  url = var.artifactory_url
}

provider "artifactory" {
  url = var.artifactory_url
}

# Artifactory resources

resource "artifactory_local_docker_v2_repository" "docker-local" {
  key             = "docker-v2-local"
  tag_retention   = 3
  max_unique_tags = 5
}

resource "artifactory_user" "user-1" {
  name     = "user1"
  email    = "test-user-1@artifactory-terraform.com"
  groups   = ["readers"]
  password = "my super secret password"
}

resource "artifactory_user" "user-2" {
  name     = "user2"
  email    = "test-user-2@artifactory-terraform.com"
  groups   = ["readers"]
  password = "my super secret password"
}

# Projects resources

variable "qa_roles" {
  type    = list(string)
  default = ["READ_REPOSITORY", "READ_RELEASE_BUNDLE", "READ_BUILD", "READ_SOURCES_PIPELINE", "READ_INTEGRATIONS_PIPELINE", "READ_POOLS_PIPELINE", "TRIGGER_PIPELINE"]
}

variable "devop_roles" {
  type    = list(string)
  default = ["READ_REPOSITORY", "ANNOTATE_REPOSITORY", "DEPLOY_CACHE_REPOSITORY", "DELETE_OVERWRITE_REPOSITORY", "TRIGGER_PIPELINE", "READ_INTEGRATIONS_PIPELINE", "READ_POOLS_PIPELINE", "MANAGE_INTEGRATIONS_PIPELINE", "MANAGE_SOURCES_PIPELINE", "MANAGE_POOLS_PIPELINE", "READ_BUILD", "ANNOTATE_BUILD", "DEPLOY_BUILD", "DELETE_BUILD", ]
}

resource "project" "myproject" {
  key          = "myproj"
  display_name = "My Project"
  description  = "My Project"
  admin_privileges {
    manage_members   = true
    manage_resources = true
    index_resources  = true
  }
  max_storage_in_gibibytes   = 10
  block_deployments_on_limit = false
  email_notification         = true

  member {
    name  = artifactory_user.user-1.name
    roles = ["Developer", "Project Admin"]
  }

  member {
    name  = artifactory_user.user-2.name
    roles = ["Developer"]
  }

  group {
    name  = "qa"
    roles = ["qa"]
  }

  group {
    name  = "release"
    roles = ["Release Manager"]
  }

  role {
    name         = "qa"
    description  = "QA role"
    type         = "CUSTOM"
    environments = ["DEV"]
    actions      = var.qa_roles
  }

  role {
    name         = "devop"
    description  = "DevOp role"
    type         = "CUSTOM"
    environments = ["DEV", "PROD"]
    actions      = var.devop_roles
  }

  repos = [artifactory_local_docker_v2_repository.docker-local.key]
}

# Pipelines resources

data "pipeline_project" "my-project" {
  name = project.myproject.key
}

resource "pipeline_project_integration" "my-project-integration" {
  name = "my-project-integration"
  project {
    key  = project.myproject.key
    name = project.myproject.display_name
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
    label        = "label-2"
    value        = "value-2"
    is_sensitive = true
  }
}

resource "pipeline_source" "my-pipeline-source" {
  name                   = "my-pipeline-source"
  project_id             = data.pipeline_project.my-project.id
  project_integration_id = pipeline_project_integration.my-project-integration.id
  repository_full_name   = "myrepo/docker-sample"
  file_filter            = "pipelines.yml"
  is_multi_branch        = false
  branch                 = "main"
  branch_exclude_pattern = "debug"
  branch_include_pattern = "features"
  environments           = ["DEV"]
  template_id            = 0
}

resource "pipeline_node_pool" "my-node-pool" {
  name                       = "my-node-pool"
  project_id                 = data.pipeline_project.my-project.id
  number_of_nodes            = 1
  is_on_demand               = true
  architecture               = "x86_64"
  operating_system           = "Ubuntu_18.04"
  node_idle_interval_in_mins = 20
  environments               = ["DEV"]
}

resource "pipeline_node" "my-node" {
  friendly_name       = "my-node"
  project_id          = data.pipeline_project.my-project.id
  node_pool_id        = pipeline_node_pool.my-node-pool.id
  is_on_demand        = true
  is_auto_initialized = true
  ip_address          = "10.0.0.1"
  is_swap_enabled     = true
}
