package provider

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccQueueResource exercises create, update and import against a live control plane.
//
// Notes on what this test can and cannot assert:
//   - `clusters = ["*"]` avoids needing real cluster names; the server skips its cluster
//     membership check when the wildcard is used.
//   - `cluster_pool_name` is left unset so it lands on the `default` pool, which the control
//     plane creates on demand. Any other pool would have to exist already.
//   - There is no CheckDestroy: destroying a queue only drains it, so the queue (and its name)
//     survives. Each run therefore uses a random name and leaves one drained queue behind.
func TestAccQueueResource(t *testing.T) {
	suffix := fmt.Sprintf("%06d", rand.Intn(1000000))
	queueName := "tf-acc-queue-" + suffix

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create, relying on the schema defaults for everything optional.
			{
				Config: testAccQueueConfig(queueName, 0, "medium"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unionai_queue.test", "name", queueName),
					resource.TestCheckResourceAttr("unionai_queue.test", "id", queueName),
					resource.TestCheckResourceAttr("unionai_queue.test", "cluster_pool_name", "default"),
					resource.TestCheckResourceAttr("unionai_queue.test", "priority", "medium"),
					resource.TestCheckResourceAttr("unionai_queue.test", "fairness", "round_robin"),
					resource.TestCheckResourceAttr("unionai_queue.test", "run_concurrency", "0"),
					resource.TestCheckResourceAttr("unionai_queue.test", "action_concurrency", "0"),
					resource.TestCheckResourceAttr("unionai_queue.test", "depth", "0"),
					resource.TestCheckResourceAttr("unionai_queue.test", "clusters.#", "1"),
					resource.TestCheckResourceAttr("unionai_queue.test", "state", "active"),
					resource.TestCheckResourceAttrSet("unionai_queue.test", "created_at"),
					resource.TestCheckResourceAttrSet("unionai_queue.test", "updated_at"),
				),
			},
			// Update the spec. UpdateQueue replaces the spec wholesale, so this also proves the
			// untouched fields (notably cluster_pool_name) survive the round trip.
			{
				Config: testAccQueueConfig(queueName, 5, "max"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unionai_queue.test", "run_concurrency", "5"),
					resource.TestCheckResourceAttr("unionai_queue.test", "priority", "max"),
					resource.TestCheckResourceAttr("unionai_queue.test", "cluster_pool_name", "default"),
					resource.TestCheckResourceAttr("unionai_queue.test", "state", "active"),
				),
			},
			// Import only seeds id, so this catches a Read that forgets to populate name.
			{
				ResourceName:      "unionai_queue.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccQueueDataSource reads back a queue created by the resource.
func TestAccQueueDataSource(t *testing.T) {
	suffix := fmt.Sprintf("%06d", rand.Intn(1000000))
	queueName := "tf-acc-queue-ds-" + suffix

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccQueueDataSourceConfig(queueName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.unionai_queue.test", "id", queueName),
					resource.TestCheckResourceAttr("data.unionai_queue.test", "priority", "max"),
					resource.TestCheckResourceAttr("data.unionai_queue.test", "run_concurrency", "3"),
					resource.TestCheckResourceAttr("data.unionai_queue.test", "cluster_pool_name", "default"),
					resource.TestCheckResourceAttr("data.unionai_queue.test", "state", "active"),
				),
			},
		},
	})
}

func testAccQueueConfig(name string, runConcurrency int, priority string) string {
	return fmt.Sprintf(`
resource "unionai_queue" "test" {
  name            = %[1]q
  clusters        = ["*"]
  run_concurrency = %[2]d
  priority        = %[3]q
}
`, name, runConcurrency, priority)
}

func testAccQueueDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "unionai_queue" "test" {
  name            = %[1]q
  clusters        = ["*"]
  run_concurrency = 3
  priority        = "max"
}

data "unionai_queue" "test" {
  id = unionai_queue.test.name
}
`, name)
}
