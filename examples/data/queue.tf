data "unionai_queue" "example" {
  id = "batch"
}

output "queue_batch" {
  value = data.unionai_queue.example
}
