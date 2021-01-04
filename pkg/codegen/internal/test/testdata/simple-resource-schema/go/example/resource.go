// *** WARNING: this file was generated by test. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

package example

import (
	"context"
	"reflect"

	"github.com/pulumi/pulumi/sdk/v2/go/pulumi"
)

type Resource struct {
	pulumi.CustomResourceState

	Bar pulumi.StringPtrOutput `pulumi:"bar"`
}

// NewResource registers a new resource with the given unique name, arguments, and options.
func NewResource(ctx *pulumi.Context,
	name string, args *ResourceArgs, opts ...pulumi.ResourceOption) (*Resource, error) {
	if args == nil {
		args = &ResourceArgs{}
	}

	var resource Resource
	err := ctx.RegisterResource("example::Resource", name, args, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// GetResource gets an existing Resource resource's state with the given name, ID, and optional
// state properties that are used to uniquely qualify the lookup (nil if not required).
func GetResource(ctx *pulumi.Context,
	name string, id pulumi.IDInput, state *ResourceState, opts ...pulumi.ResourceOption) (*Resource, error) {
	var resource Resource
	err := ctx.ReadResource("example::Resource", name, id, state, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// Input properties used for looking up and filtering Resource resources.
type resourceState struct {
	Bar *string `pulumi:"bar"`
}

type ResourceState struct {
	Bar pulumi.StringPtrInput
}

func (ResourceState) ElementType() reflect.Type {
	return reflect.TypeOf((*resourceState)(nil)).Elem()
}

type resourceArgs struct {
	Bar *string `pulumi:"bar"`
}

// The set of arguments for constructing a Resource resource.
type ResourceArgs struct {
	Bar pulumi.StringPtrInput
}

func (ResourceArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*resourceArgs)(nil)).Elem()
}

type ResourceInput interface {
	pulumi.Input

	ToResourceOutput() ResourceOutput
	ToResourceOutputWithContext(ctx context.Context) ResourceOutput
}

func (*Resource) ElementType() reflect.Type {
	return reflect.TypeOf((*Resource)(nil))
}

func (i *Resource) ToResourceOutput() ResourceOutput {
	return i.ToResourceOutputWithContext(context.Background())
}

func (i *Resource) ToResourceOutputWithContext(ctx context.Context) ResourceOutput {
	return pulumi.ToOutputWithContext(ctx, i).(ResourceOutput)
}

func (i *Resource) ToResourcePtrOutput() ResourcePtrOutput {
	return i.ToResourcePtrOutputWithContext(context.Background())
}

func (i *Resource) ToResourcePtrOutputWithContext(ctx context.Context) ResourcePtrOutput {
	return pulumi.ToOutputWithContext(ctx, i).(ResourcePtrOutput)
}

type ResourcePtrInput interface {
	pulumi.Input

	ToResourcePtrOutput() ResourcePtrOutput
	ToResourcePtrOutputWithContext(ctx context.Context) ResourcePtrOutput
}

type resourcePtrType ResourceArgs

func (*resourcePtrType) ElementType() reflect.Type {
	return reflect.TypeOf((**Resource)(nil))
}

func (i *resourcePtrType) ToResourcePtrOutput() ResourcePtrOutput {
	return i.ToResourcePtrOutputWithContext(context.Background())
}

func (i *resourcePtrType) ToResourcePtrOutputWithContext(ctx context.Context) ResourcePtrOutput {
	return pulumi.ToOutputWithContext(ctx, i).(ResourceOutput).ToResourcePtrOutput()
}

// ResourceArrayInput is an input type that accepts ResourceArray and ResourceArrayOutput values.
// You can construct a concrete instance of `ResourceArrayInput` via:
//
//          ResourceArray{ ResourceArgs{...} }
type ResourceArrayInput interface {
	pulumi.Input

	ToResourceArrayOutput() ResourceArrayOutput
	ToResourceArrayOutputWithContext(context.Context) ResourceArrayOutput
}

type ResourceArray []ResourceInput

func (ResourceArray) ElementType() reflect.Type {
	return reflect.TypeOf((*[]Resource)(nil))
}

func (i ResourceArray) ToResourceArrayOutput() ResourceArrayOutput {
	return i.ToResourceArrayOutputWithContext(context.Background())
}

func (i ResourceArray) ToResourceArrayOutputWithContext(ctx context.Context) ResourceArrayOutput {
	return pulumi.ToOutputWithContext(ctx, i).(ResourceArrayOutput)
}

// ResourceMapInput is an input type that accepts ResourceMap and ResourceMapOutput values.
// You can construct a concrete instance of `ResourceMapInput` via:
//
//          ResourceMap{ "key": ResourceArgs{...} }
type ResourceMapInput interface {
	pulumi.Input

	ToResourceMapOutput() ResourceMapOutput
	ToResourceMapOutputWithContext(context.Context) ResourceMapOutput
}

type ResourceMap map[string]ResourceInput

func (ResourceMap) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]Resource)(nil))
}

func (i ResourceMap) ToResourceMapOutput() ResourceMapOutput {
	return i.ToResourceMapOutputWithContext(context.Background())
}

func (i ResourceMap) ToResourceMapOutputWithContext(ctx context.Context) ResourceMapOutput {
	return pulumi.ToOutputWithContext(ctx, i).(ResourceMapOutput)
}

type ResourceOutput struct {
	*pulumi.OutputState
}

func (ResourceOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*Resource)(nil))
}

func (o ResourceOutput) ToResourceOutput() ResourceOutput {
	return o
}

func (o ResourceOutput) ToResourceOutputWithContext(ctx context.Context) ResourceOutput {
	return o
}

func (o ResourceOutput) ToResourcePtrOutput() ResourcePtrOutput {
	return o.ToResourcePtrOutputWithContext(context.Background())
}

func (o ResourceOutput) ToResourcePtrOutputWithContext(ctx context.Context) ResourcePtrOutput {
	return o.ApplyT(func(v Resource) *Resource {
		return &v
	}).(ResourcePtrOutput)
}

type ResourcePtrOutput struct {
	*pulumi.OutputState
}

func (ResourcePtrOutput) ElementType() reflect.Type {
	return reflect.TypeOf((**Resource)(nil))
}

func (o ResourcePtrOutput) ToResourcePtrOutput() ResourcePtrOutput {
	return o
}

func (o ResourcePtrOutput) ToResourcePtrOutputWithContext(ctx context.Context) ResourcePtrOutput {
	return o
}

type ResourceArrayOutput struct{ *pulumi.OutputState }

func (ResourceArrayOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*[]Resource)(nil))
}

func (o ResourceArrayOutput) ToResourceArrayOutput() ResourceArrayOutput {
	return o
}

func (o ResourceArrayOutput) ToResourceArrayOutputWithContext(ctx context.Context) ResourceArrayOutput {
	return o
}

func (o ResourceArrayOutput) Index(i pulumi.IntInput) ResourceOutput {
	return pulumi.All(o, i).ApplyT(func(vs []interface{}) Resource {
		return vs[0].([]Resource)[vs[1].(int)]
	}).(ResourceOutput)
}

type ResourceMapOutput struct{ *pulumi.OutputState }

func (ResourceMapOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]Resource)(nil))
}

func (o ResourceMapOutput) ToResourceMapOutput() ResourceMapOutput {
	return o
}

func (o ResourceMapOutput) ToResourceMapOutputWithContext(ctx context.Context) ResourceMapOutput {
	return o
}

func (o ResourceMapOutput) MapIndex(k pulumi.StringInput) ResourceOutput {
	return pulumi.All(o, k).ApplyT(func(vs []interface{}) Resource {
		return vs[0].(map[string]Resource)[vs[1].(string)]
	}).(ResourceOutput)
}

func init() {
	pulumi.RegisterOutputType(ResourceOutput{})
	pulumi.RegisterOutputType(ResourcePtrOutput{})
	pulumi.RegisterOutputType(ResourceArrayOutput{})
	pulumi.RegisterOutputType(ResourceMapOutput{})
}
