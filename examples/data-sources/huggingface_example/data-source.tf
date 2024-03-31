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
