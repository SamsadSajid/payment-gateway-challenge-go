# payment-gateway-challenge-go

A Golang based payment gateway that merchant can use to capture their payment requests. The gateway validates requests and send it to the acquired bank and return the response to the merchant.

# Test

- Run unit tests: `make test`
- Run integration tests: `make test-integration`.

The integration test requires the bank docker container in running state. However, you do not need to run any command to launch the container. The `make` command will launch the container. If the container is already running, the above command will just run the integration tests.

All tests are run in parallel mode to reduce test running time.

Documentation can be found in [architecture decisions](docs/architecture/decisions/0002-implement-payment-gateway-api-interface.md) file

# Instructions for candidates

This is the Go version of the Payment Gateway challenge. If you haven't already read the [README.md](https://github.com/cko-recruitment/.github/tree/beta) in the root of this organisation, please do so now. 

## Template structure
```
main.go - a skeleton Payment Gateway API
imposters/ - contains the bank simulator configuration. Don't change this
docs/docs.go - Generated file by Swaggo
.editorconfig - don't change this. It ensures a consistent set of rules for submissions when reformatting code
docker-compose.yml - configures the bank simulator
.goreleaser.yml - Goreleaser configuration
```

Feel free to change the structure of the solution, use a different test library etc.

### Swagger
This template uses Swaggo to autodocument the API and create a Swagger spec. The Swagger UI is available at http://localhost:8090/swagger/index.html.
