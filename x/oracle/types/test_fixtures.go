package types

func AddTrackingPriceHistoryProposalFixture(mutators ...func(p *AddTrackingPriceHistoryProposal)) *AddTrackingPriceHistoryProposal {
	var trackingList = DenomList{
		{
			BaseDenom:   JunoDenom,
			SymbolDenom: JunoSymbol,
			Exponent:    JunoExponent,
		},
		{
			BaseDenom:   AtomDenom,
			SymbolDenom: AtomSymbol,
			Exponent:    AtomExponent,
		},
	}

	p := &AddTrackingPriceHistoryProposal{
		Title:        "Foo",
		Description:  "Bar",
		TrackingList: trackingList,
	}

	for _, m := range mutators {
		m(p)
	}

	return p
}
