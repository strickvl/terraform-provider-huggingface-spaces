resource "huggingface-spaces_space" "test_space" {
  name     = "test-hf-api-${formatdate("YYYYMMDD", timestamp())}"
  private  = false
  sdk      = "docker"
  template = "zenml/zenml"
}
