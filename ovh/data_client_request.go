package ovh

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ovh/go-ovh/ovh"
)

var _ datasource.DataSource = (*clientRequestDataSource)(nil)

func NewClientRequestDataSource() datasource.DataSource {
	return &clientRequestDataSource{}
}

type clientRequestDataSource struct {
	OVHClient *ovh.Client
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &clientRequestDataSource{}
	_ datasource.DataSourceWithConfigure = &clientRequestDataSource{}
)

func (crdt *clientRequestDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_client_request"
}

// Configure implements datasource.DataSourceWithConfigure.
func (crdt *clientRequestDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	config, ok := req.ProviderData.(*Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	crdt.OVHClient = config.OVHClient
}

type clientRequestModel struct {
	Endpoint     types.String `tfsdk:"endpoint"`
	ResponseBody types.String `tfsdk:"response_body"`
	StatusCode   types.Int64  `tfsdk:"status_code"`
	QueryID      types.String `tfsdk:"query_id"`
}

func (crdt *clientRequestDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Description: "OVH API endpoint for the request",
				Required:    true,
			},

			"response_body": schema.StringAttribute{
				Description: "The API response body returned as a string",
				Computed:    true,
			},

			"status_code": schema.Int64Attribute{
				Description: "The HTTP response status code",
				Computed:    true,
			},
			"query_id": schema.StringAttribute{
				Description: "X-Ovh-QueryID value from the response header",
				Computed:    true,
			},
		},
	}
}

func (crdt *clientRequestDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model clientRequestModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	endpoint := model.Endpoint.String()

	log.Printf("[DEBUG] Will make request to %s endpoint", endpoint)

	ovhReq, err := crdt.OVHClient.NewRequest("GET", endpoint, nil, true)
	if err != nil {
		resp.Diagnostics.AddError("Failed to generate request", err.Error())
		return
	}

	ovhRes, err := crdt.OVHClient.Do(ovhReq.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to make request", err.Error())
		return
	}

	model.StatusCode = types.Int64Value(int64(ovhRes.StatusCode))
	model.QueryID = types.StringValue(ovhRes.Header.Get("X-Ovh-QueryID"))

	body, err := io.ReadAll(ovhRes.Body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read ovhRes", err.Error())
		return
	}
	model.ResponseBody = types.StringValue(string(body))
	if !json.Valid(body) {
		resp.Diagnostics.AddWarning("Not JSON serializable response body", "The response body is not json serializable")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
