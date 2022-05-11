// Copyright 2021 the System Transparency Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package trust

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"fmt"

	"github.com/system-transparency/stboot/stlog"
)

const (
	ErrRSAPSSSigner  = Error("RSAPSSSigner error")
	ErrED25519Signer = Error("ED25519Signer error")
)

// Signer is used by OSPackage to sign and varify the OSPackage.
type Signer interface {
	Sign(key crypto.PrivateKey, data []byte) ([]byte, error)
	Verify(sig, hash []byte, key crypto.PublicKey) error
}

// DummySigner implements the Signer interface. It creates signatures
// that are always valid.
type DummySigner struct{}

var _ Signer = DummySigner{}

// Sign returns a signature containing just 8 random bytes.
func (DummySigner) Sign(key crypto.PrivateKey, data []byte) ([]byte, error) {
	const n = 8
	sig := make([]byte, n)

	if _, err := rand.Read(sig); err != nil {
		return nil, fmt.Errorf("sign: %w", err)
	}

	return sig, nil
}

// Verify will never return an error.
func (DummySigner) Verify(sig, hash []byte, key crypto.PublicKey) error {
	return nil
}

// RSAPSSSigner implements the Signer interface. It uses SHA256 hashes
// and PSS signatures along with x509 certificates.
type RSAPSSSigner struct{}

var _ Signer = RSAPSSSigner{}

// Sign signes the provided data with the key named by privKey. The returned
// byte slice contains a PSS signature value.
func (RSAPSSSigner) Sign(key crypto.PrivateKey, data []byte) ([]byte, error) {
	if len(data) == 0 {
		stlog.Debug("RSAPSSSigner: input data has zero length")

		return nil, ErrRSAPSSSigner
	}

	priv, ok := key.(*rsa.PrivateKey)
	if !ok {
		stlog.Debug("RSAPSSSigner: invalid key type %T", key)

		return nil, ErrRSAPSSSigner
	}

	opts := &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash}

	ret, err := rsa.SignPSS(rand.Reader, priv, crypto.SHA256, data, opts)
	if err != nil {
		return nil, fmt.Errorf("sign: %w", err)
	}

	return ret, nil
}

// Verify checks if sig contains a valid signature of hash.
func (RSAPSSSigner) Verify(sig, hash []byte, key crypto.PublicKey) error {
	if len(sig) == 0 {
		stlog.Debug("RSAPSSSigner: signature has zero length")

		return ErrRSAPSSSigner
	}

	if len(hash) == 0 {
		stlog.Debug("RSAPSSSigner: hash has zero length")

		return ErrRSAPSSSigner
	}

	pub, ok := key.(*rsa.PublicKey)
	if !ok {
		stlog.Debug("RSAPSSSigner: invalid key type %T", key)

		return ErrRSAPSSSigner
	}

	opts := &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash}

	err := rsa.VerifyPSS(pub, crypto.SHA256, hash, sig, opts)
	if err != nil {
		stlog.Debug("RSAPSSSigner: verification failed")

		return ErrRSAPSSSigner
	}

	return nil
}

type ED25519Signer struct{}

var _ Signer = ED25519Signer{}

// Sign signes the provided data with the key named by privKey.
func (ED25519Signer) Sign(key crypto.PrivateKey, data []byte) ([]byte, error) {
	if len(data) == 0 {
		stlog.Debug("ED25519Signer: input data has zero length")

		return nil, ErrED25519Signer
	}

	priv, ok := key.(ed25519.PrivateKey)
	if !ok {
		stlog.Debug("ED25519Signer: invalid key type %T", key)

		return nil, ErrED25519Signer
	}

	return ed25519.Sign(priv, data), nil
}

// Verify checks if sig contains a valid signature of hash.
func (ED25519Signer) Verify(sig, hash []byte, key crypto.PublicKey) error {
	if len(sig) == 0 {
		stlog.Debug("ED25519Signer: signature has zero length")

		return ErrED25519Signer
	}

	if len(hash) == 0 {
		stlog.Debug("ED25519Signer: hash has zero length")

		return ErrED25519Signer
	}

	pub, ok := key.(ed25519.PublicKey)
	if !ok {
		stlog.Debug("ED25519Signer: invalid key type %T", key)

		return ErrED25519Signer
	}

	isValid := ed25519.Verify(pub, hash, sig)
	if !isValid {
		stlog.Debug("ED25519Signer: verification failed")

		return ErrED25519Signer
	}

	return nil
}
