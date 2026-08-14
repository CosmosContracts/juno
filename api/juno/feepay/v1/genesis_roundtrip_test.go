package feepayv1

import (
	"testing"

	"google.golang.org/protobuf/proto"
)

func TestGenesisStateWalletUsagesGeneratedRoundTrip(t *testing.T) {
	usage := &FeePayWalletUsage{
		ContractAddress: "juno1contract",
		WalletAddress:   "juno1wallet",
		Uses:            7,
	}
	original := &GenesisState{WalletUsages: []*FeePayWalletUsage{usage}}

	field := original.ProtoReflect().Descriptor().Fields().ByName("wallet_usages")
	if field == nil {
		t.Fatal("wallet_usages missing from generated descriptor")
	}
	if field.Number() != 3 || !field.IsList() || field.Message().FullName() != "juno.feepay.v1.FeePayWalletUsage" {
		t.Fatalf("unexpected wallet_usages descriptor: number=%d list=%t message=%s", field.Number(), field.IsList(), field.Message().FullName())
	}
	if got := original.GetWalletUsages(); len(got) != 1 || got[0].GetUses() != 7 {
		t.Fatalf("generated getter lost wallet usage: %#v", got)
	}

	wire, err := proto.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var restored GenesisState
	if err := proto.Unmarshal(wire, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !proto.Equal(original, &restored) {
		t.Fatalf("wallet usages did not round trip: got %#v", restored.GetWalletUsages())
	}
}
