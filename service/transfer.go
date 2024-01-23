package service

import (
	"context"
	"math/big"
	"orders-manager/config"
	"orders-manager/models"
	"orders-manager/service/abi"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	logger "github.com/sirupsen/logrus"
)

func getClient() (*ethclient.Client, error) {
	var client *ethclient.Client
	var err error
	provider := config.ReadEnvString("PROVIDER")
	client, err = ethclient.Dial(provider)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func getTransactor(client *ethclient.Client, privateKey string, estimatedGas uint64, chainID int) (auth *bind.TransactOpts, err error) {

	privKey, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		logger.Panicf("generate private key error: %v", err)
	}

	auth, err = bind.NewKeyedTransactorWithChainID(privKey, big.NewInt(int64(chainID)))
	if err != nil {
		return nil, err
	}

	gasPrice := big.NewInt(150000000)

	// gasPrice = gasPrice.Mul(gasPrice, big.NewInt(6))
	// gasPrice = gasPrice.Div(gasPrice, big.NewInt(5))
	// auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0) // in wei
	auth.GasLimit = 150000000  // in units
	auth.GasPrice = gasPrice
	return auth, nil
}

func TransferTokens(orderMatches []*models.OrderMatch) (err error) {
	client, err := getClient()
	if err != nil {
		logger.Errorf("error in getting ethereum client %s", err.Error())
		return err
	}

	privateKey := config.ReadEnvString("PRIVATE_KEY")

	estimatedGas, error := client.EstimateGas(context.Background(), ethereum.CallMsg{})
	if error != nil {
		logger.Errorf("error in estimating gas %s", error.Error())
		return error
	}
	chainID := config.ReadEnvInt("CHAIN_ID")

	transactor, err := getTransactor(client, privateKey, estimatedGas, chainID)
	if err != nil {
		logger.Errorf("error in getting transactor %s", err.Error())
		return err
	}
	for _, orderMatch := range orderMatches {

		BaseErc20, err := abi.NewERC20(common.HexToAddress(orderMatch.MakeOrder.BaseAsset().VirtualToken), client)
		if err != nil {
			logger.Errorf("error in getting erc20 contract %s", err.Error())
			return err
		}

		QuoteErc20, err := abi.NewERC20(common.HexToAddress(orderMatch.MakeOrder.QuoteAsset().VirtualToken), client)
		if err != nil {
			logger.Errorf("error in getting erc20 contract %s", err.Error())
			return err
		}

		makerTrader := common.HexToAddress(orderMatch.MakeOrder.Trader)
		takerTrader := common.HexToAddress(orderMatch.TakeOrder.Trader)

		fillsFloat, _ := orderMatch.NewFills.Float64()
		var takerSendsAmount, makerSendsAmount *big.Int

		var transaction1, transaction2 *types.Transaction

		if orderMatch.TakeOrder.IsUpForSale {
			takerSendsAmount = orderMatch.NewFills
			makerSendsAmount = big.NewInt(int64(fillsFloat / float64(orderMatch.TakeOrder.Price)))

			transaction1, err = BaseErc20.TransferFrom(transactor, takerTrader, makerTrader, takerSendsAmount)
			if err != nil {
				logger.Errorf("error in transferring tokens from Address: %s to Address: %s, due to error: %v", takerTrader, makerTrader, err)
				return err
			}

			transaction2, err = QuoteErc20.TransferFrom(transactor, makerTrader, takerTrader, makerSendsAmount)
			if err != nil {
				logger.Errorf("error in transferring tokens from Address: %s to Address: %s, due to error: %v", makerTrader, takerTrader, err)
				return err
			}
		} else {
			takerSendsAmount = big.NewInt(int64(fillsFloat / float64(orderMatch.MakeOrder.Price)))
			makerSendsAmount = orderMatch.NewFills

			transaction1, err = QuoteErc20.TransferFrom(transactor, takerTrader, makerTrader, takerSendsAmount)
			if err != nil {
				logger.Errorf("error in transferring tokens from Address: %s to Address: %s, due to error: %v", takerTrader, makerTrader, err)
				return err
			}

			transaction2, err = BaseErc20.TransferFrom(transactor, makerTrader, takerTrader, makerSendsAmount)
			if err != nil {
				logger.Errorf("error in transferring tokens from Address: %s to Address: %s, due to error: %v", makerTrader, takerTrader, err)
				return err
			}

		}

		// Update the order status
		err = models.UpdateOrder(orderMatch.MakeOrder)
		if err != nil {
			logger.Errorf("Error updating order %d: %s", orderMatch.MakeOrder.OrderID, err.Error())
			continue
		}

		err = models.UpdateOrder(orderMatch.TakeOrder)
		if err != nil {
			logger.Errorf("Error updating order %d: %s", orderMatch.TakeOrder.OrderID, err.Error())
			continue
		}

		logger.Infof("Tokens transferred between Address: %s and Address: %s with transaction hash 1: %s and transaction hash 2: %s", takerTrader, makerTrader, transaction1.Hash().Hex(), transaction2.Hash().Hex())
	}
	client.Close()
	return nil
}
