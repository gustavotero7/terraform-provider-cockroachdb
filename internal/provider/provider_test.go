package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// providerFactories are used to instantiate a provider during acceptance testing.
// The factory function will be invoked for every Terraform CLI command executed
// to create a provider server to which the CLI can reattach.
var providerFactories = map[string]func() (*schema.Provider, error){
	"cockroachdb": func() (*schema.Provider, error) {
		p := New("dev")()
		p.Configure(context.Background(), terraform.NewResourceConfigRaw(map[string]interface{}{
			"host":     os.Getenv("TEST_COCKROACHDB_HOST"),
			"port":     os.Getenv("TEST_COCKROACHDB_PORT"),
			"username": os.Getenv("TEST_COCKROACHDB_USERNAME"),
			"password": os.Getenv("TEST_COCKROACHDB_PASSWORD"),
			"database": os.Getenv("TEST_COCKROACHDB_DATABASE"),
			"cluster":  os.Getenv("TEST_COCKROACHDB_CLUSTER"),
		}))
		return p, nil
	},
}

func TestProvider(t *testing.T) {
	p := New("dev")()
	p.Configure(context.Background(), terraform.NewResourceConfigRaw(map[string]interface{}{
		"host":     os.Getenv("TEST_COCKROACHDB_HOST"),
		"port":     os.Getenv("TEST_COCKROACHDB_PORT"),
		"username": os.Getenv("TEST_COCKROACHDB_USERNAME"),
		"password": os.Getenv("TEST_COCKROACHDB_PASSWORD"),
		"database": os.Getenv("TEST_COCKROACHDB_DATABASE"),
		"cluster":  os.Getenv("TEST_COCKROACHDB_CLUSTER"),
	}))
	if err := p.InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func testAccPreCheck(t *testing.T) {
	// Do not run tests if a dsn is not set.
	if os.Getenv("TEST_COCKROACHDB_HOST") == "" ||
		os.Getenv("TEST_COCKROACHDB_PORT") == "" ||
		os.Getenv("TEST_COCKROACHDB_USERNAME") == "" ||
		os.Getenv("TEST_COCKROACHDB_PASSWORD") == "" ||
		os.Getenv("TEST_COCKROACHDB_DATABASE") == "" ||
		os.Getenv("TEST_COCKROACHDB_CLUSTER") == "" {
		t.Log("Skipping CockroachDB tests, no dsn set")
		t.SkipNow()
	}
}
