# Terraform Provider for Hugging Face Spaces

The Terraform provider for [Hugging Face Spaces](https://huggingface.co/spaces) enables users to deploy, manage, and automate Hugging Face Spaces with Terraform. Inspired by the power and flexibility of Hugging Face's platform, this provider seeks to leverage Terraform's infrastructure as code approach to streamline the management of machine learning environments and applications.

## Requirements

- Terraform 0.14.x
- Go 1.17 (to build the provider plugin)

## Installing the Provider

To install this provider, you need to build the provider binary and configure Terraform to use it:

```bash
git clone https://github.com/YOUR_GITHUB/terraform-provider-huggingface-spaces.git
cd terraform-provider-huggingface-spaces
go build -o terraform-provider-huggingface-spaces
```

Next, follow the Terraform documentation to place the binary in the correct location for your platform, ensuring Terraform can discover and use the provider plugin.

## Provider Configuration

To use this provider, you must configure it with your Hugging Face API token. This token is used to authenticate API requests on your behalf.

```hcl
provider "huggingface-spaces" {
  token = "your_hugging_face_api_token_here"
}
```

## Usage

After installing and configuring the provider, you can start defining resources in your Terraform configurations. Here is a basic example:

```hcl
resource "huggingface-spaces_space" "zenml_server" {
  name     = "test-zenml-space"
  private  = false
  template = "zenml/zenml"
}
```

This configuration will deploy a new Hugging Face Space using the zenml/zenml
template.

Other supported actions include:

- destroying a space with `terraform destroy`
- updating the name of a space by changing the resource's name in your HCL
  definition and then rerunning `terraform apply`
- updating the visibility (i.e. public vs private) of a space by changing the `private`
  attribute and then rerunning `terraform apply`
- updating and including variables and secrets for the space that is being
  deployed / created
- setting hardware requirements for the space
- adding persistent storage for the space

## Advanced Usage

A more full use might look something like this:

```hcl
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

resource "huggingface-spaces_space" "test_space" {
  name     = "test-hf-api-${formatdate("YYYYMMDD", timestamp())}"
  private  = true
  sdk      = "docker"
  template = "zenml/zenml"

  secrets = {
    SECRET_KEY_1 = "secret_value_2"
    SECRET_KEY_2 = "secret_value_3"
  }

  variables = {
    VARIABLE_KEY_1 = "variable_value_1"
    VARIABLE_KEY_2 = "variable_value_2"
  }

  hardware   = "cpu-upgrade"
  storage    = "small"
  sleep_time = 3600
}

data "huggingface-spaces_space" "test_space_data" {
  id = huggingface-spaces_space.test_space.id
}

output "test_space_id" {
  value = huggingface-spaces_space.test_space.id
}

output "test_space_name" {
  value = data.huggingface-spaces_space.test_space_data.name
}

output "test_space_author" {
  value = data.huggingface-spaces_space.test_space_data.author
}

output "test_space_last_modified" {
  value = data.huggingface-spaces_space.test_space_data.last_modified
}

output "test_space_likes" {
  value = data.huggingface-spaces_space.test_space_data.likes
}

output "test_space_private" {
  value = data.huggingface-spaces_space.test_space_data.private
}

output "test_space_sdk" {
  value = data.huggingface-spaces_space.test_space_data.sdk
}

output "test_space_hardware" {
  value = data.huggingface-spaces_space.test_space_data.hardware
}

output "test_space_storage" {
  value = data.huggingface-spaces_space.test_space_data.storage
}

output "test_space_sleep_time" {
  value = data.huggingface-spaces_space.test_space_data.sleep_time
}
```

This example demonstrates all the functionality of the Hugging Face Hub that
this provider implements.

## Making a Release

To make a release, follow these steps (using v0.0.2 as an example):

```
git tag v0.0.2
git push origin v0.0.2
```

## Development

Contributions to this provider are welcome. To contribute, please follow the standard GitHub pull request process.

- Fork the repository
- Make your changes
- Submit a pull request

### Building the Provider

```
go build -o terraform-provider-huggingface-spaces
```

### Running Tests

To run the provider's tests, use:

```
go test ./...
```

## Contributing

Contributions to improve the provider are welcome from the community. Please submit issues and pull requests with any suggestions or improvements.

## License

This project is licensed under the Apache License 2.0. See the [LICENSE](LICENSE.md) file for details.

## Acknowledgments

- Thanks to Hugging Face for their incredible platform.
- Special thanks to Sean Kane and the HashiCorp documentation for guidance in creating Terraform providers.
