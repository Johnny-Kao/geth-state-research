package state

import (
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
)

// researchS01Reader returns a fresh account for each underlying read.
// Cross-reader mutation must therefore come from the shared cache.
type researchS01Reader struct{}

func (researchS01Reader) Account(common.Address) (*types.StateAccount, error) {
	return &types.StateAccount{
		Nonce: 7, Balance: uint256.NewInt(100), CodeHash: []byte{0x01, 0x02},
	}, nil
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
	first.Balance.SetUint64(200)
	first.CodeHash[0] = 0xff

	second, err := cache.Account(addr)
	if err != nil {
		t.Fatal(err)
	}
	if second.Nonce != 7 || second.Balance.Uint64() != 100 || second.CodeHash[0] != 0x01 {
		t.Fatalf("shared cache leaked caller mutation: nonce=%d balance=%d code=%x",
			second.Nonce, second.Balance.Uint64(), second.CodeHash)
	}
	second.Nonce = 77
	third, err := cache.Account(addr)
	if err != nil {
		t.Fatal(err)
	}
	if third.Nonce != 7 {
		t.Fatalf("cached account changed on cache hit: nonce=%d", third.Nonce)
	}
}

// The shared account cache is used by speculative block prefetch and execution.
func TestResearchS01ConcurrentReaderAccountIsolation(t *testing.T) {
	cache := newStateReaderWithCache(researchS01Reader{})
	addr := common.Address{0x02}
	if _, err := cache.Account(addr); err != nil {
		t.Fatal(err)
	}
	const readers = 12
	const iterations = 200
	var wg sync.WaitGroup
	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				acct, err := cache.Account(addr)
				if err != nil {
					t.Error(err)
					return
				}
				if acct.Nonce != 7 || acct.Balance.Uint64() != 100 {
					t.Errorf("shared account corrupted: nonce=%d balance=%d", acct.Nonce, acct.Balance.Uint64())
					return
				}
				acct.Nonce = 99
				acct.Balance.SetUint64(200)
			}
		}()
	}
	wg.Wait()
}

func BenchmarkResearchS01AccountCacheHit(b *testing.B) {
	cache := newStateReaderWithCache(researchS01Reader{})
	addr := common.Address{0x03}
	if _, err := cache.Account(addr); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := cache.Account(addr); err != nil {
			b.Fatal(err)
		}
	}
}
