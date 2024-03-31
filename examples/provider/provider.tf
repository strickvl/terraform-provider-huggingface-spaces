terraform {
  required_providers {
    huggingface-spaces = {
      source = "strickvl/huggingface-spaces"
    }
  }
}

provider "huggingface-spaces" {
  token = var.huggingface_token
}

variable "huggingface_token" {
  type        = string
  description = "The Hugging Face API token."
  sensitive   = true
}
