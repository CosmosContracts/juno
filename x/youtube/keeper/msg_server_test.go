package keeper_test

import (
	_ "embed"
	"fmt"

	"github.com/CosmosContracts/juno/v19/x/youtube/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
)

//go:embed download.jpeg
var imageContent []byte

func (s *IntegrationTestSuite) TestFullUpload() {
	_, _, creator := testdata.KeyTestPubAddr()

	// TODO: content should be pre-processed into reduced bytes.
	// only pixels that change actually need updating. Else use the previous frame values if 0 is provided (if 0,0,0 is actually used in the video, make it 0,0,1)
	// this means a single black image upload = only 1 frame. Audio? unsure

	// This would be xz compressed video bytes. Will store multiple frames in a single content that is cut by some unique character
	content := make([][]byte, 0)
	for i := 0; i < 1000; i++ {
		content = append(content, []byte(imageContent))
	}

	for idx, c := range content {
		s.msgServer.Upload(s.ctx, &types.MsgUploadContentBlob{
			Sender:  creator.String(),
			IdKey:   uint64(idx),
			Content: c,
		})
	}

	// upload metadata
	title := "this is my long test title!"
	s.msgServer.UploadMetadata(s.ctx, &types.MsgUploadMetadata{
		Sender:      creator.String(),
		Title:       title,
		Description: "test",
		IdStart:     0,
		IdEnd:       uint64(len(content) - 1),
	})

	// get this metadata
	metadata := s.k.GetMetadata(s.ctx, creator, title)
	fmt.Println("Content Found. metadata:")
	fmt.Println(metadata)

	// get back values from this metadata
	// for idx := range content {
	// 	// get this data from the store with k.GetContent
	// 	bz := s.k.GetContent(s.ctx, creator, uint64(idx))
	// 	fmt.Println(len(bz))
	// }

	// TODO: IdStart may always be 0 if using title / sender as unique key
	for i := metadata.IdStart; i <= metadata.IdEnd; i++ {
		// get this data from the store with k.GetContent
		bz := s.k.GetContent(s.ctx, creator, uint64(i))
		fmt.Println("img bytes: ", i, len(bz))
	}

}

// func (s *IntegrationTestSuite) TestUpdateClockParams() {
// 	_, _, addr := testdata.KeyTestPubAddr()
// 	_, _, addr2 := testdata.KeyTestPubAddr()

// 	for _, tc := range []struct {
// 		desc              string
// 		isEnabled         bool
// 		ContractAddresses []string
// 		success           bool
// 	}{
// 		{
// 			desc:              "Success - Valid on",
// 			isEnabled:         true,
// 			ContractAddresses: []string{},
// 			success:           true,
// 		},
// 		{
// 			desc:              "Success - Valid off",
// 			isEnabled:         false,
// 			ContractAddresses: []string{},
// 			success:           true,
// 		},
// 		{
// 			desc:              "Success - On and 1 allowed address",
// 			isEnabled:         true,
// 			ContractAddresses: []string{addr.String()},
// 			success:           true,
// 		},
// 		{
// 			desc:              "Fail - On and 2 duplicate addresses",
// 			isEnabled:         true,
// 			ContractAddresses: []string{addr.String(), addr.String()},
// 			success:           false,
// 		},
// 		{
// 			desc:              "Success - On and 2 unique",
// 			isEnabled:         true,
// 			ContractAddresses: []string{addr.String(), addr2.String()},
// 			success:           true,
// 		},
// 		{
// 			desc:              "Success - On and 2 duplicate 1 unique",
// 			isEnabled:         true,
// 			ContractAddresses: []string{addr.String(), addr2.String(), addr.String()},
// 			success:           false,
// 		},
// 	} {
// 		tc := tc
// 		s.Run(tc.desc, func() {
// 			params := types.DefaultParams()
// 			params.ContractAddresses = tc.ContractAddresses

// 			err := s.app.AppKeepers.ClockKeeper.SetParams(s.ctx, params)

// 			if !tc.success {
// 				s.Require().Error(err)
// 			} else {
// 				s.Require().NoError(err)
// 			}
// 		})
// 	}
// }
