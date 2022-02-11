package bigcommerce

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"store_hash": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("BIGCOMMERCE_STORE_HASH", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"bigcommerce_webhook": resourceWebhook(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"bigcommerce_webhook": dataSourceWebhook(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	storeHash := d.Get("store_hash").(string)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if storeHash == "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Missing store_hash from provider configuration",
			Detail:   "store_hash is a required parmeter and must be defined, you can also use BIGCOMMERCE_STORE_HASH environment variable.",
		})
	}

	if storeHash == "" {
		return nil, diags
	}

	return storeHash, diags
}
