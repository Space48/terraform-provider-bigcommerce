package bigcommerce

import (
	"context"
	"net/http"
	"strconv"

	"github.com/ashsmith/bigcommerce-api-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceWebhook() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a BigCommerce Webhook resource.",
		CreateContext: resourceWebhookCreate,
		ReadContext:   resourceWebhookRead,
		UpdateContext: resourceWebhookUpdate,
		DeleteContext: resourceWebhookDelete,
		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"client_id": &schema.Schema{
				Type:      schema.TypeString,
				Sensitive: true,
				Required:  true,
			},
			"access_token": &schema.Schema{
				Type:      schema.TypeString,
				Sensitive: true,
				Required:  true,
			},
			"scope": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"destination": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"is_active": &schema.Schema{
				Type:     schema.TypeBool,
				Required: true,
			},
			"header": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceWebhookCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	storeHash := m.(string)
	clientId := d.Get("client_id").(string)
	accessToken := d.Get("access_token").(string)

	client := createBigCommerceClient(storeHash, clientId, accessToken)
	var diags diag.Diagnostics

	scope := d.Get("scope").(string)
	destination := d.Get("destination").(string)
	isActive := d.Get("is_active").(bool)

	webhook := bigcommerce.Webhook{
		Scope:       scope,
		Destination: destination,
		IsActive:    isActive,
	}

	webhook.Headers = formatHeaders(d)

	result, err := client.Webhooks.Create(webhook)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(int(result.ID)))

	return diags
}

func resourceWebhookRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	storeHash := m.(string)
	clientId := d.Get("client_id").(string)
	accessToken := d.Get("access_token").(string)

	client := createBigCommerceClient(storeHash, clientId, accessToken)
	var diags diag.Diagnostics

	webhookID, _ := strconv.ParseInt(d.Id(), 10, 64)

	webhook, whErr := client.Webhooks.Get(webhookID)
	if whErr != nil {
		return diag.FromErr(whErr)
	}

	err := setWebhookData(webhook, d)
	if err != nil {
		return err
	}

	return diags
}

func resourceWebhookUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	prevClientId, currentClientId := d.GetChange("client_id")
	prevAccessToken, currentAccessToken := d.GetChange("access_token")
	webhookID, _ := strconv.ParseInt(d.Id(), 10, 64)
	storeHash := m.(string)
	if d.HasChange("client_id") || d.HasChange("access_token") {
		prevClient := createBigCommerceClient(storeHash, prevClientId.(string), prevAccessToken.(string))
		currentClient := createBigCommerceClient(storeHash, currentClientId.(string), currentAccessToken.(string))
		err := prevClient.Webhooks.Delete(webhookID)
		if err != nil {
			return diag.FromErr(err)
		}
		webhook := bigcommerce.Webhook{
			Scope:       d.Get("scope").(string),
			Destination: d.Get("destination").(string),
			IsActive:    d.Get("is_active").(bool),
		}

		webhook.Headers = formatHeaders(d)
		result, err := currentClient.Webhooks.Create(webhook)
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(strconv.Itoa(int(result.ID)))
	} else {
		if d.HasChange("scope") || d.HasChange("destination") || d.HasChange("is_active") || d.HasChange("header") {
			clientId := d.Get("client_id").(string)
			accessToken := d.Get("access_token").(string)

			client := createBigCommerceClient(storeHash, clientId, accessToken)
			webhook := bigcommerce.Webhook{
				ID:          webhookID,
				Scope:       d.Get("scope").(string),
				Destination: d.Get("destination").(string),
				IsActive:    d.Get("is_active").(bool),
			}

			webhook.Headers = formatHeaders(d)

			_, err := client.Webhooks.Update(webhook)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	return resourceWebhookRead(ctx, d, m)
}

func resourceWebhookDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	storeHash := m.(string)
	clientId := d.Get("client_id").(string)
	accessToken := d.Get("access_token").(string)

	client := createBigCommerceClient(storeHash, clientId, accessToken)
	var diags diag.Diagnostics

	webhookID, _ := strconv.ParseInt(d.Id(), 10, 64)

	err := client.Webhooks.Delete(webhookID)
	if err != nil {
		return diag.FromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}

func createBigCommerceClient(storeHash string, clientID string, accessToken string) *bigcommerce.Client {
	app := bigcommerce.App{
		ClientID:    clientID,
		StoreHash:   storeHash,
		AccessToken: accessToken,
	}

	httpClient := http.Client{}
	return app.NewClient(httpClient)
}

func formatHeaders(d *schema.ResourceData) map[string]string {
	wbHeaders := make(map[string]string)
	headers := d.Get("header").(*schema.Set).List()
	for _, item := range headers {
		header := item.(map[string]interface{})
		wbHeaders[header["key"].(string)] = header["value"].(string)
	}
	return wbHeaders
}
