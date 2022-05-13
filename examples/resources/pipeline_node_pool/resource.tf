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
