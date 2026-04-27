// SPDX-License-Identifier: MIT-0
// SPDX-FileCopyrightText:  2026 Istvan Pasztor

package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"io"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const usage = `
Usage: totp-gen <<< totp_secret_or_uri

Reads a TOTP secret key or an otpauth://totp/ URI from
stdin and writes the current one-time password to stdout.

Example usage:
  # Reading the secret key or URI from an encrypted file:
  gpg -qd secret.gpg | totp-gen

  # Reading the secret key or URI with a password manager cli:
  pass show web/github | totp-gen

  # Unsafe examples for testing (may record secrets in your shell history file):
  echo "M3B3Y43UYO3GY5BAONXW423BEAVSAZWFSF2HIIDUN5VMHILT" | totp-gen
  totp-gen <<< "M3B3 Y43U YO3G Y5BA ONXW 423B EAVS AZWF SF2H IIDU N5VM HILT"
  totp-gen <<< "otpauth://totp/LABEL?secret=M3B3Y43UYO3GY5BAONXW423BEAVSAZWFSF2HIIDUN5VMHILT&digits=6&period=30&algorithm=SHA1&issuer=me"

Defaults (de-facto standard):
  digits: 6
  period: 30 (seconds)
  algorithm: SHA1
`

func main() {
	if len(os.Args) != 1 {
		_, _ = fmt.Fprint(os.Stderr, strings.Replace(usage, "totp-gen", filepath.Base(os.Args[0]), -1))
		os.Exit(2)
	}

	err := run(os.Stdin, os.Stdout, time.Now())
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

const maxInputSize = 1024 * 16 // this is plenty for a secret or URI

func run(input io.Reader, output io.Writer, at time.Time) error {
	secretOrURI, err := io.ReadAll(io.LimitReader(input, maxInputSize+1))
	if err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}
	if len(secretOrURI) > maxInputSize {
		return errors.New("input too large")
	}

	totp, err := parseSecretOrURI(strings.TrimSpace(string(secretOrURI)))
	if err != nil {
		return fmt.Errorf("error parsing input: %w", err)
	}
	_, err = fmt.Fprintln(output, totp.Generate(at))
	if err != nil {
		return fmt.Errorf("error writing output: %w", err)
	}
	return nil
}

type TOTP struct {
	HOTP
	Period int
}

func (t *TOTP) Generate(at time.Time) string {
	counter := uint64(at.Unix()) / uint64(t.Period)
	return t.HOTP.Generate(counter)
}

type HOTP struct {
	Secret    []byte
	Digits    int
	Algorithm func() hash.Hash
}

func (h *HOTP) Generate(counter uint64) string {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, counter)

	mac := hmac.New(h.Algorithm, h.Secret)
	mac.Write(buf)
	sum := mac.Sum(nil)

	offset := sum[len(sum)-1] & 0xf
	code := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff
	return fmt.Sprintf("%0*d", h.Digits, code%uint32(math.Pow10(h.Digits)))
}

func parseSecretOrURI(secretOrURI string) (*TOTP, error) {
	if strings.HasPrefix(secretOrURI, "otpauth://totp/") {
		return parseURI(secretOrURI)
	}
	return parseSecret(secretOrURI)
}

func parseSecret(secretBase32 string) (*TOTP, error) {
	// Remove whitespace characters from the secret.
	// Many services provide the secret with spaces (often in blocks of
	// four or five characters) to make it easier to read and transcribe.
	secretBase32 = strings.Join(strings.Fields(secretBase32), "")

	if secretBase32 == "" {
		return nil, errors.New("empty secret")
	}

	// The standard (uppercase base32 without padding) isn't always respected.
	enc := base32.StdEncoding
	if len(secretBase32)%8 != 0 {
		enc = enc.WithPadding(base32.NoPadding)
	}
	secret, err := enc.DecodeString(strings.ToUpper(secretBase32))
	if err != nil {
		return nil, fmt.Errorf("failed to base32 decode secret: %w", err)
	}

	return &TOTP{
		HOTP: HOTP{
			Secret:    secret,
			Digits:    6,        // this default is the de-facto standard
			Algorithm: sha1.New, // this default is the de-facto standard
		},
		Period: 30, // this default is the de-facto standard
	}, nil
}

func parseURI(totpURI string) (*TOTP, error) {
	u, err := url.Parse(totpURI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse TOTP URI: %w", err)
	}
	q := u.Query()

	totp, err := parseSecret(q.Get("secret"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse TOTP URI: %w", err)
	}

	if q.Has("digits") {
		v, err := strconv.ParseUint(q.Get("digits"), 10, 31)
		if err != nil || v < 1 || v > 9 {
			// The 0x7fffffff mask of the algorithm limits the max number of useful digits to 9.
			// Google Authenticator allows only 6, 7, or 8 digits. We are less strict.
			return nil, fmt.Errorf("TOTP: invalid 'digits' parameter: %q (requirement: 1 <= digits <= 9)", q.Get("digits"))
		}
		totp.Digits = int(v)
	}

	if q.Has("period") {
		v, err := strconv.ParseUint(q.Get("period"), 10, 31)
		if err != nil {
			return nil, fmt.Errorf("TOTP: invalid 'period' parameter: %q", q.Get("period"))
		}
		totp.Period = int(v)
	}

	if q.Has("algorithm") {
		switch strings.ToUpper(q.Get("algorithm")) {
		case "SHA512":
			totp.Algorithm = sha512.New
		case "SHA256":
			totp.Algorithm = sha256.New
		case "SHA1": // de-facto standard, already set in the TOTP struct
		default:
			return nil, fmt.Errorf("TOTP: unsupported 'algorithm' parameter: %q", q.Get("algorithm"))
		}
	}

	return totp, nil
}
