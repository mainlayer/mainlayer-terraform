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

# Create a subscription-based API resource.
resource "mainlayer_resource" "analytics_api" {
  slug         = "analytics-api"
  type         = "api"
  price_usdc   = 0.00
  fee_model    = "subscription"
  description  = "Analytics API with tiered subscription plans"
  callback_url = "https://analytics.example.com/mainlayer/callback"
}

# Starter plan — 1,000 calls/month.
resource "mainlayer_plan" "starter" {
  resource_id  = mainlayer_resource.analytics_api.id
  name         = "Starter"
  description  = "Up to 1,000 API calls per month. Great for indie developers."
  price_usdc   = 9.99
  call_limit   = 1000
  period       = "monthly"
}

# Pro plan — 25,000 calls/month.
resource "mainlayer_plan" "pro" {
  resource_id  = mainlayer_resource.analytics_api.id
  name         = "Pro"
  description  = "Up to 25,000 API calls per month. Built for growing teams."
  price_usdc   = 49.00
  call_limit   = 25000
  period       = "monthly"
}

# Enterprise plan — unlimited calls.
resource "mainlayer_plan" "enterprise" {
  resource_id  = mainlayer_resource.analytics_api.id
  name         = "Enterprise"
  description  = "Unlimited API calls per month with priority support."
  price_usdc   = 299.00
  call_limit   = 0
  period       = "monthly"
}

# Query all existing resources as a data source.
data "mainlayer_resources" "all" {}

output "resource_id" {
  description = "The Analytics API resource ID."
  value       = mainlayer_resource.analytics_api.id
}

output "starter_plan_id" {
  description = "Starter plan ID."
  value       = mainlayer_plan.starter.id
}

output "pro_plan_id" {
  description = "Pro plan ID."
  value       = mainlayer_plan.pro.id
}

output "enterprise_plan_id" {
  description = "Enterprise plan ID."
  value       = mainlayer_plan.enterprise.id
}

output "all_resource_count" {
  description = "Total number of resources in the account."
  value       = length(data.mainlayer_resources.all.resources)
}
