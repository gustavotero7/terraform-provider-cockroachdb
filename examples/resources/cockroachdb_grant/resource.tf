resource "cockroachdb_role" "test_role" {
  name = "test_role"
}

resource "cockroachdb_grant" "test_role_grant" {
  role           = cockroachdb_role.test_role.name
  object_type    = "table"
  objects        = ["*"]
  privileges     = ["SELECT", "UPDATE"]
  force_recreate = true
}
