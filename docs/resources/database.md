---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "cockroachdb_database Resource - terraform-provider-cockroachdb"
subcategory: ""
description: |-
  Database in a CockroachDB cluster.
---

# cockroachdb_database (Resource)

Database in a CockroachDB cluster.

## Example Usage

```terraform
resource "cockroachdb_database" "test_database" {
  name = "test_database"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Name of the database.

### Optional

- `owner` (String) Owner of the database.

### Read-Only

- `id` (String) The ID of this resource.


