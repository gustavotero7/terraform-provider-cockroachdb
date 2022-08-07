package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceDatabase(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDatabase,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"cockroachdb_database.test_database", "name", regexp.MustCompile("test_database")),
				),
			},
		},
	})
}

const testAccResourceDatabase = `
resource "cockroachdb_database" "test_database" {
  name = "test_database"
}
`
