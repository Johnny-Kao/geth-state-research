package state

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
)

// researchS01Reader returns a fresh account for each underlying read.
// Any cross-reader mutation must therefore come from the shared cache.
type researchS01Reader struct{}

func (researchS01Reader) Account(common.Address) (*types.StateAccount, error) {
	return &types.StateAccount{Nonce: 7, Balance: uint256.NewInt(100)}, nil
}

func (researchS01Reader) Storage(common.Address, common.Hash) (common.Hash, error) {
	return common.Hash{}, nil
}

func TestResearchS01SharedReaderAccountIsolation(t *testing.T) {
	cache := newStateReaderWithCache(researchS01Reader{})
	addr := common.Address{0x01}

	first, err := cache.Account(addr)
	if err != nil {
		t.Fatal(err)
	}
	first.Nonce = 99

	second, err := cache.Account(addr)
	if err != nil {
		t.Fatal(err)
	}
	if second.Nonce != 7 {
		t.Fatalf("shared cache leaked caller mutation: got nonce %d, want 7", second.Nonce)
	}
}
