# Multi-app In-App Purchases

Defines an In-App Purchase catalog once and applies it to several apps
(production and staging) using `for_each` over a map of bundle IDs and the
[`app-products`](../modules/app-products) module.

Editing the `in_app_purchases` local and running `terraform apply` propagates
the change to **every** app, because all module instances consume the same
definition.

## Why a module instead of one shared product

App Store Connect binds an In-App Purchase to exactly one app and does not allow
sharing a single product record across apps. The "shared product" is therefore
modelled at the Terraform layer: a single definition (`local.in_app_purchases`)
instantiated per app. The `appleappstoreconnect_app` data source resolves each
bundle ID to its app ID.

## Usage

```bash
export APP_STORE_CONNECT_ISSUER_ID=...
export APP_STORE_CONNECT_KEY_ID=...
export APP_STORE_CONNECT_PRIVATE_KEY="$(cat AuthKey_XXXXXXXXXX.p8)"

terraform init
terraform apply
```

Adjust `local.apps` to your real bundle IDs and `local.in_app_purchases` to your
product catalog. The same pattern extends to auto-renewable subscriptions using
the `appleappstoreconnect_subscription_group` / `appleappstoreconnect_subscription`
resources.
