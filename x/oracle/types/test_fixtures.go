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

func AddTrackingPriceHistoryWithAcceptListProposalFixture(
	mutators ...func(p *AddTrackingPriceHistoryWithAcceptListProposal),
) *AddTrackingPriceHistoryWithAcceptListProposal {
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

	p := &AddTrackingPriceHistoryWithAcceptListProposal{
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
	removeTrackingList := DenomList{
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
		Title:              "Foo",
		Description:        "Bar",
		RemoveTrackingList: removeTrackingList,
	}

	for _, m := range mutators {
		m(p)
	}

	return p
}
