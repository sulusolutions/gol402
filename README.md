# Golang L402 SDK

The Golang L402 SDK is a comprehensive Go package designed to simplify the integration and handling of L402 protocol payments within the Lightning Network ecosystem. This SDK offers convenient abstractions for wallet interactions, invoice payments, and token management, making it an essential tool for developers working on Go-based applications requiring L402 compliance.

## Features

- **Wallet Interface**: Facilitates invoice payments through various wallet implementations, starting with Alby wallet support.
- **Token Store Interface**: Manages and stores L402 tokens, allowing for efficient retrieval based on URL, host, and path with support for closest match searching.
- **L402 Protocol Compliance**: Ensures that interactions adhere to the L402 payment protocol standards.

## Getting Started

### Prerequisites

- Go version 1.16 or higher
- Access to an L402 compliant payment gateway

### Installation

To start using the Golang L402 SDK, install it using `go get`:

```sh
go get github.com/sulusolutions/gol402
