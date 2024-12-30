# Kontroler Go Client Library

This repository contains the Go client library for interacting with Kontroler, a REST API for managing and automating your workflows.

## Installation

To install the library, run:

```sh
go get github.com/greedykomododragon/kontroler-client
```

## Getting Started

### Prerequisites

* Go 1.23.3 or newer
* Access to a Kontroler server with the API enabled

### Client Creation Code

```go
package main

import (
    "fmt"
    "log"
    "github.com/greedykomododragon/kontroler-client"
)

func main() {
    config := &kontroler.ClientConfig{
		Url:            "https://example-url-kontroler.com",
		Username:       "admin",
		Password:       os.Getenv("PASSWORD"),
		AuthCookieName: os.Getenv("AUTH-COOKIE"),
	}

    client, _ := kontroler.NewClient(config)
}
```


## Contributing

We welcome contributions! Please submit pull requests or open issues to help improve the library.


## License

This library is licensed under the Apache 2.0 License. See the LICENSE file for more information.


## Why a Separate Repository?

The main Kontroler project is licensed under the GNU General Public License (GPL), which enforces strict copyleft requirements. To support commercial applications that may not be compatible with GPL licensing, this client library is maintained under the more permissive Apache 2.0 License.

By using this library, you can build commercial and proprietary software that integrates with Kontroler without needing to release your source code under the GPL.

If you make changes to the Core of Kontroler you will need to abide by the GPL license.
