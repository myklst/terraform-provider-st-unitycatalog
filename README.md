# terraform-provider-st-unitycatalog

This custom Terraform provider was built for our specific use case.

A Terraform provider for [Unity Catalog](https://www.unitycatalog.io/), managing
resources through the API.

## Supported Versions

| Terraform version | minimum provider version | maxmimum provider version
| ---- | ---- | ----|
| >= 1.13.x	| 0.1.1	| latest |

## Requirements

-	[Terraform](https://www.terraform.io/downloads.html) 1.13.x
-	[Go](https://golang.org/doc/install) 1.26 (to build the provider plugin)

## Provider Configuration

The provider authenticates to the Unity Catalog control plane using either a
bearer token (`Authorization: Bearer <token>`) or a username and password (HTTP
Basic authentication). When neither is configured, no authentication is applied
(Unity Catalog may be run with authentication disabled). When a token is
supplied it takes precedence over username/password.

```terraform
provider "st-unitycatalog" {
  host     = "https://unity.example.com"
  username = "admin"
  password = "your-password"
}
```

or, using a token:

```terraform
provider "st-unitycatalog" {
  host  = "https://unity.example.com"
  token = "your-token"
}
```

Credentials may also be provided via environment variables:

| Attribute   | Environment variable    |
|-------------|-------------------------|
| `host`      | `UNITYCATALOG_HOST`     |
| `username`  | `UNITYCATALOG_USERNAME` |
| `password`  | `UNITYCATALOG_PASSWORD` |
| `token`     | `UNITYCATALOG_TOKEN`    |

## Local Installation

1. Run make file `make install-local-custom-provider` to install the provider under ~/.terraform.d/plugins.

2. The provider source should be change to the path that configured in the *Makefile*:

    ```
    terraform {
      required_providers {
        st-unitycatalog = {
          source = "example.local/myklst/st-unitycatalog"
        }
      }
    }

    provider "st-unitycatalog" {
      host = "http://localhost:8080/unitycatalog"
    }
    ```

## Why Custom Provider

At the time this custom provider was developed, Unity Catalog did not have an official Terraform provider.

## Resources

* [`st-unitycatalog_user`](docs/resources/user.md) - Manages a Unity Catalog user.
* [`st-unitycatalog_permissions`](docs/resources/permissions.md) - Manages privileges for a principal on a Unity Catalog securable.

## Data Sources

* [`st-unitycatalog_tables`](docs/data-sources/tables.md) - Lists tables in a Unity Catalog catalog.
