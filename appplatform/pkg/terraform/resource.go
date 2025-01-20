package terraform

import (
	"context"
	"encoding/json"

	"github.com/dave/dst"
	"github.com/grafana/grafana-app-sdk/resource"
	"github.com/grafana/grafana/pkg/apimachinery/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// ResourceMetadataModel represents the metadata for a Grafana resource Terraform model.
type ResourceMetadataModel struct {
	UUID      types.String `tfsdk:"uuid"`
	UID       types.String `tfsdk:"uid"`
	Version   types.String `tfsdk:"version"`
	FolderUID types.String `tfsdk:"folder_uid"`
	// TODO: should this be in metadata?
	URL types.String `tfsdk:"url"`
}

// ResourceMetadata represents the metadata for a Grafana resource.
// TODO: use SDK metadata?
type ResourceMetadata struct {
	UUID      string
	UID       string
	Version   string
	FolderUID string
	URL       string
}

// ResourceOptionsModel represents the options for a Grafana resource Terraform model.
type ResourceOptionsModel struct {
	Overwrite types.Bool `tfsdk:"overwrite"`
	Validate  types.Bool `tfsdk:"validate"`
	// TODO: This is only available for the dashboard resources.
	LintRules types.List `tfsdk:"lint_rules"`
}

// ResourceOptions represents the options for a Grafana resource.
type ResourceOptions struct {
	Overwrite bool
	Validate  bool
	// TODO: This is only available for the dashboard resources.
	LintRules []string
}

// ResourceModel represents a Grafana resource Terraform model.
type ResourceModel struct {
	Metadata types.Object `tfsdk:"metadata"`
	Spec     types.Object `tfsdk:"spec"`
	Options  types.Object `tfsdk:"options"`
}

// ParseResource parses a dashboard model into a dashboard resource.
// TODO: M should be a terraform model.
func ParseResource[M any, T resource.Object](
	ctx context.Context, plan tfsdk.Plan,
) (T, ResourceModel, diag.Diagnostics) {
	var res T
	tflog.Debug(ctx, "parsing dashboard from model to resource")

	var data M
	if diag := plan.Get(ctx, &data); diag.HasError() {
		return res, diag
	}

	meta, err := utils.MetaAccessor(dst)
	if err != nil {
		res.AddError("Failed to get request dashboard metadata", err.Error())
		return res
	}

	// Set required attributes.
	meta.SetName(src.UID.ValueString())

	// Normally a resource would be constructed from the data model,
	// but for the dashboard we expect the spec to be provided as a stringified JSON.
	if err := json.Unmarshal([]byte(src.Spec.ValueString()), &dst.Spec.Object); err != nil {
		res.AddError("Failed to parse dashboard spec", err.Error())
		return res
	}

	if src.Title.ValueString() != "" {
		dst.Spec.Object["title"] = src.Title.ValueString()
	}

	// Set optional overrides.
	if fid := src.FolderUID.ValueString(); fid != "" {
		meta.SetFolder(fid)
	}

	// Add extra tags, if set.
	if len(src.Tags.Elements()) > 0 {
		tags := make([]types.String, 0, len(src.Tags.Elements()))

		if diag := src.Tags.ElementsAs(ctx, &tags, false); diag.HasError() {
			return diag
		}

		// HACK: because the tags are not a known field in the dashboard spec,
		// we need to manually wrangle them here.
		dashtags := getTags(dst)
		for _, tag := range tags {
			dashtags = append(dashtags, tag.ValueString())
		}

		dst.Spec.Object["tags"] = dashtags
	}

	return res
}

// // Resource is a generic resource that can be used to manage any Grafana resource,
// // which is built using Grafana App Platform.
// type Resource[T sdkresource.Object, L sdkresource.ListObject] struct {
// 	schema     schema.Schema
// 	kind       sdkresource.Kind
// 	clientfunc ClientProvider[T, L]

// 	client *client.NamespacedClient[T, L]
// }

// // ClientProvider is a function that creates a new client for the resource.
// type ClientProvider[T sdkresource.Object, L sdkresource.ListObject] func(
// 	registry *k8s.ClientRegistry,
// ) (*client.ResourceClient[T, L], error)

// // NewResource creates a new Resource.
// func NewResource[T sdkresource.Object, L sdkresource.ListObject](
// 	kind sdkresource.Kind,
// 	tfschema schema.Schema,
// 	client *client.NamespacedClient[T, L],
// ) func() resource.Resource {
// 	return func() resource.Resource {
// 		return &Resource[T, L]{
// 			kind:   kind,
// 			schema: tfschema,
// 			client: client,
// 		}
// 	}
// }

// // Schema returns the schema for the Resource.
// func (r *Resource[T, L]) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
// 	res.Schema = r.schema
// }

// // Metadata returns the metadata for the Resource.
// func (r *Resource[T, L]) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
// 	resp.TypeName = fmt.Sprintf(
// 		"grafana_%s_%s", strings.ToLower(r.kind.Plural()), strings.ToLower(r.kind.Kind()),
// 	)
// }

// // Configure initializes the Resource.
// func (r *Resource[T, L]) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
// 	// TODO: provide clients from `pkg/provider/framework_provider.go`, using `req.ProviderData`.
// 	registry := k8s.NewClientRegistry(rest.Config{
// 		Host:        "https://localhost:3000",
// 		APIPath:     "/apis",
// 		BearerToken: "",
// 		UserAgent: fmt.Sprintf(
// 			"Terraform/%s (+https://www.terraform.io) terraform-provider-grafana/%s",
// 			"v1.10.1",
// 			"v0alpha1.0.1+testing",
// 		),
// 		TLSClientConfig: rest.TLSClientConfig{
// 			Insecure: true,
// 		},
// 	}, k8s.DefaultClientConfig())

// 	rcli, err := r.clientfunc(registry)
// 	if err != nil {
// 		resp.Diagnostics.AddError(
// 			"Error creating API client", err.Error(),
// 		)
// 		return
// 	}

// 	r.client = client.NewNamespaced(rcli, 1, true)
// }

// // Create creates a new dashboard.
// func (r *Resource[T, L]) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
// 	var data DashboardModel
// 	if diag := req.Plan.Get(ctx, &data); diag.HasError() {
// 		resp.Diagnostics.Append(diag...)
// 		return
// 	}

// 	var dash v0alpha1.Dashboard
// 	if diag := ParseDashboard(ctx, data, &dash); diag.HasError() {
// 		resp.Diagnostics.Append(diag...)
// 		return
// 	}

// 	res, err := r.client.Create(ctx, &dash, sdkresource.CreateOptions{})
// 	if err != nil {
// 		resp.Diagnostics.AddError("Failed to create dashboard", err.Error())
// 		return
// 	}

// 	if diag := SaveDashboardState(ctx, res, &data); diag.HasError() {
// 		resp.Diagnostics.Append(diag...)
// 		return
// 	}

// 	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
// }

// // Update updates the dashboard.
// func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
// 	var data DashboardModel
// 	if diag := req.Plan.Get(ctx, &data); diag.HasError() {
// 		resp.Diagnostics.Append(diag...)
// 		return
// 	}

// 	var opts DashboardOptions
// 	if diag := ParseOptions(ctx, data.Options, &opts); diag.HasError() {
// 		resp.Diagnostics.Append(diag...)
// 		return
// 	}

// 	var dash v0alpha1.Dashboard
// 	if diag := ParseDashboard(ctx, data, &dash); diag.HasError() {
// 		resp.Diagnostics.Append(diag...)
// 		return
// 	}

// 	reqopts := sdkresource.UpdateOptions{
// 		ResourceVersion: dash.ResourceVersion,
// 	}

// 	if opts.Overwrite {
// 		reqopts.ResourceVersion = ""
// 	}

// 	res, err := r.client.Update(ctx, &dash, reqopts)
// 	if err != nil {
// 		resp.Diagnostics.AddError("Failed to create dashboard", err.Error())
// 		return
// 	}

// 	if diag := SaveDashboardState(ctx, res, &data); diag.HasError() {
// 		resp.Diagnostics.Append(diag...)
// 		return
// 	}

// 	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
// }

// // Delete deletes the dashboard.
// func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
// 	var data DashboardModel
// 	if diag := req.State.Get(ctx, &data); diag.HasError() {
// 		resp.Diagnostics.Append(diag...)
// 		return
// 	}

// 	if err := r.client.Delete(ctx, data.UID.ValueString()); err != nil {
// 		resp.Diagnostics.AddError("Failed to delete dashboard", err.Error())
// 		return
// 	}
// }

// // ImportState imports the state of the dashboard.
// func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
// 	res, err := r.client.Get(ctx, req.ID)
// 	if err != nil {
// 		resp.Diagnostics.AddError("Failed to get dashboard", err.Error())
// 		return
// 	}

// 	var data DashboardModel
// 	if diag := SaveDashboardState(ctx, res, &data); diag.HasError() {
// 		resp.Diagnostics.Append(diag...)
// 		return
// 	}

// 	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
// }

// // Read reads the dashboard.
// func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
// 	var data DashboardModel
// 	if diag := req.State.Get(ctx, &data); diag.HasError() {
// 		resp.Diagnostics.Append(diag...)
// 		return
// 	}

// 	res, err := r.client.Get(ctx, data.UID.ValueString())
// 	if err != nil {
// 		resp.Diagnostics.AddError("Failed to create dashboard", err.Error())
// 		return
// 	}

// 	if diag := SaveDashboardState(ctx, res, &data); diag.HasError() {
// 		resp.Diagnostics.Append(diag...)
// 		return
// 	}

// 	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
// }

// func (r *Resource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
// 	var data DashboardModel
// 	if diag := req.Config.Get(ctx, &data); diag.HasError() {
// 		resp.Diagnostics.Append(diag...)
// 		return
// 	}

// 	var opts DashboardOptions
// 	if diag := ParseOptions(ctx, data.Options, &opts); diag.HasError() {
// 		resp.Diagnostics.Append(diag...)
// 		return
// 	}

// 	if opts.Validate {
// 		if err := ValidateDashboard([]byte(data.Spec.ValueString())); err != nil {
// 			resp.Diagnostics.AddError("Invalid dashboard spec", err.Error())
// 			return
// 		}
// 	}

// 	if len(opts.LintRules) > 0 {
// 		results, ok, err := LintDashboard(
// 			GetLintRules(opts.LintRules), []byte(data.Spec.ValueString()),
// 		)
// 		if err != nil {
// 			resp.Diagnostics.AddError("Failed to lint dashboard", err.Error())
// 			return
// 		}

// 		if ok {
// 			return
// 		}

// 		resp.Diagnostics.AddWarning(
// 			path.Root("spec").String(), results.Warnings,
// 		)

// 		resp.Diagnostics.AddError(
// 			path.Root("spec").String(), results.Errors,
// 		)
// 	}
// }
