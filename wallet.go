package wallet

import (
	"context"
	"encoding/hex"
	"io/ioutil"
	"math"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	stablecoin "github.com/savechainwallet/savex-zk-evm/erc20"
	"golang.org/x/crypto/sha3"
)

// Wallet that
type StableCoinWallet struct {
	client        *ethclient.Client
	tokenAddress  common.Address
	lockAddress   common.Address
	walletAddress common.Address
	decimals      uint
}

// Get erc20 token balance
func (s *StableCoinWallet) Balance() (float64, error) {

	token, err := stablecoin.NewContracts(s.tokenAddress, s.client)
	if err != nil {
		return 0, err
	}
	balance, err := token.BalanceOf(&bind.CallOpts{}, s.walletAddress)
	if err != nil {
		return 0, err
	}
	balanceDecimals := float64(balance.Int64()) / math.Pow(10, float64(s.decimals))
	return balanceDecimals, nil
}

// Make unsigned raw transaction to sign with client
func (s *StableCoinWallet) MakeERC20Transaction(toAddress common.Address, value float64) (string, error) {
	transferFnSignature := []byte("transfer(address,uint256)") // do not include spaces in the string
	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)
	methodID := hash.Sum(nil)[:4]
	paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)

	amount := FloatToBigInt(value, int(s.decimals))

	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)
	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)

	nonce, err := s.client.PendingNonceAt(context.Background(), s.walletAddress)
	if err != nil {
		return "", err
	}
	gasLimit := uint64(80000)

	gasPrice, err := s.client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", err
	}

	tx := types.NewTransaction(nonce, s.tokenAddress, big.NewInt(0), gasLimit, gasPrice, data)

	ts := types.Transactions{tx}
	rawTxBytes, _ := rlp.EncodeToBytes(ts[0])
	rawTxHex := hex.EncodeToString(rawTxBytes)
	return rawTxHex, nil

}

// Send signed transaction to chain
// Returns TX Hash
func (s *StableCoinWallet) SendRawTX(tx string) (string, error) {
	rawTxBytes, err := hex.DecodeString(tx[2:])
	if err != nil {
		return "", err
	}
	transaction := new(types.Transaction)

	err = rlp.DecodeBytes(rawTxBytes, &transaction)
	if err != nil {
		return "", err
	}
	err = s.client.SendTransaction(context.Background(), transaction)
	if err != nil {
		return "", err
	}

	return transaction.Hash().Hex(), nil

}

// Format transaction to send locked payment with SaveXLock contract
func (s *StableCoinWallet) FormatLockedTX(toAddress common.Address, value float64) (string, error) {
	abiJSON, err := ioutil.ReadFile("./abi/savexlock.json")
	if err != nil {
		return "", err
	}

	contractABI, err := abi.JSON(strings.NewReader(string(abiJSON)))
	if err != nil {
		return "", err
	}
	gasPrice, err := s.client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", err
	}
	data, err := contractABI.Pack("createPayment", toAddress, value)
	if err != nil {
		return "", err
	}
	nonce, err := s.client.PendingNonceAt(context.Background(), s.walletAddress)
	if err != nil {
		return "", err
	}
	tx := types.NewTransaction(nonce, s.tokenAddress, big.NewInt(0), 20000, gasPrice, data)

	ts := types.Transactions{tx}
	rawTxBytes, _ := rlp.EncodeToBytes(ts[0])
	rawTxHex := hex.EncodeToString(rawTxBytes)
	return rawTxHex, nil
}

// Format transaction to cancel locked payment
func (s *StableCoinWallet) FormatCancelTX(paymentIndex big.Int) (string, error) {
	abiJSON, err := ioutil.ReadFile("./abi/savexlock.json")
	if err != nil {
		return "", err
	}

	contractABI, err := abi.JSON(strings.NewReader(string(abiJSON)))
	if err != nil {
		return "", err
	}
	gasPrice, err := s.client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", err
	}
	data, err := contractABI.Pack("cancelPayment", paymentIndex)
	if err != nil {
		return "", err
	}
	nonce, err := s.client.PendingNonceAt(context.Background(), s.walletAddress)
	if err != nil {
		return "", err
	}
	tx := types.NewTransaction(nonce, s.tokenAddress, big.NewInt(0), 20000, gasPrice, data)

	ts := types.Transactions{tx}
	rawTxBytes, _ := rlp.EncodeToBytes(ts[0])
	rawTxHex := hex.EncodeToString(rawTxBytes)
	return rawTxHex, nil
}

// Format transaction to cancel locked payment
func (s *StableCoinWallet) FormatWithdrawlTX(paymentIndex big.Int) (string, error) {
	abiJSON, err := ioutil.ReadFile("./abi/savexlock.json")
	if err != nil {
		return "", err
	}

	contractABI, err := abi.JSON(strings.NewReader(string(abiJSON)))
	if err != nil {
		return "", err
	}
	gasPrice, err := s.client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", err
	}
	data, err := contractABI.Pack("withdrawPayment", paymentIndex)
	if err != nil {
		return "", err
	}
	nonce, err := s.client.PendingNonceAt(context.Background(), s.walletAddress)
	if err != nil {
		return "", err
	}
	tx := types.NewTransaction(nonce, s.tokenAddress, big.NewInt(0), 20000, gasPrice, data)

	ts := types.Transactions{tx}
	rawTxBytes, _ := rlp.EncodeToBytes(ts[0])
	rawTxHex := hex.EncodeToString(rawTxBytes)
	return rawTxHex, nil
}

func FloatToBigInt(val float64, decimals int) *big.Int {
	bigval := new(big.Float)
	bigval.SetFloat64(val)
	// Set precision if required.
	// bigval.SetPrec(64)

	coin := new(big.Float)
	coin.SetInt(big.NewInt(int64(math.Pow(10, float64(decimals)))))

	bigval.Mul(bigval, coin)
	result := new(big.Int)
	bigval.Int(result) // store converted number in result

	return result
}
