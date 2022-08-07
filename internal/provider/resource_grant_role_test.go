package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceGrantRole(t *testing.T) {
	t.SkipNow()

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceGrantRole,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cockroachdb_grant_role.test_grant_role", attrRole, "test_user"),
					resource.TestCheckResourceAttr(
						"cockroachdb_grant_role.test_grant_role", attrGrantRole, "test_role"),
				),
			},
		},
	})
}

const testAccResourceGrantRole = `
resource "cockroachdb_role" "test_role" {
  name = "test_role"
}

resource "cockroachdb_role" "test_user" {
  name = "test_user"
  login = true
  password = "test_password_a1s2d3f4"
}

resource "cockroachdb_grant_role" "test_grant_role" {
  role = cockroachdb_role.test_user.name
  grant_role = cockroachdb_role.test_role.name
}
`
