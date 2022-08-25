# Terraform Provider CockroachDB (Terraform Plugin SDK)

CockroachDB terraform provider with support for databases, grants  (+ grant role) and roles.

## Requirements

-	[Terraform](https://www.terraform.io/downloads.html) >= 0.13.x
-	[Go](https://golang.org/doc/install) >= 1.17

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command: 
```sh
$ go install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Using the provider

1. Init the provider
```terraform
provider "cockroachdb" {
  host     = "remote_host"
  port     = 26257
  username = "test_user"
  password = "a1s2d3f4g5h6j7k8l9"
  database = "test_database"
  cluster  = "test_cluster_123"
}
```
2. Create resources
```terraform
resource "cockroachdb_database" "test_database" {
  name = "test_database"
}

resource "cockroachdb_role" "test_role" {
  name = "test_role"
}

resource "cockroachdb_grant" "test_role_grant" {
  role        = cockroachdb_role.test_role.name
  object_type = "table"
  objects     = [cockroachdb_database.test_database.public.*]
  privileges  = ["SELECT", "UPDATE"]
  force_recreate = true
}

```

See more examples in the **examples** folder

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ export TEST_COCKROACHDB_HOST={your_db_host}
$ export TEST_COCKROACHDB_PORT={your_db_port}
$ export TEST_COCKROACHDB_USERNAME={your_db_username}
$ export TEST_COCKROACHDB_PASSWORD={your_db_password}
$ export TEST_COCKROACHDB_DATABASE={your_db_database}
$ export TEST_COCKROACHDB_CLUSTER={your_db_cluster}
$ make testacc
```
