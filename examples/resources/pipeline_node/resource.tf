resource "pipeline_node" "my-node" {
  friendly_name       = "my-node"
  project_id          = 0
  node_pool_id        = 0
  is_on_demand        = true
  is_auto_initialized = true
  ip_address          = "10.0.0.1"
  is_swap_enabled     = true
}
