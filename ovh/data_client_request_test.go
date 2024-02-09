package ovh

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const testClientRequestDataSource = `
data "ovh_client_request" "request" {
  endpoint = "%s"
}
`

func TestClientRequestDataSource(t *testing.T) {
	config := fmt.Sprintf(testClientRequestDataSource, "/auth/details")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCredentials(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.ovh_client_request.request",
						"status_code",
						"200",
					),
					resource.TestMatchResourceAttr(
						"data.ovh_client_request.request",
						"response_body",
						regexp.MustCompile(`.+`),
					),
				),
			},
		},
	})
}
