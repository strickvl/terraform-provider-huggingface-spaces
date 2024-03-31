package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccHuggingFaceSpacesSpace(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccHuggingFaceSpacesSpaceConfig("test-space", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("huggingface-spaces_space.test", "name", "test-space"),
					resource.TestCheckResourceAttr("huggingface-spaces_space.test", "private", "false"),
					resource.TestCheckResourceAttr("huggingface-spaces_space.test", "sdk", "docker"),
					resource.TestCheckResourceAttr("huggingface-spaces_space.test", "template", "zenml/zenml"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "huggingface-spaces_space.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccHuggingFaceSpacesSpaceConfig("updated-test-space", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("huggingface-spaces_space.test", "name", "updated-test-space"),
					resource.TestCheckResourceAttr("huggingface-spaces_space.test", "private", "true"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccHuggingFaceSpacesSpaceConfig(name string, private bool) string {
	return fmt.Sprintf(`
resource "huggingface-spaces_space" "test" {
  name     = %[1]q
  private  = %[2]t
  sdk      = "docker"
  template = "zenml/zenml"
}
`, name, private)
}
