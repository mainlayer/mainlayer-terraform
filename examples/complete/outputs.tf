# Complete example outputs configuration
output "vendor_details" {
  description = "Vendor account information"
  value = {
    id          = mainlayer_vendor.acme.id
    name        = mainlayer_vendor.acme.name
    email       = mainlayer_vendor.acme.email
    created_at  = mainlayer_vendor.acme.created_at
    updated_at  = mainlayer_vendor.acme.updated_at
  }
}

output "resources_by_type" {
  description = "All resources organized by type"
  value = {
    api     = mainlayer_resource.text_analysis_api.slug
    model   = mainlayer_resource.gpt_wrapper.slug
    dataset = mainlayer_resource.training_data.slug
    tool    = mainlayer_resource.image_enhancement.slug
  }
}

output "subscription_pricing" {
  description = "Pricing for all subscription plans"
  value = {
    starter    = "${mainlayer_plan.gpt_starter.price_usdc} USDC/month"
    pro        = "${mainlayer_plan.gpt_pro.price_usdc} USDC/month"
    enterprise = "${mainlayer_plan.gpt_enterprise.price_usdc} USDC/month"
    annual     = "${mainlayer_plan.gpt_annual.price_usdc} USDC/year"
  }
}
