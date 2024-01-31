![tests](https://github.com/sulusolutions/gol402/actions/workflows/tests.yaml/badge.svg)

# gol402: Golang L402 Library

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
```

## Example Usage

This example demonstrates how to use the L402 client with the Alby wallet to make a request to the `rnd.ln.sulu.sh/randomnumber` API, which returns a random number.

### Quick Start

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "os"

    "github.com/sulusolutions/gol402/client"
    "github.com/sulusolutions/gol402/tokenstore"
    "github.com/sulusolutions/gol402/wallet/alby"
)

func main() {
    // Initialize the Alby wallet with your bearer token.
    albyWallet := alby.NewAlbyWallet(os.Getenv("ALBY_BEARER_TOKEN"))

    // Use an in-memory token store.
    tokenStore := tokenstore.NewInMemoryStore()

    // Create the L402 client.
    l402Client := client.New(albyWallet, tokenStore)

    // Create a new HTTP request to the rnd.ln.sulu.sh/randomnumber API.
    req, err := http.NewRequest("GET", "https://rnd.ln.sulu.sh/randomnumber", nil)
    if err != nil {
        fmt.Printf("Failed to create request: %v\n", err)
        return
    }

    // Use the modified MakeRequest function which takes *http.Request.
    response, err := l402Client.MakeRequest(context.Background(), req)
    if err != nil {
        fmt.Printf("Error making request: %v\n", err)
        return
    }
    defer response.Body.Close()

    fmt.Println("Request successful, status code:", response.StatusCode)
}
```

### Notes

- Ensure the __ALBY_BEARER_TOKEN__ environment variable is set with your Alby wallet bearer token before running the example.
- The client automatically handles the L402 payment if required by the API.