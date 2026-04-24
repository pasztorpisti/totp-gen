totp-gen
========

A command-line tool that generates time-based one-time passwords (TOTP) just
like Google Authenticator. It supports the same input formats: the standard
[TOTP URI](https://github.com/google/google-authenticator/wiki/Key-Uri-Format)
(usually [scanned from a QR code](#working-with-qr-codes)) or a base32-encoded
secret key.

This stateless minimalist tool only generates codes from provided secrets
and does not store or manage your secrets like mobile authenticator apps.

The implementation is [less than 200 lines of code in one file](./main.go), and
depends only on the builtin standard library. Written in golang, so a single `go`
command can (cross-)compile it into a self-contained dependency-free native binary.
See the ["Installation"](#installation) section below.


Usage
-----

The `totp-gen` cli reads a TOTP URI (`otpauth://totp/...`) or a base32-encoded
TOTP secret key from stdin and writes the current one-time password to stdout.
There are no commandline parameters.

```bash
# Safe examples: the secret is piped into stdin from another program,
# so it isn't exposed on the screen or recorded in shell history files:

$ gpg -qd secret.gpg | totp-gen
479991
$ pass show test/secret | totp-gen
479991

# Unsafe examples for testing: the commands include the secret key, so
# it is exposed on the screen and recorded in the shell history file:

# Base32-encoded secret key as input:
$ echo "M3B3Y43UYO3GY5BAONXW423BEAVSAZWFSF2HIIDUN5VMHILT" | totp-gen
479991
# Piping the secret into stdin with the "here string" operator (<<<).
# Supported by bash/zsh/ksh/... but not all shells.
$ totp-gen <<< "M3B3 Y43U YO3G Y5BA ONXW 423B EAVS AZWF SF2H IIDU N5VM HILT"
479991

# TOTP URI as input:
$ totp-gen <<< "otpauth://totp/LABEL?secret=M3B3Y43UYO3GY5BAONXW423BEAVSAZWFSF2HIIDUN5VMHILT&digits=6&period=30&algorithm=SHA1&issuer=me"
479991
```


Installation
------------

**Prerequisite**: Install the [go (golang) compiler](https://go.dev/doc/install).
Don't worry, you don't have to be familiar with the go programming language.
The compiler is straightforward compared to others, and in most cases it
produces self-contained, dependency-free native binaries.

Compile the sources directly from github, and output the `totp-gen` binary into
the current directory:

```bash
GOBIN="$PWD" go install github.com/pasztorpisti/totp-gen@latest
./totp-gen help
```

Alternatively, if you prefer compiling from local sources:

```bash
git clone https://github.com/pasztorpisti/totp-gen
cd totp-gen
go build
./totp-gen help
```


Working with QR Codes
---------------------

There are a lot of QR code encoder/decoder libraries and tools, including online
ones. Make sure to use a local-only solution when working with QR codes
containing sensitive info like your TOTP secrets.

`qrencode` and `zbarimg` are popular cli tools used to manipulate QR codes:

```bash
# Debian-based linux distros (Ubuntu, Mint, ...):
sudo apt install qrencode
sudo apt install zbar-tools

# MacOS
brew install qrencode
brew install zbar
```

**Example**: Encoding the previously used test TOTP URI into a QR code image
(in PNG file format), and then decoding it back to the URI:

```bash
# Encode: URI -> qr.png
$ qrencode -t png -o qr.png "otpauth://totp/LABEL?secret=M3B3Y43UYO3GY5BAONXW423BEAVSAZWFSF2HIIDUN5VMHILT&digits=6&period=30&algorithm=SHA1&issuer=me"

# Decode: qr.png -> URI
$ zbarimg -q --raw qr.png
otpauth://totp/LABEL?secret=M3B3Y43UYO3GY5BAONXW423BEAVSAZWFSF2HIIDUN5VMHILT&digits=6&period=30&algorithm=SHA1&issuer=me
```

`zbarimg` and similar cli tools struggle to decode rotated, blurry, or distorted
QR codes because they haven't caught up with the ML-based approaches used by
mobile devices. However, during a typical 2FA onboarding process, this isn't an
issue as a pristine QR code image is usually available in the browser for direct
download.

**Bonus**: For a quick test with standard authenticator mobile apps,
let's encode our test TOTP URI into a QR code and print it to the terminal
as ASCII art with the `-t ascii` or `-t utf8` parameter:

```bash
$ qrencode -t utf8 "otpauth://totp/LABEL?secret=M3B3Y43UYO3GY5BAONXW423BEAVSAZWFSF2HIIDUN5VMHILT&digits=6&period=30&algorithm=SHA1&issuer=me"
█████████████████████████████████████████████████
█████████████████████████████████████████████████
████ ▄▄▄▄▄ █  ▀▄ █▄█▀ █ ▄  ▄█ ▄▄▀▀▄ ▀█ ▄▄▄▄▄ ████
████ █   █ █  ▄█▄▄▀ ▄▀▀ ▀█ █ ▀█▄▀▄█▄▀█ █   █ ████
████ █▄▄▄█ █▀█ ▄▀█ █▀ ▄ ▄  █▄   ▀███ █ █▄▄▄█ ████
████▄▄▄▄▄▄▄█▄█▄█▄█ ▀ ▀▄█▄█▄▀▄▀ ▀ ▀ █▄█▄▄▄▄▄▄▄████
████ ▄▀▀ ▄▄▀█▄▄▀▄ ▄█▀▄█  █▄     ▄█▀ ▄▀▀ ▀    ████
████ ▀▀ ▀▀▄█▄▄▄▄▀█▀█▀ ▀███▀ ▄▀█ ▄█▀█▄▀ █ ▀█ ▄████
█████████▄▄▀█▄▀▄▄ ▄█▀▀▄▀▄ ▀▀▀ ▄▀█ ▄▄▀▀ ▄▀ █▀ ████
████▄▄▄▀  ▄  █ █▀ ██▄ ▀ ▀████ ▀█▄ ▄▀▄▀ ▀▀▄▀▀ ████
████▀ ▀▄█ ▄█▄▄  ▀▄▀ ▄ ▀█ █▄▀ █▀  ▄▄ ▄▄▀█▀▀  ▀████
█████▄▀ ▀ ▄▄▄ ▄▄█ ▄ ▀█▄█▄█  ▄▀█ ██ █▀█▄█  █ ▀████
████▀▄▀▀▀█▄▀▀  █▀ ▀█▄▀▀▀▀▀█▀▀ ▄█▄ ▄▀▀▀ ▄███ ▀████
████▄▀▄▀█▀▄▀▀ ▄ ▄ ▀ ▀  █▄ ▀███▀█▄▀ ████ ▄█ █▄████
██████▀█▀ ▄ ▀▀▄▀▀ █▀ ▄█ ▄█▄ ▄▀▄▀▀ █▀▄ ▀█▀█   ████
████▄ ▀▀▄▄▄▄▄ █▄▄█▀▀▄ ▀██  ██▄▀▄  ▀▀  ▄   █▄█████
████ ▀▀█▀▀▄█ ▄▀█▄▄▄▄▀▀▄▀▄▀█▀▀█▄▀▄▄▄▀▀▀█▄▀▀██▀████
███████▄█▄▄█ ▀▀█▀▄▄█▄▄▄ ▀██▀▄▄▄ ▄▀▀▄▄▀  ▀▄▄▀ ████
████▄▄█▄▄█▄▄▀  ▄▀██ ▀ █▀█ ▀▄█ █▀ ▄▀▀ ▄▄▄ █   ████
████ ▄▄▄▄▄ █▀▀ ▄▀ ▄ ██▄ ██▀ █▀▄▄▀▄▄▄ █▄█ ▀█ ▀████
████ █   █ █▄▀ █▀ ▀▄▄▀▀ ▀▀▄ ▀▀▄▀▄▄▄   ▄ ▄ ▀ █████
████ █▄▄▄█ █▀▀▀ ▄▄█ ▀▀▀█▄▄█▀▄  ▄   ▀███ █▄▄  ████
████▄▄▄▄▄▄▄█▄▄███▄▄███▄▄▄▄█▄███▄█▄█▄██▄████▄█████
█████████████████████████████████████████████████
█████████████████████████████████████████████████
```
