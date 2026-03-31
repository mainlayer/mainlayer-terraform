terraform {
  required_providers {
    mainlayer = {
      source  = "mainlayer/mainlayer"
      version = "~> 1.0"
    }
  }
  required_version = ">= 1.5"
}

# Configure the Mainlayer provider.
# The api_key can also be set via the MAINLAYER_API_KEY environment variable.
provider "mainlayer" {
  api_key = var.mainlayer_api_key
}

variable "mainlayer_api_key" {
  description = "Mainlayer API key. Set via TF_VAR_mainlayer_api_key or the MAINLAYER_API_KEY environment variable."
  type        = string
  sensitive   = true
}
