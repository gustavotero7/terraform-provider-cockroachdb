package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceGrant(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceGrant,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cockroachdb_grant.test_grant", attrPrivileges+".1", "UPDATE"),
				),
			},
		},
	})
}

const testAccResourceGrant = `
resource "cockroachdb_role" "test_role" {
  name = "test_role"
}

resource "cockroachdb_grant" "test_grant" {
  role = cockroachdb_role.test_role.name
  object_type = "table"
  objects = ["*"]
  privileges = ["SELECT", "UPDATE", "DELETE"]
}
`
