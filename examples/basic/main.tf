terraform {
  required_providers {
    mainlayer = {
      source  = "mainlayer/mainlayer"
      version = "~> 1.0"
    }
  }
  required_version = ">= 1.5"
}

provider "mainlayer" {
  # api_key is read from the MAINLAYER_API_KEY environment variable.
}

# Create a simple pay-per-call API resource.
resource "mainlayer_resource" "my_api" {
  slug        = "my-api"
  type        = "api"
  price_usdc  = 1.00
  fee_model   = "pay_per_call"
  description = "My API — powered by Mainlayer"
  callback_url = "https://api.example.com/callback"
}

output "resource_id" {
  description = "The Mainlayer resource ID."
  value       = mainlayer_resource.my_api.id
}

output "resource_slug" {
  description = "The resource slug used in Mainlayer URLs."
  value       = mainlayer_resource.my_api.slug
}
