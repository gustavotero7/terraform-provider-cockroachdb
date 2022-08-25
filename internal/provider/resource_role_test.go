package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceRole(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRole,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"cockroachdb_role.test_role", "name", regexp.MustCompile("test_role")),
				),
			},
		},
	})
}

const testAccResourceRole = `
resource "cockroachdb_role" "a_role" {
  name = "a_role"
}

resource "cockroachdb_role" "test_role" {
  name = "test_role"
  login = true
  password = "test_password_a1s2d3f4"
}

resource "cockroachdb_role" "other_role" {
  name = "other_test_role"
}
`
