# Variables for Mainlayer Terraform provider examples.
#
# Usage:
#   export TF_VAR_mainlayer_api_key="$MAINLAYER_API_KEY"
#   terraform plan

variable "mainlayer_api_key" {
  description = "Mainlayer API key. Can also be set via MAINLAYER_API_KEY env var."
  type        = string
  sensitive   = true
  default     = ""
}

variable "mainlayer_base_url" {
  description = "Override the Mainlayer API base URL. Defaults to https://api.mainlayer.xyz."
  type        = string
  default     = "https://api.mainlayer.xyz"
}

variable "resource_slug" {
  description = "URL-friendly slug for the Mainlayer resource."
  type        = string
  default     = "my-api"

  validation {
    condition     = can(regex("^[a-z0-9-]+$", var.resource_slug))
    error_message = "resource_slug must be lowercase alphanumeric with hyphens only."
  }
}

variable "resource_type" {
  description = "Type of resource: api, dataset, or model."
  type        = string
  default     = "api"

  validation {
    condition     = contains(["api", "dataset", "model"], var.resource_type)
    error_message = "resource_type must be one of: api, dataset, model."
  }
}

variable "price_usdc" {
  description = "Price per call in USD."
  type        = number
  default     = 1.00
}

variable "fee_model" {
  description = "Fee model: pay_per_call or subscription."
  type        = string
  default     = "pay_per_call"

  validation {
    condition     = contains(["pay_per_call", "subscription"], var.fee_model)
    error_message = "fee_model must be one of: pay_per_call, subscription."
  }
}

variable "callback_url" {
  description = "Webhook URL for Mainlayer payment notifications."
  type        = string
  default     = "https://api.example.com/mainlayer/callback"
}
