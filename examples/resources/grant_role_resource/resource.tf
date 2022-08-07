resource "cockroachdb_role" "test_role" {
  name = "test_role"
}

resource "cockroachdb_role" "test_user" {
  name     = "test_user"
  login    = true
  password = "test_password_a1s2d3f4"
}

resource "cockroachdb_grant_role" "test_user_grant_test_role" {
  role       = cockroachdb_role.test_user.name
  grant_role = cockroachdb_role.test_role.name
}