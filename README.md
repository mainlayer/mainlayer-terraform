# Terraform Provider for Mainlayer

Manage [Mainlayer](https://mainlayer.fr) resources and plans as infrastructure code.
Mainlayer is the payments and monetisation layer for AI agents — create monetised API
endpoints, define subscription plans, and control billing models without writing payment
logic yourself.

---

## Requirements

| Tool      | Version  |
|-----------|----------|
| Terraform | >= 1.5   |
| Go        | >= 1.21  |

---

## Installation

### From the Terraform Registry

Add the provider to your `terraform` block:

```hcl
terraform {
  required_providers {
    mainlayer = {
      source  = "mainlayer/mainlayer"
      version = "~> 1.0"
    }
  }
  required_version = ">= 1.5"
}
```

Run `terraform init` to download the provider.

---

## Provider Configuration

```hcl
provider "mainlayer" {
  api_key = var.mainlayer_api_key   # or set MAINLAYER_API_KEY
}
```

| Attribute  | Type   | Required | Description |
|------------|--------|----------|-------------|
| `api_key`  | string | yes      | Your Mainlayer API key. Can also be set via the `MAINLAYER_API_KEY` environment variable. |
| `base_url` | string | no       | Override the API base URL. Defaults to `https://api.mainlayer.fr`. Useful for testing. |

**Recommended**: store your API key as a secret and pass it via an environment variable:

```bash
export MAINLAYER_API_KEY="ml_live_..."
terraform plan
```

---

## Resources

### `mainlayer_resource`

Represents a monetised API endpoint, tool, model, or dataset on the Mainlayer platform.

#### Example

```hcl
resource "mainlayer_resource" "my_api" {
  slug         = "my-api"
  type         = "api"
  price_usdc   = 1.00
  fee_model    = "pay_per_call"
  description  = "My API — powered by Mainlayer"
  callback_url = "https://api.example.com/callback"
}
```

#### Argument Reference

| Argument      | Type   | Required | Description |
|---------------|--------|----------|-------------|
| `slug`        | string | yes      | URL-safe identifier for the resource. Must be unique within your account. |
| `type`        | string | yes      | Resource type. Common values: `api`, `tool`, `model`, `dataset`. |
| `price_usdc`  | number | yes      | Price per call in USD (e.g. `0.01` = $0.01 per call). |
| `fee_model`   | string | yes      | Billing model: `pay_per_call`, `subscription`, or `free`. |
| `description` | string | no       | Human-readable description of the resource. |
| `callback_url`| string | no       | HTTPS URL Mainlayer forwards requests to after processing payment. |

#### Attributes Reference

| Attribute    | Description |
|--------------|-------------|
| `id`         | Unique resource ID assigned by Mainlayer. |
| `created_at` | RFC3339 timestamp of resource creation. |
| `updated_at` | RFC3339 timestamp of last update. |

#### Import

Import an existing resource by its Mainlayer ID:

```bash
terraform import mainlayer_resource.my_api res_abc123
```

---

### `mainlayer_plan`

A subscription plan attached to a `mainlayer_resource`. Plans let you offer tiered
pricing with different call limits and billing periods.

#### Example

```hcl
resource "mainlayer_plan" "starter" {
  resource_id = mainlayer_resource.my_api.id
  name        = "Starter"
  description = "Up to 1,000 calls per month."
  price_usdc  = 9.99
  call_limit  = 1000
  period      = "monthly"
}
```

#### Argument Reference

| Argument      | Type   | Required | Description |
|---------------|--------|----------|-------------|
| `resource_id` | string | yes      | ID of the `mainlayer_resource` this plan belongs to. Forces replacement on change. |
| `name`        | string | yes      | Display name of the plan (e.g. `Starter`, `Pro`). |
| `price_usdc`  | number | yes      | Recurring subscription price in USD per billing period. |
| `description` | string | no       | Human-readable description of what the plan includes. |
| `call_limit`  | number | no       | Maximum API calls per billing period. `0` means unlimited. |
| `period`      | string | no       | Billing period: `monthly` or `yearly`. Defaults to `monthly`. |

#### Attributes Reference

| Attribute    | Description |
|--------------|-------------|
| `id`         | Unique plan ID assigned by Mainlayer. |
| `created_at` | RFC3339 timestamp of plan creation. |
| `updated_at` | RFC3339 timestamp of last update. |

#### Import

Import an existing plan using the `<resource_id>/<plan_id>` format:

```bash
terraform import mainlayer_plan.starter res_abc123/plan_xyz456
```

---

## Data Sources

### `mainlayer_resources`

Lists all Mainlayer resources associated with the authenticated API key.

#### Example

```hcl
data "mainlayer_resources" "all" {}

output "resource_count" {
  value = length(data.mainlayer_resources.all.resources)
}
```

#### Attributes Reference

| Attribute   | Description |
|-------------|-------------|
| `id`        | Always `mainlayer_resources`. |
| `resources` | List of resource objects. Each object contains: `id`, `slug`, `type`, `price_usdc`, `fee_model`, `description`, `callback_url`, `created_at`, `updated_at`. |

---

## Complete Example: Resource with Tiered Plans

```hcl
provider "mainlayer" {}

resource "mainlayer_resource" "analytics_api" {
  slug         = "analytics-api"
  type         = "api"
  price_usdc   = 0.00
  fee_model    = "subscription"
  description  = "Analytics API with tiered subscription plans"
  callback_url = "https://analytics.example.com/mainlayer/callback"
}

resource "mainlayer_plan" "starter" {
  resource_id = mainlayer_resource.analytics_api.id
  name        = "Starter"
  price_usdc  = 9.99
  call_limit  = 1000
  period      = "monthly"
}

resource "mainlayer_plan" "pro" {
  resource_id = mainlayer_resource.analytics_api.id
  name        = "Pro"
  price_usdc  = 49.00
  call_limit  = 25000
  period      = "monthly"
}

resource "mainlayer_plan" "enterprise" {
  resource_id = mainlayer_resource.analytics_api.id
  name        = "Enterprise"
  price_usdc  = 299.00
  call_limit  = 0       # unlimited
  period      = "monthly"
}
```

---

## Development

### Building the provider

```bash
go build ./...
```

### Running unit tests

```bash
go test ./...
```

### Running acceptance tests

Acceptance tests make real API calls and require a valid API key:

```bash
export MAINLAYER_API_KEY="ml_test_..."
export TF_ACC=1
go test -v -timeout 30m ./tests/...
```

### Linting

```bash
go vet ./...
```

---

## License

Apache 2.0. See [LICENSE](LICENSE).
