# Mainlayer Terraform Provider

The Mainlayer Terraform provider lets you manage [Mainlayer](https://mainlayer.fr) resources and subscription plans as infrastructure code.

Use it to:

- Create and manage monetised API endpoints as `mainlayer_resource` resources
- Define subscription plans with usage limits as `mainlayer_plan` resources
- Query all existing resources as data sources

## Example Usage

```hcl
terraform {
  required_providers {
    mainlayer = {
      source  = "mainlayer/mainlayer"
      version = "~> 1.0"
    }
  }
}

provider "mainlayer" {
  # api_key can also be set via MAINLAYER_API_KEY environment variable
  api_key = var.mainlayer_api_key
}

resource "mainlayer_resource" "weather_api" {
  slug         = "weather-api"
  type         = "api"
  price_usdc   = 0.001
  fee_model    = "pay_per_call"
  description  = "Real-time weather data for AI agents"
  callback_url = "https://api.example.com/mainlayer/callback"
}

resource "mainlayer_plan" "pro" {
  resource_id  = mainlayer_resource.weather_api.id
  name         = "Pro"
  description  = "10,000 calls per month"
  price_usdc   = 9.99
  call_limit   = 10000
  period       = "monthly"
}
```

## Authentication

The provider requires a Mainlayer API key. You can set it in three ways:

1. **Provider block** (not recommended — shows in state):
   ```hcl
   provider "mainlayer" {
     api_key = "ml_live_..."
   }
   ```

2. **Environment variable** (recommended):
   ```bash
   export MAINLAYER_API_KEY="ml_live_..."
   terraform apply
   ```

3. **Terraform variable** (recommended for CI):
   ```bash
   export TF_VAR_mainlayer_api_key="ml_live_..."
   terraform apply
   ```

## Schema

### Provider Arguments

| Argument | Type | Required | Description |
|----------|------|----------|-------------|
| `api_key` | string | yes* | Mainlayer API key. *Can be set via `MAINLAYER_API_KEY` env var instead. |
| `base_url` | string | no | Override the API base URL. Default: `https://api.mainlayer.fr` |

### Resources

- [`mainlayer_resource`](resources/resource.md) — A monetised API endpoint or dataset
- [`mainlayer_plan`](resources/plan.md) — A subscription plan for a resource

### Data Sources

- [`mainlayer_resources`](data-sources/resources.md) — Query all resources in your account

## Get Your API Key

Sign up at [mainlayer.fr](https://mainlayer.fr) to get your API key.
