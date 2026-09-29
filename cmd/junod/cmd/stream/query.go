package stream

import (
	"fmt"
	"maps"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	proto "github.com/cosmos/gogoproto/proto"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/client/v2/autocli"
	flagbuilder "cosmossdk.io/client/v2/autocli/flag"

	"github.com/CosmosContracts/juno/v31/x/stream/types/encoding"
)

const StreamFlagName = "stream"

func EnableQueryStreaming(rootCmd *cobra.Command, appOpts autocli.AppOptions) error {
	queryCmd := findCommand(rootCmd, "query")
	if queryCmd == nil {
		return nil
	}

	moduleOptions := collectModuleOptions(appOpts)
	if len(moduleOptions) == 0 {
		return nil
	}

	var fileResolver flagbuilder.FileResolver
	var protoFiles *protoregistry.Files
	var err error
	protoFiles, err = proto.MergedRegistry()
	if err != nil {
		fileResolver = appOpts.ClientCtx.InterfaceRegistry
	} else {
		fileResolver = protoFiles
	}

	builder := &autocli.Builder{
		Builder: flagbuilder.Builder{
			TypeResolver:          protoregistry.GlobalTypes,
			FileResolver:          fileResolver,
			AddressCodec:          appOpts.AddressCodec,
			ValidatorAddressCodec: appOpts.ValidatorAddressCodec,
			ConsensusAddressCodec: appOpts.ConsensusAddressCodec,
		},
	}

	if err := builder.ValidateAndComplete(); err != nil {
		return err
	}

	var registry *encoding.DynamicRegistry
	if protoFiles != nil {
		if reg, regErr := encoding.NewFilesRegistry(protoFiles); regErr == nil {
			registry = reg
		}
	}

	for moduleName, options := range moduleOptions {
		if options == nil || options.Query == nil {
			continue
		}

		moduleCmd := findCommand(queryCmd, moduleName)
		if moduleCmd == nil {
			continue
		}

		if err := wrapQueryServiceWithStreaming(moduleCmd, options.Query, builder, registry); err != nil {
			return err
		}
	}

	return nil
}

func findCommand(parent *cobra.Command, name string) *cobra.Command {
	for _, cmd := range parent.Commands() {
		use := cmd.Use
		if use == name || strings.HasPrefix(use, name+" ") {
			return cmd
		}
		for _, alias := range cmd.Aliases {
			if alias == name || strings.HasPrefix(alias, name+" ") {
				return cmd
			}
		}
	}
	return nil
}

func protoNameToCLIName(name protoreflect.Name) string {
	out := make([]rune, 0, len(name))
	for i, r := range name {
		if unicode.IsUpper(r) {
			if i != 0 {
				out = append(out, '-')
			}
			out = append(out, unicode.ToLower(r))
			continue
		}
		out = append(out, r)
	}

	return string(out)
}

func collectModuleOptions(appOpts autocli.AppOptions) map[string]*autocliv1.ModuleOptions {
	moduleOptions := map[string]*autocliv1.ModuleOptions{}

	if appOpts.ModuleOptions != nil {
		maps.Copy(moduleOptions, appOpts.ModuleOptions)
	}

	for name, module := range appOpts.Modules {
		if _, ok := moduleOptions[name]; ok {
			continue
		}

		if withConfig, ok := module.(autocli.HasAutoCLIConfig); ok {
			moduleOptions[name] = withConfig.AutoCLIOptions()
		}
	}

	return moduleOptions
}

func wrapQueryServiceWithStreaming(cmd *cobra.Command, desc *autocliv1.ServiceCommandDescriptor, builder *autocli.Builder, registry *encoding.DynamicRegistry) error {
	if desc == nil {
		return nil
	}
	if registry == nil {
		return nil
	}

	for subName, subDesc := range desc.SubCommands {
		subCmd := findCommand(cmd, subName)
		if subCmd == nil {
			continue
		}
		if err := wrapQueryServiceWithStreaming(subCmd, subDesc, builder, registry); err != nil {
			return err
		}
	}

	if desc.Service == "" {
		return nil
	}

	serviceDescriptor, err := builder.FileResolver.FindDescriptorByName(protoreflect.FullName(desc.Service))
	if err != nil {
		return fmt.Errorf("resolve service %s: %w", desc.Service, err)
	}

	service := serviceDescriptor.(protoreflect.ServiceDescriptor)
	methodOptions := map[protoreflect.Name]*autocliv1.RpcCommandOptions{}
	for _, opt := range desc.RpcCommandOptions {
		methodOptions[protoreflect.Name(opt.RpcMethod)] = opt
	}

	for i := 0; i < service.Methods().Len(); i++ {
		method := service.Methods().Get(i)
		opts := methodOptions[method.Name()]
		if opts != nil && opts.Skip {
			continue
		}

		use := ""
		if opts != nil && opts.Use != "" {
			use = opts.Use
		} else {
			use = protoNameToCLIName(method.Name())
		}
		commandName := strings.SplitN(use, " ", 2)[0]
		methodCmd := findCommand(cmd, commandName)
		if methodCmd == nil {
			continue
		}

		fullMethod := fullMethodName(service, method)
		attachStreamCommand(methodCmd, method, opts, builder, registry, fullMethod)
	}

	return nil
}

func fullMethodName(service protoreflect.ServiceDescriptor, method protoreflect.MethodDescriptor) string {
	return "/" + string(service.FullName()) + "/" + string(method.Name())
}
