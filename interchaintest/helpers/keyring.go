package helpers

// credit: https://github.com/persistenceOne/persistenceCore/blob/main/interchaintest/helpers/keyring.go

import (
	"crypto/rand"

	"github.com/cosmos/cosmos-sdk/codec"
	cosmcrypto "github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bip39 "github.com/cosmos/go-bip39"
	"github.com/pkg/errors"
)

func NewKeyringFromMnemonic(cdc codec.Codec, keyName, mnemonic string) (keyring.Keyring, error) {
	if !bip39.IsMnemonicValid(mnemonic) {
		err := errors.New("provided memnemonic is not a valid BIP39 mnemonic")
		return nil, err
	}

	cfg := sdk.GetConfig()
	pkBytes, err := hd.Secp256k1.Derive()(
		mnemonic,
		keyring.DefaultBIP39Passphrase,
		cfg.GetFullBIP44Path(),
	)
	if err != nil {
		err = errors.Wrap(err, "failed to derive secp256k1 private key")
		return nil, err
	}

	cosmosAccPk := hd.Secp256k1.Generate()(pkBytes)

	return newKeyringFromPrivKey(cdc, keyName, cosmosAccPk)
}

func NewMnemonic() string {
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		panic(err)
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		panic(err)
	}

	return mnemonic
}

// newKeyringFromPrivKey creates a temporary in-mem keyring for a PrivKey.
// Allows to init Context when the key has been provided in plaintext and parsed.
func newKeyringFromPrivKey(cdc codec.Codec, name string, privKey cryptotypes.PrivKey) (keyring.Keyring, error) {
	kb := keyring.NewInMemory(cdc)
	tmpPhrase := randPhrase(64)
	armored := cosmcrypto.EncryptArmorPrivKey(privKey, tmpPhrase, privKey.Type())
	err := kb.ImportPrivKey(name, armored, tmpPhrase)
	if err != nil {
		err = errors.Wrap(err, "failed to import privkey")
		return nil, err
	}

	return kb, nil
}

func randPhrase(size int) string {
	buf := make([]byte, size)
	_, err := rand.Read(buf)
	if err != nil {
		panic(err)
	}

	return string(buf)
}
