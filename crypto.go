package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/sirupsen/logrus"
)

func checkSignature(publickey, signature string, contents []byte) (bool, error) {
	pk, err := hex.DecodeString(publickey)
	if err != nil {
		return false, fmt.Errorf("pubkey is invalid hex: %w", err)
	}

	pubkey, err := schnorr.ParsePubKey(pk)
	if err != nil {
		return false, fmt.Errorf("event has invalid pubkey '%s': %w", pubkey, err)
	}

	s, err := hex.DecodeString(signature)
	if err != nil {
		return false, fmt.Errorf("signature is invalid hex: %w", err)
	}

	sig, err := schnorr.ParseSignature(s)
	if err != nil {
		return false, fmt.Errorf("failed to parse signature: %w", err)
	}

	shasum := sha256.Sum256(contents)

	logrus.WithFields(logrus.Fields{
		"fileShasum": hex.EncodeToString(shasum[:]),
		"pubkey":     publickey,
		"signature":  signature,
	}).Debug("checking signature")

	return sig.Verify(shasum[:], pubkey), nil
}
