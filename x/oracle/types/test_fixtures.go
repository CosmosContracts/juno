package types

func AddTrackingPriceHistoryProposalFixture(
	mutators ...func(p *AddTrackingPriceHistoryProposal),
) *AddTrackingPriceHistoryProposal {
	trackingList := DenomList{
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

func AddTrackingPriceHistoryWithWhitelistProposalFixture(
	mutators ...func(p *AddTrackingPriceHistoryWithWhitelistProposal),
) *AddTrackingPriceHistoryWithWhitelistProposal {
	trackingList := DenomList{
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

	p := &AddTrackingPriceHistoryWithWhitelistProposal{
		Title:        "Foo",
		Description:  "Bar",
		TrackingList: trackingList,
	}

	for _, m := range mutators {
		m(p)
	}

	return p
}

func RemoveTrackingPriceHistoryProposalFixture(
	mutators ...func(p *RemoveTrackingPriceHistoryProposal),
) *RemoveTrackingPriceHistoryProposal {
	removeTwapList := DenomList{
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

	p := &RemoveTrackingPriceHistoryProposal{
		Title:          "Foo",
		Description:    "Bar",
		RemoveTwapList: removeTwapList,
	}

	for _, m := range mutators {
		m(p)
	}

	return p
}
