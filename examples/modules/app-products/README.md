# app-products module

Creates a set of In-App Purchases (with localizations, pricing and territory
availability) in a single app, identified by its bundle ID.

The module is designed to be instantiated once per app with `for_each` while
sharing a single product definition, so the same products can be applied to
several apps (for example production and staging) from one source of truth. See
[`examples/subscriptions-multi-app`](../../subscriptions-multi-app) for the
multi-app wiring.

## Usage

```hcl
module "products" {
  source    = "../modules/app-products"
  bundle_id = "com.example.app"

  in_app_purchases = {
    premium = {
      product_id           = "com.example.app.premium"
      name                 = "Premium"
      in_app_purchase_type = "NON_CONSUMABLE"
      base_territory       = "USA"
      prices               = { USA = "9.99" }
      available_territories = ["USA", "FRA", "GBR"]
      localizations = {
        "en-US" = { name = "Premium", description = "Unlock all features." }
        "fr-FR" = { name = "Premium", description = "Débloquez toutes les fonctionnalités." }
      }
    }
  }
}
```

## Notes

- An In-App Purchase is bound to exactly one app; the same product is created
  independently in each app it is applied to. Apple validates purchases by
  product ID + bundle ID, so the same `product_id` can be reused across apps.
- `customer_price` values must match a price returned by App Store Connect for
  the territory (the module resolves them to price point IDs via the
  `appleappstoreconnect_in_app_purchase_price_point` data source).
