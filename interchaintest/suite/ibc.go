package suite

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
)

func (s *E2ETestSuite) SendIBCTransfer(
	chain *cosmos.CosmosChain,
	channelID string,
	keyName string,
	amount ibc.WalletAmount,
	options ibc.TransferOptions,
) (tx ibc.Tx, _ error) {
	port := "transfer"
	if options.Port != "" {
		port = options.Port
	}
	command := []string{
		"ibc-transfer", "transfer", port, channelID,
		amount.Address, fmt.Sprintf("%s%s", amount.Amount.String(), amount.Denom),
		"--gas", "auto",
	}
	if options.Timeout != nil {
		if options.Timeout.NanoSeconds > 0 {
			command = append(command, "--packet-timeout-timestamp", fmt.Sprint(options.Timeout.NanoSeconds))
		}

		if options.Timeout.Height > 0 {
			command = append(command, "--packet-timeout-height", fmt.Sprintf("0-%d", options.Timeout.Height))
		}

		if options.AbsoluteTimeouts {
			// ibc-go doesn't support relative heights for packet timeouts
			// so the absolute height flag must be manually set:
			command = append(command, "--absolute-timeouts")
		}
	}
	if options.Memo != "" {
		command = append(command, "--memo", options.Memo)
	}
	txHash, err := s.ExecTx(chain, keyName, false, false, command...)

	if err != nil {
		return tx, fmt.Errorf("send ibc transfer: %w", err)
	}
	txResp, err := chain.GetTransaction(txHash)
	if err != nil {
		return tx, fmt.Errorf("failed to get transaction %s: %w", txHash, err)
	}
	if txResp.Code != 0 {
		return tx, fmt.Errorf("error in transaction (code: %d): %s", txResp.Code, txResp.RawLog)
	}
	tx.Height = txResp.Height
	tx.TxHash = txHash
	// In cosmos, user is charged for entire gas requested, not the actual gas used.
	tx.GasSpent = txResp.GasWanted

	const evType = "send_packet"
	events := txResp.Events

	var (
		seq, _           = AttributeValue(events, evType, "packet_sequence")
		srcPort, _       = AttributeValue(events, evType, "packet_src_port")
		srcChan, _       = AttributeValue(events, evType, "packet_src_channel")
		dstPort, _       = AttributeValue(events, evType, "packet_dst_port")
		dstChan, _       = AttributeValue(events, evType, "packet_dst_channel")
		timeoutHeight, _ = AttributeValue(events, evType, "packet_timeout_height")
		timeoutTS, _     = AttributeValue(events, evType, "packet_timeout_timestamp")
		dataHex, _       = AttributeValue(events, evType, "packet_data_hex")
	)
	tx.Packet.SourcePort = srcPort
	tx.Packet.SourceChannel = srcChan
	tx.Packet.DestPort = dstPort
	tx.Packet.DestChannel = dstChan
	tx.Packet.TimeoutHeight = timeoutHeight

	data, err := hex.DecodeString(dataHex)
	if err != nil {
		return tx, fmt.Errorf("malformed data hex %s: %w", dataHex, err)
	}
	tx.Packet.Data = data

	seqNum, err := strconv.ParseUint(seq, 10, 64)
	if err != nil {
		return tx, fmt.Errorf("invalid packet sequence from events %s: %w", seq, err)
	}
	tx.Packet.Sequence = seqNum

	timeoutNano, err := strconv.ParseUint(timeoutTS, 10, 64)
	if err != nil {
		return tx, fmt.Errorf("invalid packet timestamp timeout %s: %w", timeoutTS, err)
	}
	tx.Packet.TimeoutTimestamp = ibc.Nanoseconds(timeoutNano)

	return tx, nil
}
