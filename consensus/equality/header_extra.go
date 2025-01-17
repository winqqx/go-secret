package equality

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"strings"

	"github.com/SecretBlockChain/go-secret/common"
	"github.com/SecretBlockChain/go-secret/core/types"
	"github.com/SecretBlockChain/go-secret/params"
	"github.com/SecretBlockChain/go-secret/rlp"
)

// Root is the state tree root.
type Root struct {
	EpochHash     common.Hash
	CandidateHash common.Hash
	MintCntHash   common.Hash
	ConfigHash    common.Hash
}

func (root Root) PrintDifference(number uint64, other Root) {
	slice := make([]string, 0)
	slice = append(slice, fmt.Sprintf("BlockNumber: %d", number))
	if root.EpochHash != other.EpochHash {
		slice = append(slice, fmt.Sprintf("EpochHash: %s ---- %s", root.EpochHash.String(), other.EpochHash.String()))
	}
	if root.CandidateHash != other.CandidateHash {
		slice = append(slice, fmt.Sprintf("CandidateHash: %s ---- %s", root.CandidateHash.String(), other.CandidateHash.String()))
	}
	if root.MintCntHash != other.MintCntHash {
		slice = append(slice, fmt.Sprintf("MintCntHash: %s ---- %s", root.MintCntHash.String(), other.MintCntHash.String()))
	}
	if root.ConfigHash != other.ConfigHash {
		slice = append(slice, fmt.Sprintf("ConfigHash: %s ---- %s", root.ConfigHash.String(), other.ConfigHash.String()))
	}
	fmt.Printf("######### Root Hash Difference #########\n%s\n", strings.Join(slice, "\n"))
}

// HeaderExtra is the struct of info in header.Extra[extraVanity:len(header.extra)-extraSeal].
// HeaderExtra is the current struct.
type HeaderExtra struct {
	Root                          Root
	Epoch                         uint64
	EpochBlock                    uint64
	CurrentBlockCandidates        []common.Address
	CurrentBlockKickOutCandidates []common.Address
	CurrentBlockCancelCandidates  []common.Address
	CurrentEpochValidators        []common.Address
	ChainConfig                   []params.EqualityConfig
}

// NewHeaderExtra new HeaderExtra from rlp bytes.
func NewHeaderExtra(data []byte) (HeaderExtra, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return HeaderExtra{}, err
	}

	buffer := bytes.NewBuffer(nil)
	for {
		var temp [128]byte
		n, err := r.Read(temp[:])
		if n > 0 {
			buffer.Write(temp[:n])
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return HeaderExtra{}, err
		}
	}

	var headerExtra HeaderExtra
	if err := rlp.DecodeBytes(buffer.Bytes(), &headerExtra); err != nil {
		return HeaderExtra{}, err
	}
	return headerExtra, nil
}

// Encode encode header extra as rlp bytes.
func (headerExtra HeaderExtra) Encode() ([]byte, error) {
	data, err := rlp.EncodeToBytes(headerExtra)
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewBuffer(nil)
	w := gzip.NewWriter(buffer)
	w.Write(data)
	w.Close()
	return buffer.Bytes(), nil
}

// Equal compares two HeaderExtras for equality.
func (headerExtra HeaderExtra) Equal(other HeaderExtra) bool {
	if headerExtra.Root != other.Root {
		return false
	}
	if headerExtra.Epoch != other.Epoch {
		return false
	}
	if headerExtra.EpochBlock != other.EpochBlock {
		return false
	}

	if len(headerExtra.ChainConfig) != len(other.ChainConfig) {
		return false
	}
	for idx, config := range headerExtra.ChainConfig {
		if !config.Equal(other.ChainConfig[idx]) {
			return false
		}
	}

	if len(headerExtra.CurrentBlockCandidates) != len(other.CurrentBlockCandidates) {
		return false
	}
	for idx, candidate := range headerExtra.CurrentBlockCandidates {
		if candidate != other.CurrentBlockCandidates[idx] {
			return false
		}
	}

	if len(headerExtra.CurrentBlockKickOutCandidates) != len(other.CurrentBlockKickOutCandidates) {
		return false
	}
	for idx, candidate := range headerExtra.CurrentBlockKickOutCandidates {
		if candidate != other.CurrentBlockKickOutCandidates[idx] {
			return false
		}
	}

	if len(headerExtra.CurrentBlockCancelCandidates) != len(other.CurrentBlockCancelCandidates) {
		return false
	}
	for idx, candidate := range headerExtra.CurrentBlockCancelCandidates {
		if candidate != other.CurrentBlockCancelCandidates[idx] {
			return false
		}
	}

	if len(headerExtra.CurrentEpochValidators) != len(other.CurrentEpochValidators) {
		return false
	}
	for idx, validator := range headerExtra.CurrentEpochValidators {
		if validator != other.CurrentEpochValidators[idx] {
			return false
		}
	}
	return true
}

func DecodeHeaderExtra(header *types.Header) (HeaderExtra, error) {
	headerExtra := header.Extra
	if len(headerExtra) < extraVanity {
		return HeaderExtra{}, errMissingVanity
	}
	if len(headerExtra) < extraVanity+extraSeal {
		return HeaderExtra{}, errMissingSignature
	}
	return NewHeaderExtra(headerExtra[extraVanity : len(headerExtra)-extraSeal])
}

// Returns whether an address exists in the address list.
func addressesExist(slice []common.Address, addr common.Address) bool {
	for _, address := range slice {
		if address == addr {
			return true
		}
	}
	return false
}

// Ensure each element of an common.Address slice are not the same.
func addressesDistinct(slice []common.Address) []common.Address {
	if len(slice) <= 1 {
		return slice
	}

	set := make(map[common.Address]struct{})
	result := make([]common.Address, 0, len(slice))
	for _, address := range slice {
		if _, ok := set[address]; !ok {
			set[address] = struct{}{}
			result = append(result, address)
		}
	}
	return result
}

// Remove an element from the address list.
func addressesRemove(slice []common.Address, addr common.Address) []common.Address {
	result := make([]common.Address, 0, len(slice))
	for _, address := range slice {
		if address != addr {
			result = append(result, address)
		}
	}
	return result
}
