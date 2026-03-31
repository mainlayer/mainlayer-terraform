package tests

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/mainlayer/terraform-provider-mainlayer/internal/provider"
)

// providerFactories maps the provider name to a protocol-v6 server factory.
// This is used by all acceptance tests.
var providerFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"mainlayer": providerserver.NewProtocol6WithError(provider.New("test")()),
}

// testAccPreCheck validates that required environment variables are set before
// running acceptance tests.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	if v := os.Getenv("MAINLAYER_API_KEY"); v == "" {
		t.Skip("MAINLAYER_API_KEY must be set for acceptance tests")
	}
}

// --- Provider configuration tests ---

func TestAccProvider_MissingAPIKey(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccProviderConfigNoKey(),
				ExpectError: regexp.MustCompile(`Missing API Key`),
			},
		},
	})
}

func testAccProviderConfigNoKey() string {
	return `
provider "mainlayer" {}

data "mainlayer_resources" "all" {}
`
}

// --- mainlayer_resource acceptance tests ---

func TestAccResourceResource_basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Set TF_ACC=1 to run acceptance tests")
	}
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			// Create and verify.
			{
				Config: testAccResourceResourceConfig("test-api", "api", 0.01, "pay_per_call", "Test API resource"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mainlayer_resource.test", "slug", "test-api"),
					resource.TestCheckResourceAttr("mainlayer_resource.test", "type", "api"),
					resource.TestCheckResourceAttr("mainlayer_resource.test", "fee_model", "pay_per_call"),
					resource.TestCheckResourceAttr("mainlayer_resource.test", "description", "Test API resource"),
					resource.TestCheckResourceAttrSet("mainlayer_resource.test", "id"),
					resource.TestCheckResourceAttrSet("mainlayer_resource.test", "created_at"),
					resource.TestCheckResourceAttrSet("mainlayer_resource.test", "updated_at"),
				),
			},
			// Update description and price.
			{
				Config: testAccResourceResourceConfig("test-api", "api", 0.05, "pay_per_call", "Updated API resource"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mainlayer_resource.test", "slug", "test-api"),
					resource.TestCheckResourceAttr("mainlayer_resource.test", "price_usdc", "0.05"),
					resource.TestCheckResourceAttr("mainlayer_resource.test", "description", "Updated API resource"),
				),
			},
			// Verify import.
			{
				ResourceName:      "mainlayer_resource.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceResource_withCallback(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Set TF_ACC=1 to run acceptance tests")
	}
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceResourceWithCallbackConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mainlayer_resource.with_callback", "slug", "callback-api"),
					resource.TestCheckResourceAttr("mainlayer_resource.with_callback", "callback_url", "https://api.example.com/callback"),
				),
			},
		},
	})
}

func TestAccResourceResource_disappears(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Set TF_ACC=1 to run acceptance tests")
	}
	testAccPreCheck(t)

	var resourceID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceResourceConfig("disappears-api", "api", 0.01, "pay_per_call", "Will be deleted"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists("mainlayer_resource.test", &resourceID),
				),
			},
		},
	})
}

func testAccCheckResourceExists(n string, id *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID is not set")
		}
		*id = rs.Primary.ID
		return nil
	}
}

func testAccResourceResourceConfig(slug, resType string, price float64, feeModel, description string) string {
	return fmt.Sprintf(`
provider "mainlayer" {}

resource "mainlayer_resource" "test" {
  slug        = %q
  type        = %q
  price_usdc  = %g
  fee_model   = %q
  description = %q
}
`, slug, resType, price, feeModel, description)
}

func testAccResourceResourceWithCallbackConfig() string {
	return `
provider "mainlayer" {}

resource "mainlayer_resource" "with_callback" {
  slug         = "callback-api"
  type         = "api"
  price_usdc   = 0.02
  fee_model    = "pay_per_call"
  description  = "Resource with a callback URL"
  callback_url = "https://api.example.com/callback"
}
`
}

// --- mainlayer_plan acceptance tests ---

func TestAccPlanResource_basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Set TF_ACC=1 to run acceptance tests")
	}
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPlanResourceConfig("Starter", 9.99, 1000, "monthly"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mainlayer_plan.test", "name", "Starter"),
					resource.TestCheckResourceAttr("mainlayer_plan.test", "price_usdc", "9.99"),
					resource.TestCheckResourceAttr("mainlayer_plan.test", "call_limit", "1000"),
					resource.TestCheckResourceAttr("mainlayer_plan.test", "period", "monthly"),
					resource.TestCheckResourceAttrSet("mainlayer_plan.test", "id"),
					resource.TestCheckResourceAttrSet("mainlayer_plan.test", "resource_id"),
				),
			},
			// Update plan name and price.
			{
				Config: testAccPlanResourceConfig("Starter Pro", 14.99, 2000, "monthly"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mainlayer_plan.test", "name", "Starter Pro"),
					resource.TestCheckResourceAttr("mainlayer_plan.test", "price_usdc", "14.99"),
					resource.TestCheckResourceAttr("mainlayer_plan.test", "call_limit", "2000"),
				),
			},
			// Verify import using <resource_id>/<plan_id> format.
			{
				ResourceName:      "mainlayer_plan.test",
				ImportState:       true,
				ImportStateIdFunc: testAccPlanImportStateIDFunc("mainlayer_plan.test"),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccPlanImportStateIDFunc(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("plan resource not found: %s", n)
		}
		resourceID := rs.Primary.Attributes["resource_id"]
		planID := rs.Primary.ID
		return fmt.Sprintf("%s/%s", resourceID, planID), nil
	}
}

func testAccPlanResourceConfig(name string, price float64, callLimit int, period string) string {
	return fmt.Sprintf(`
provider "mainlayer" {}

resource "mainlayer_resource" "parent" {
  slug       = "plan-parent-api"
  type       = "api"
  price_usdc = 0.01
  fee_model  = "subscription"
}

resource "mainlayer_plan" "test" {
  resource_id = mainlayer_resource.parent.id
  name        = %q
  price_usdc  = %g
  call_limit  = %d
  period      = %q
}
`, name, price, callLimit, period)
}

// --- mainlayer_resources data source tests ---

func TestAccResourcesDataSource_basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Set TF_ACC=1 to run acceptance tests")
	}
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourcesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.mainlayer_resources.all", "id"),
					resource.TestCheckResourceAttrSet("data.mainlayer_resources.all", "resources.#"),
				),
			},
		},
	})
}

func testAccResourcesDataSourceConfig() string {
	return `
provider "mainlayer" {}

data "mainlayer_resources" "all" {}
`
}
