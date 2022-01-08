provider "pipeline" {
  url = "${var.artifactory_url}"
  access_token = "${var.artifactory_access_token}"
}
