package account_message

import (
	"fmt"

	"github.com/crypto-com/chain-indexing/internal/tmcosmosutils"

	"github.com/crypto-com/chain-indexing/projection/account_message/view"

	projection_entity "github.com/crypto-com/chain-indexing/entity/projection"

	"github.com/crypto-com/chain-indexing/appinterface/projection/rdbprojectionbase"
	"github.com/crypto-com/chain-indexing/appinterface/rdb"
	event_entity "github.com/crypto-com/chain-indexing/entity/event"
	applogger "github.com/crypto-com/chain-indexing/internal/logger"
	"github.com/crypto-com/chain-indexing/internal/utctime"
	event_usecase "github.com/crypto-com/chain-indexing/usecase/event"
)

var _ projection_entity.Projection = &AccountMessage{}

type AccountMessage struct {
	*rdbprojectionbase.Base

	rdbConn rdb.Conn
	logger  applogger.Logger

	accountAddressPrefix string
}

func NewAccountMessage(
	logger applogger.Logger,
	rdbConn rdb.Conn,
	accountAddressPrefix string,
) *AccountMessage {
	return &AccountMessage{
		rdbprojectionbase.NewRDbBase(rdbConn.ToHandle(), "AccountMessage"),

		rdbConn,
		logger,

		accountAddressPrefix,
	}
}

func (_ *AccountMessage) GetEventsToListen() []string {
	return append([]string{
		event_usecase.BLOCK_CREATED,
	}, event_usecase.MSG_EVENTS...)
}

func (projection *AccountMessage) OnInit() error {
	return nil
}

func (projection *AccountMessage) HandleEvents(height int64, events []event_entity.Event) error {
	rdbTx, err := projection.rdbConn.Begin()
	if err != nil {
		return fmt.Errorf("error beginning transaction: %v", err)
	}

	committed := false
	defer func() {
		if !committed {
			_ = rdbTx.Rollback()
		}
	}()

	rdbTxHandle := rdbTx.ToHandle()

	// TODO: Handle genesis transaction
	if height == int64(0) {
		if err := projection.UpdateLastHandledEventHeight(rdbTxHandle, height); err != nil {
			return fmt.Errorf("error updating last handled event height: %v", err)
		}

		if err := rdbTx.Commit(); err != nil {
			return fmt.Errorf("error committing changes: %v", err)
		}
		committed = true
		return nil
	}

	accountMessagesView := view.NewAccountMessages(rdbTxHandle)
	accountMessagesTotalView := view.NewAccountMessagesTotal(rdbTxHandle)

	var blockTime utctime.UTCTime
	var blockHash string
	accountMessages := make([]view.AccountMessageRecord, 0)
	for _, event := range events {
		if blockCreatedEvent, ok := event.(*event_usecase.BlockCreated); ok {
			blockTime = blockCreatedEvent.Block.Time
			blockHash = blockCreatedEvent.Block.Hash
		}
	}

	for _, event := range events {
		if typedEvent, ok := event.(*event_usecase.MsgSend); ok {
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					typedEvent.FromAddress,
					typedEvent.ToAddress,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgMultiSend); ok {
			involvedAccounts := make([]string, 0)
			for _, input := range typedEvent.Inputs {
				involvedAccounts = append(involvedAccounts, input.Address)
			}
			for _, output := range typedEvent.Outputs {
				involvedAccounts = append(involvedAccounts, output.Address)
			}
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: involvedAccounts,
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgSetWithdrawAddress); ok {
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					typedEvent.DelegatorAddress,
					typedEvent.WithdrawAddress,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgWithdrawDelegatorReward); ok {
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					typedEvent.DelegatorAddress,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgWithdrawValidatorCommission); ok {
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					typedEvent.RecipientAddress,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgFundCommunityPool); ok {
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					typedEvent.Depositor,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgFundCommunityPool); ok {
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					typedEvent.Depositor,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgSubmitParamChangeProposal); ok {
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					typedEvent.ProposerAddress,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgSubmitCommunityPoolSpendProposal); ok {
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					typedEvent.ProposerAddress,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgSubmitSoftwareUpgradeProposal); ok {
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					typedEvent.ProposerAddress,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgSubmitCancelSoftwareUpgradeProposal); ok {
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					typedEvent.ProposerAddress,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgDeposit); ok {
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					typedEvent.Depositor,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgVote); ok {
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					typedEvent.Voter,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgCreateValidator); ok {
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					typedEvent.DelegatorAddress,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgEditValidator); ok {
			accountAddress := tmcosmosutils.MustAccountAddressFromValidatorAddress(
				projection.accountAddressPrefix,
				typedEvent.ValidatorAddress,
			)

			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					accountAddress,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgDelegate); ok {
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					typedEvent.DelegatorAddress,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgUndelegate); ok {
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					typedEvent.DelegatorAddress,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgBeginRedelegate); ok {
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					typedEvent.DelegatorAddress,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgUnjail); ok {
			accountAddress := tmcosmosutils.MustAccountAddressFromValidatorAddress(
				projection.accountAddressPrefix,
				typedEvent.ValidatorAddr,
			)

			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					accountAddress,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgNFTIssueDenom); ok {
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					typedEvent.Sender,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgNFTMintNFT); ok {
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					typedEvent.Sender,
					typedEvent.Recipient,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgNFTTransferNFT); ok {
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					typedEvent.Sender,
					typedEvent.Recipient,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgNFTEditNFT); ok {
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					typedEvent.Sender,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgNFTBurnNFT); ok {
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					typedEvent.Sender,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgIBCCreateClient); ok {
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					typedEvent.Params.Signer,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgIBCUpdateClient); ok {
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					typedEvent.Params.Signer,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgIBCConnectionOpenInit); ok {
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					typedEvent.Params.Signer,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgIBCConnectionOpenAck); ok {
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					typedEvent.Params.Signer,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgIBCConnectionOpenTry); ok {
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					typedEvent.Params.Signer,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgIBCConnectionOpenConfirm); ok {
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					typedEvent.Params.Signer,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgIBCChannelOpenInit); ok {
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					typedEvent.Params.Signer,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgIBCChannelOpenAck); ok {
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					typedEvent.Params.Signer,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgIBCChannelOpenTry); ok {
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					typedEvent.Params.Signer,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgIBCChannelOpenConfirm); ok {
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					typedEvent.Params.Signer,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgIBCAcknowledgement); ok {
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					typedEvent.Params.Signer,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgIBCRecvPacket); ok {
			if typedEvent.Params.MaybeFungibleTokenPacketData != nil {
				accountMessages = append(accountMessages, view.AccountMessageRecord{
					Row: view.AccountMessageRow{
						BlockHeight:     height,
						BlockHash:       "",
						BlockTime:       utctime.UTCTime{},
						TransactionHash: typedEvent.TxHash(),
						Success:         typedEvent.TxSuccess(),
						MessageIndex:    typedEvent.MsgIndex,
						MessageType:     typedEvent.MsgType(),
						Data:            typedEvent,
					},
					Accounts: []string{
						typedEvent.Params.MaybeFungibleTokenPacketData.Receiver,
					},
				})
			}
		} else if typedEvent, ok := event.(*event_usecase.MsgIBCTransferTransfer); ok {
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					typedEvent.Params.Sender,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgIBCTimeout); ok {
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					typedEvent.Params.Signer,
				},
			})
		} else if typedEvent, ok := event.(*event_usecase.MsgIBCTimeoutOnClose); ok {
			accountMessages = append(accountMessages, view.AccountMessageRecord{
				Row: view.AccountMessageRow{
					BlockHeight:     height,
					BlockHash:       "",
					BlockTime:       utctime.UTCTime{},
					TransactionHash: typedEvent.TxHash(),
					Success:         typedEvent.TxSuccess(),
					MessageIndex:    typedEvent.MsgIndex,
					MessageType:     typedEvent.MsgType(),
					Data:            typedEvent,
				},
				Accounts: []string{
					typedEvent.Params.Signer,
				},
			})
		}
	}

	for i, accountMessage := range accountMessages {
		// TODO: Change to use InsertAll
		accountMessages[i].Row.BlockHash = blockHash
		accountMessages[i].Row.BlockTime = blockTime

		insertedAccounts := make(map[string]bool)
		deduplicatedAccounts := make([]string, 0)
		for _, involvedAccount := range accountMessage.Accounts {
			// Deduplication
			if _, exist := insertedAccounts[involvedAccount]; exist {
				continue
			}

			if err := accountMessagesTotalView.Increment(fmt.Sprintf("%s:-", involvedAccount), 1); err != nil {
				return fmt.Errorf("error incremnting total account message of account: %w", err)
			}
			if err := accountMessagesTotalView.Increment(
				fmt.Sprintf("%s:%s", involvedAccount, accountMessage.Row.MessageType), 1,
			); err != nil {
				return fmt.Errorf("error incremnting total account message of account: %w", err)
			}
			deduplicatedAccounts = append(deduplicatedAccounts, involvedAccount)
			insertedAccounts[involvedAccount] = true
		}

		if err := accountMessagesView.Insert(&accountMessages[i].Row, deduplicatedAccounts); err != nil {
			return fmt.Errorf("error inserting account message: %w", err)
		}
	}

	if err := projection.UpdateLastHandledEventHeight(rdbTxHandle, height); err != nil {
		return fmt.Errorf("error updating last handled event height: %v", err)
	}

	if err := rdbTx.Commit(); err != nil {
		return fmt.Errorf("error committing changes: %v", err)
	}
	committed = true
	return nil
}
