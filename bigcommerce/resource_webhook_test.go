package bigcommerce

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccWebhook_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckWebhookBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExampleWebhookExists("bigcommerce_webhook.order_event"),
				),
			},
		},
	})
}

func TestAccWebhook_Update_Credentials(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckWebhookUpdatePre(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExampleWebhookExists("bigcommerce_webhook.order_event_update"),
					resource.TestCheckResourceAttr("bigcommerce_webhook.order_event_update", "access_token", os.Getenv("PREV_BIGCOMMERCE_ACCESS_TOKEN")),
					resource.TestCheckResourceAttr("bigcommerce_webhook.order_event_update", "client_id", os.Getenv("PREV_BIGCOMMERCE_CLIENT_ID")),
				),
			},
			{
				Config: testAccCheckWebhookUpdatePost(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExampleWebhookExists("bigcommerce_webhook.order_event_update"),
					resource.TestCheckResourceAttr("bigcommerce_webhook.order_event_update", "access_token", os.Getenv("BIGCOMMERCE_ACCESS_TOKEN")),
					resource.TestCheckResourceAttr("bigcommerce_webhook.order_event_update", "client_id", os.Getenv("BIGCOMMERCE_CLIENT_ID")),
				),
			},
		},
	})
}

func TestAccWebhook_Update_Simple(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSimpleWebhookUpdatePre(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExampleWebhookExists("bigcommerce_webhook.order_event_update"),
					resource.TestCheckResourceAttr("bigcommerce_webhook.order_event_update", "is_active", "true"),
				),
			},
			{
				Config: testAccCheckSimpleWebhookUpdatePost(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExampleWebhookExists("bigcommerce_webhook.order_event_update"),
					resource.TestCheckResourceAttr("bigcommerce_webhook.order_event_update", "is_active", "false"),
				),
			},
		},
	})
}

func TestAccWebhook_Multiple(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckWebhooksMultiple(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExampleWebhookExists("bigcommerce_webhook.order_event_multiple"),
					testAccCheckExampleWebhookExists("bigcommerce_webhook.refund_event_multiple"),
				),
			},
		},
	})
}

func testAccCheckWebhookDestroy(s *terraform.State) error {
	storeHash := testAccProvider.Meta().(string)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bigcommerce_webhook" {
			continue
		}
		accessToken := rs.Primary.Attributes["access_token"]
		clientId := rs.Primary.Attributes["client_id"]
		client := createBigCommerceClient(storeHash, clientId, accessToken)
		webhookID, _ := strconv.ParseInt(rs.Primary.ID, 10, 64)
		_, err := client.Webhooks.Get(webhookID)
		if err == nil {
			return fmt.Errorf("Webhookd still exists")
		}
		notFoundErr := "not found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}

func testAccCheckExampleWebhookExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}
		storeHash := testAccProvider.Meta().(string)
		accessToken := rs.Primary.Attributes["access_token"]
		clientId := rs.Primary.Attributes["client_id"]
		client := createBigCommerceClient(storeHash, clientId, accessToken)
		webhookID, _ := strconv.ParseInt(rs.Primary.ID, 10, 64)
		_, err := client.Webhooks.Get(webhookID)
		if err != nil {
			return fmt.Errorf("error fetching webhook with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckWebhookBasic() string {
	return fmt.Sprintf(`
resource "bigcommerce_webhook" "order_event" {
  scope       = "store/order/created"
  client_id = "%s"
  access_token = "%s"
  destination = "https://127.0.0.1/test123"
  is_active   = true
	  
  header {
    key   = "X-Functions-Key"
    value = "test123"
  }
}
`, os.Getenv("BIGCOMMERCE_CLIENT_ID"), os.Getenv("BIGCOMMERCE_ACCESS_TOKEN"))
}

func testAccCheckSimpleWebhookUpdatePre() string {
	return fmt.Sprintf(`
resource "bigcommerce_webhook" "order_event_update" {
  scope       = "store/order/created"
  client_id = "%s"
  access_token = "%s"
  destination = "https://127.0.0.1/test123"
  is_active   = true
	  
  header {
    key   = "X-Functions-Key"
    value = "test123"
  }
}
`, os.Getenv("BIGCOMMERCE_CLIENT_ID"), os.Getenv("BIGCOMMERCE_ACCESS_TOKEN"))
}

func testAccCheckSimpleWebhookUpdatePost() string {
	return fmt.Sprintf(`
resource "bigcommerce_webhook" "order_event_update" {
  scope       = "store/order/created"
  client_id = "%s"
  access_token = "%s"
  destination = "https://127.0.0.1/test123"
  is_active   = false
	  
  header {
    key   = "X-Functions-Key"
    value = "test123"
  }
}
`, os.Getenv("BIGCOMMERCE_CLIENT_ID"), os.Getenv("BIGCOMMERCE_ACCESS_TOKEN"))
}

func testAccCheckWebhookUpdatePre() string {
	return fmt.Sprintf(`
resource "bigcommerce_webhook" "order_event_update" {
  scope       = "store/order/created"
  client_id = "%s"
  access_token = "%s"
  destination = "https://127.0.0.1/test123"
  is_active   = true
	  
  header {
    key   = "X-Functions-Key"
    value = "test123"
  }
}
`, os.Getenv("PREV_BIGCOMMERCE_CLIENT_ID"), os.Getenv("PREV_BIGCOMMERCE_ACCESS_TOKEN"))
}

func testAccCheckWebhookUpdatePost() string {
	return fmt.Sprintf(`
resource "bigcommerce_webhook" "order_event_update" {
  scope       = "store/order/created"
  client_id = "%s"
  access_token = "%s"
  destination = "https://127.0.0.1/test123"
  is_active   = true
	  
  header {
    key   = "X-Functions-Key"
    value = "test123"
  }
}
`, os.Getenv("BIGCOMMERCE_CLIENT_ID"), os.Getenv("BIGCOMMERCE_ACCESS_TOKEN"))
}

func testAccCheckWebhooksMultiple() string {
	return fmt.Sprintf(`
resource "bigcommerce_webhook" "order_event_multiple" {
  scope       = "store/order/created"
  client_id = "%s"
  access_token = "%s"
  destination = "https://127.0.0.1/test123"
  is_active   = true
	  
  header {
    key   = "X-Functions-Key"
    value = "test123"
  }
}

resource "bigcommerce_webhook" "refund_event_multiple" {
	scope       = "store/order/refund/created"
	client_id = "%s"
	access_token = "%s"
	destination = "https://127.0.0.1/test123"
	is_active   = true
		
	header {
	  key   = "X-Functions-Key"
	  value = "test123"
	}
  }
`, os.Getenv("BIGCOMMERCE_CLIENT_ID"), os.Getenv("BIGCOMMERCE_ACCESS_TOKEN"), os.Getenv("BIGCOMMERCE_CLIENT_ID"), os.Getenv("BIGCOMMERCE_ACCESS_TOKEN"))
}
