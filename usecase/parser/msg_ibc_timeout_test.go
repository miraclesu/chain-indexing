package parser_test

import (
	"regexp"
	"strings"

	"github.com/crypto-com/chain-indexing/internal/json"

	"github.com/crypto-com/chain-indexing/usecase/parser/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/crypto-com/chain-indexing/infrastructure/tendermint"
	"github.com/crypto-com/chain-indexing/usecase/event"
	"github.com/crypto-com/chain-indexing/usecase/parser"
	usecase_parser_test "github.com/crypto-com/chain-indexing/usecase/parser/test"
)

var _ = Describe("ParseMsgCommands", func() {
	Describe("MsgIBCTimeout", func() {
		It("should parse Msg commands when there is MsgIBCTimeout in the transaction", func() {
			expected := `{
  "name": "MsgTimeoutCreated",
  "version": 1,
  "height": 643189,
  "uuid": "{UUID}",
  "msgName": "MsgTimeout",
  "txHash": "E86F52B60F3EC63EEAA9B77DB41BE90E99B7E1230BE94F3A7E4C06B7C7A20C2D",
  "msgIndex": 1,
  "params": {
	"packet": {
  	"sequence": "5",
  	"sourcePort": "transfer",
  	"sourceChannel": "channel-9",
  	"destinationPort": "transfer",
  	"destinationChannel": "channel-109",
  	"data": "eyJhbW91bnQiOiIxIiwiZGVub20iOiJiYXNlY3JvIiwicmVjZWl2ZXIiOiJjb3Ntb3Mxenp6NXhjd3l6bWs5dHBreTR0ZTkwM3ZzeGg1NXNmc2x2ZnplY2YiLCJzZW5kZXIiOiJjcm8xczdjdTI4NDAzZ3pkdnk1dHR0eXNrbTN6eGplanhjdjYzZXNwcmUifQ==",
  	"timeoutHeight": {
  	  "revisionNumber": "4",
  	  "revisionHeight": "6182017"
  	},
  	"timeoutTimestamp": "1620753450655319559"
    },
    "proofUnreceived": "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",
    "proofHeight": {
      "revisionNumber": "4",
      "revisionHeight": "6185877"
    },
    "nextSequenceRecv": "5",
    "signer": "cro1q040rm026jmpfmxdsj6q9phm9tdceepnsau6me",

    "application": "transfer",
    "messageType": "MsgTransfer",
    "maybeMsgTransfer": {
      "refundReceiver": "cro1s7cu28403gzdvy5tttyskm3zxjejxcv63espre",
      "refundDenom": "basecro",
      "refundAmount": "1"
    },

    "packetTimeoutHeight": {
      "revisionNumber": "4",
      "revisionHeight": "6182017"
    },
    "packetTimeoutTimestamp": "1620753450655319559",
    "packetSequence": "5",
    "channelOrdering": "ORDER_UNORDERED"
  }
}
`

			txDecoder := utils.NewTxDecoder()
			block, _, _ := tendermint.ParseBlockResp(strings.NewReader(
				usecase_parser_test.TX_MSG_TIMEOUT_BLOCK_RESP,
			))
			blockResults, _ := tendermint.ParseBlockResultsResp(strings.NewReader(
				usecase_parser_test.TX_MSG_TIMEOUT_BLOCK_RESULTS_RESP,
			))

			accountAddressPrefix := "cro"
			stakingDenom := "basecro"
			cmds, err := parser.ParseBlockResultsTxsMsgToCommands(
				txDecoder,
				block,
				blockResults,
				accountAddressPrefix,
				stakingDenom,
			)
			Expect(err).To(BeNil())
			Expect(cmds).To(HaveLen(4))
			cmd := cmds[3]
			Expect(cmd.Name()).To(Equal("CreateMsgIBCTimeout"))

			untypedEvent, _ := cmd.Exec()
			typedEvent := untypedEvent.(*event.MsgIBCTimeout)

			regex, _ := regexp.Compile("\n?\r?\\s?")

			Expect(json.MustMarshalToString(typedEvent)).To(Equal(
				strings.Replace(
					regex.ReplaceAllString(expected, ""),
					"{UUID}",
					typedEvent.UUID(),
					-1,
				),
			))
		})
	})
})
