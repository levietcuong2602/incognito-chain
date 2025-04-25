package transaction

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/transaction/utils"
	utils2 "github.com/incognitochain/incognito-chain/utils"
	"github.com/pkg/errors"
)

const (
	UnexpectedError = iota
	WrongTokenTxTypeError
	CustomTokenExistedError
	WrongInputError
	WrongSigError
	DoubleSpendError
	TxNotExistError
	RandomCommitmentError
	InvalidSanityDataPRVError
	InvalidSanityDataPrivacyTokenError
	InvalidDoubleSpendPRVError
	InvalidDoubleSpendPrivacyTokenError
	InputCoinIsVeryLargeError
	PaymentInfoIsVeryLargeError
	SumInputCoinsAndOutputCoinsError
	InvalidInputCoinVersionErr
	TokenIDInvalidError
	TokenIDExistedError
	TokenIDExistedByCrossShardError
	PrivateKeySenderInvalidError
	SignTxError
	DecompressPaymentAddressError
	CanNotGetCommitmentFromIndexError
	CanNotDecompressCommitmentFromIndexError
	InitWithnessError
	WithnessProveError
	EncryptOutputError
	DecompressSigPubKeyError
	InitTxSignatureFromBytesError
	VerifyTxSigFailError
	DuplicatedOutputSndError
	SndExistedError
	InputCommitmentIsNotExistedError
	TxProofVerifyFailError
	VerifyOneOutOfManyProofFailedErr
	BatchTxProofVerifyFailError
	VerifyMinerCreatedTxBeforeGettingInBlockError
	CommitOutputCoinError
	GetShardIDByPublicKeyError

	NormalTokenPRVJsonError
	NormalTokenJsonError //nolint:revive

	PrivacyTokenInitFeeParamsError
	PrivacyTokenInitPRVError
	PrivacyTokenInitTokenDataError
	PrivacyTokenPRVJsonError
	PrivacyTokenJsonError //nolint:revive
	PrivacyTokenTxTypeNotHandleError

	ExceedSizeTx
	ExceedSizeInfoTxError
	ExceedSizeInfoOutCoinError

	RejectInvalidLockTime
	RejectTxSize
	RejectTxVersion
	RejectTxPublickeySigSize
	RejectTxType
	RejectTxInfoSize
	RejectTxMedataWithBlockChain

	GetCommitmentsInDatabaseError
	InvalidPaymentAddressError
	OnetimeAddressAlreadyExists
)

var ErrCodeMessage = utils.ErrCodeMessage

type TransactionError = utils.TransactionError //nolint:revive

func NewTransactionErr(key int, err error, params ...interface{}) *TransactionError {
	e := utils.NewTransactionErr(key, errors.Wrap(err, utils2.EmptyString), params)
	e.Message = ErrCodeMessage[key].Message
	if len(params) > 0 {
		e.Message = fmt.Sprintf(ErrCodeMessage[key].Message, params)
	}
	return e
}
