resource "unionai_queue" "example" {
  name     = "batch"
  clusters = ["*"]

  run_concurrency    = 10
  action_concurrency = 100
  depth              = 1000

  priority = "medium"
  fairness = "round_robin"
}

output "queue_batch" {
  value = unionai_queue.example
}
