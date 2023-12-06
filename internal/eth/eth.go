package eth

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/chenzhijie/go-web3"
	"github.com/ethereum/go-ethereum/common"
)

var (
	ErrSecretNotString = errors.New("secret is not string")
	ErrOwnerNotString  = errors.New("owner is not address")
)

func Check(address string, tokenID int, token string) (bool, error) {
	w3, err := web3.NewWeb3("https://rpc.sepolia.org")
	if err != nil {
		return false, fmt.Errorf("connecting to sepolia: %w", err)
	}

	contract, err := w3.Eth.NewContract(abiStr, contractAddr)
	if err != nil {
		return false, fmt.Errorf("creating contract client: %w", err)
	}

	secret, err := contract.Call("secret", big.NewInt(int64(tokenID)))
	if err != nil {
		return false, fmt.Errorf("getting secret for token: %w", err)
	}

	secretStr, ok := secret.(string)
	if !ok {
		return false, ErrSecretNotString
	}

	if secretStr != token {
		return false, nil
	}

	owner, err := contract.Call("ownerOf", big.NewInt(int64(tokenID)))
	if err != nil {
		return false, fmt.Errorf("getting owner of token: %w", err)
	}

	ownerAddr, ok := owner.(common.Address)
	if !ok {
		return false, ErrOwnerNotString
	}

	if ownerAddr.String() == address {
		return true, nil
	}

	return false, nil
}
