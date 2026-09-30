package stream

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/reflect/protoreflect"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/client/v2/autocli"
	flagbuilder "cosmossdk.io/client/v2/autocli/flag"
	errorsmod "cosmossdk.io/errors"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/CosmosContracts/juno/v31/x/stream/types/encoding"
)

type methodContext struct {
	builder    *autocli.Builder
	method     protoreflect.MethodDescriptor
	opts       *autocliv1.RpcCommandOptions
	registry   *encoding.DynamicRegistry
	fullMethod string
}

func (m methodContext) buildBinder() (*flagbuilder.MessageBinder, *pflag.FlagSet, error) {
	if m.builder == nil || m.method == nil {
		return nil, nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "autocli builder and method are required")
	}
	if m.builder.TypeResolver == nil {
		return nil, nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "type resolver not configured")
	}

	inputType, err := m.builder.TypeResolver.FindMessageByName(m.method.Input().FullName())
	if err != nil {
		return nil, nil, fmt.Errorf("resolve input type for %s: %w", m.method.FullName(), err)
	}

	flagSet := pflag.NewFlagSet(string(m.method.FullName()), pflag.ContinueOnError)
	flagSet.ParseErrorsAllowlist.UnknownFlags = true

	ctx := context.Background()
	opts := m.opts
	if opts == nil {
		opts = &autocliv1.RpcCommandOptions{}
	}

	binder, err := m.builder.AddMessageFlags(&ctx, flagSet, inputType, opts)
	if err != nil {
		return nil, nil, err
	}

	return binder, flagSet, nil
}

func (m methodContext) descriptor() (*encoding.MethodDescriptor, error) {
	if m.registry == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidType, "stream registry is not available")
	}

	desc, ok := m.registry.LookupByFullMethod(m.fullMethod)
	if !ok || desc == nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "stream descriptor not registered: %s", m.fullMethod)
	}

	return desc, nil
}

func attachStreamCommand(
	cmd *cobra.Command,
	method protoreflect.MethodDescriptor,
	opts *autocliv1.RpcCommandOptions,
	builder *autocli.Builder,
	registry *encoding.DynamicRegistry,
	methodName string,
) {
	if cmd == nil {
		return
	}

	ctx := methodContext{
		builder:    builder,
		method:     method,
		opts:       opts,
		registry:   registry,
		fullMethod: methodName,
	}

	origRun, origRunE := cmd.Run, cmd.RunE
	cmd.Run = nil
	cmd.RunE = func(current *cobra.Command, args []string) error {
		useStream, err := resolveStreamFlag(current)
		if err != nil {
			return err
		}

		if !useStream {
			if origRunE != nil {
				return origRunE(current, args)
			}
			if origRun != nil {
				origRun(current, args)
			}
			return nil
		}

		binder, flagSet, err := ctx.buildBinder()
		if err != nil {
			return err
		}

		if err := populateBinderFlags(current, flagSet); err != nil {
			return err
		}

		input, err := binder.BuildMessage(args)
		if err != nil {
			return err
		}

		desc, err := ctx.descriptor()
		if err != nil {
			return fmt.Errorf("resolve streaming descriptor for %s: %w", ctx.fullMethod, err)
		}

		params, err := encoding.ExtractParams(desc, input)
		if err != nil {
			return fmt.Errorf("prepare streaming params for %s: %w", ctx.fullMethod, err)
		}

		streamCtx, err := buildStreamingClientContext(current)
		if err != nil {
			return err
		}

		return runDynamicStreamCommand(current, streamCtx, desc, params)
	}
}

func populateBinderFlags(cmd *cobra.Command, dest *pflag.FlagSet) error {
	var firstErr error

	visit := func(set *pflag.FlagSet) {
		set.Visit(func(flag *pflag.Flag) {
			if firstErr != nil {
				return
			}

			if destFlag := dest.Lookup(flag.Name); destFlag == nil {
				return
			}

			if err := dest.Set(flag.Name, flag.Value.String()); err != nil {
				firstErr = err
			}
		})
	}

	visit(cmd.Flags())
	visit(cmd.InheritedFlags())

	return firstErr
}

func resolveStreamFlag(cmd *cobra.Command) (bool, error) {
	if cmd.Flags().Lookup(StreamFlagName) != nil {
		return cmd.Flags().GetBool(StreamFlagName)
	}
	if cmd.InheritedFlags().Lookup(StreamFlagName) != nil {
		return cmd.InheritedFlags().GetBool(StreamFlagName)
	}
	return false, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "stream flag is not registered on command %s", cmd.CommandPath())
}
