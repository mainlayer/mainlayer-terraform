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
  # API key from MAINLAYER_API_KEY environment variable
}

# Register or manage a vendor account
resource "mainlayer_vendor" "acme" {
  name        = "Acme AI"
  email       = "payments@acmeai.com"
  website     = "https://acmeai.com"
  country     = "US"
  description = "AI-powered API provider specializing in NLP and computer vision"
}

# Create an API resource (pay-per-call)
resource "mainlayer_resource" "text_analysis_api" {
  slug         = "text-analysis"
  type         = "api"
  price_usdc   = 0.001      # $0.001 per call
  fee_model    = "pay_per_call"
  description  = "Real-time text analysis API with sentiment and entity extraction"
  callback_url = "https://api.acmeai.com/mainlayer/payment-webhook"
}

# Create a model resource with subscription plans
resource "mainlayer_resource" "gpt_wrapper" {
  slug         = "gpt-wrapper"
  type         = "model"
  price_usdc   = 0.0        # Subscription only
  fee_model    = "subscription"
  description  = "Access to curated GPT models with fine-tuning"
  callback_url = "https://api.acmeai.com/models/callback"
}

# Define tiered subscription plans for the model
resource "mainlayer_plan" "gpt_starter" {
  resource_id = mainlayer_resource.gpt_wrapper.id
  name        = "Starter"
  description = "1,000 calls per month. Perfect for testing."
  price_usdc  = 9.99
  call_limit  = 1000
  period      = "monthly"
}

resource "mainlayer_plan" "gpt_pro" {
  resource_id = mainlayer_resource.gpt_wrapper.id
  name        = "Pro"
  description = "50,000 calls per month with priority support"
  price_usdc  = 99.00
  call_limit  = 50000
  period      = "monthly"
}

resource "mainlayer_plan" "gpt_enterprise" {
  resource_id = mainlayer_resource.gpt_wrapper.id
  name        = "Enterprise"
  description = "Unlimited calls with dedicated support and SLA"
  price_usdc  = 999.00
  call_limit  = 0  # Unlimited
  period      = "monthly"
}

resource "mainlayer_plan" "gpt_annual" {
  resource_id = mainlayer_resource.gpt_wrapper.id
  name        = "Annual Pro"
  description = "Pro tier with annual billing discount (2 months free)"
  price_usdc  = 1089.00      # 11 months of payment
  call_limit  = 50000
  period      = "yearly"
}

# Create a dataset resource
resource "mainlayer_resource" "training_data" {
  slug         = "training-dataset"
  type         = "dataset"
  price_usdc   = 100.0
  fee_model    = "pay_per_call"
  description  = "High-quality labeled dataset for ML model training"
  callback_url = "https://api.acmeai.com/datasets/callback"
}

# Create a tool resource
resource "mainlayer_resource" "image_enhancement" {
  slug         = "image-enhancement"
  type         = "tool"
  price_usdc   = 0.05
  fee_model    = "pay_per_call"
  description  = "AI-powered image upscaling and enhancement tool"
  callback_url = "https://api.acmeai.com/tools/callback"
}

# Data source: list all resources
data "mainlayer_resources" "all" {}

# Outputs
output "vendor_id" {
  description = "The vendor ID for Acme AI"
  value       = mainlayer_vendor.acme.id
}

output "vendor_api_key" {
  description = "API key for the vendor (store securely)"
  value       = mainlayer_vendor.acme.api_key
  sensitive   = true
}

output "text_analysis_id" {
  description = "Resource ID for the text analysis API"
  value       = mainlayer_resource.text_analysis_api.id
}

output "gpt_wrapper_id" {
  description = "Resource ID for the GPT model"
  value       = mainlayer_resource.gpt_wrapper.id
}

output "gpt_plans" {
  description = "Plans available for the GPT model"
  value = {
    starter    = mainlayer_plan.gpt_starter.id
    pro        = mainlayer_plan.gpt_pro.id
    enterprise = mainlayer_plan.gpt_enterprise.id
    annual     = mainlayer_plan.gpt_annual.id
  }
}

output "all_resources_count" {
  description = "Total number of resources"
  value       = length(data.mainlayer_resources.all.resources)
}
