# Golang L402 Library

gol402 is a comprehensive Go package designed to simplify the integration and handling of L402 protocol payments within the Lightning Network ecosystem. This SDK offers convenient abstractions for wallet interactions, invoice payments, and token management, making it an essential tool for developers working on Go-based applications requiring L402 API access.

## Features

- **L402 Client**: Composable L402 HTTP client to handle L402 API requests.
- **Wallet Interface**: Facilitates invoice payments through various wallet implementations, starting with Alby wallet support.
- **Token Store Interface**: Manages and stores L402 tokens, allowing for efficient retrieval based on URL, host, and path with support for closest match searching.

## Getting Started

### Prerequisites

- Go version 1.21 or higher
- Access to an L402 compliant payment gateway

### Installation

To start using the Golang L402 SDK, install it using `go get`:

```sh
go get github.com/sulusolutions/gol402
