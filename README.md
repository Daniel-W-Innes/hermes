# Hermes

[![Go](https://github.com/Daniel-W-Innes/hermes/actions/workflows/tests.yml/badge.svg?branch=main)](https://github.com/Daniel-W-Innes/hermes/actions/workflows/tests.yml)

Hermes is a simple messaging service. It allows users do all four key operations on messages; add, remove, get, and
edit. Using these operations messages can be sent by one individual and seen by whomever that individual selects. Hermes
is built on the Fiber framework.

## Architecture

![Generic flow for all routes](./docs/genericFlow.png)

## Building and Deployment

Hermes is fully Dockerized. In order to run it one must generate or check out the appropriate secrets. These secrets are
stored in files in the secrets folder, in order to be loaded into Docker secrets. It should be noted that this in not
the only way of loading Docker secrets it is just the simplest way and therefore is the way used for this project. After
all of the secrets are in place the application can be started using the Docker up command. This will bring up and
configure PostgreSQL, along with Hermes. Hermes is configured to listen on port 8080 by default. There is a go module
file also included, in case anyone wants to run Hermes without Docker.

## Testing

For the sake of time, Hermes only has the bare minimum of tests. It has a best path test, for every single API endpoint
and specialized tests for the palindrome function. These tests can be run using go test and they are automatically run
using GitHub actions. Hermes has some example unit tests for add user in controllers. It should be noted that these tests
are not fully working sqlmock is going back automatically when it should not

## Environment variables

| Name | Type | Required | Description
|--|--|--|--|--|
| DB_HOST | String | Yes | Database server host
| DB_PORT | Integer | Yes | Port number for database
| MAX_OPEN_CONNS | Integer | Yes | Maximum number of open connections to the database
| MAX_IDLE_CONNS | Integer | Yes | Maximum number of connections in the idle connection pool
| DB_PASSWORD_FILE | File path | Yes | Password for the database
| DB_PASSWORD | String | Alternatives to DB_PASSWORD_FILE |  Password for the database
| DB_USER_FILE | File path | Yes | Username for database
| DB_USER | String | Alternatives to DB_USER_FILE | Username for database
| DB_NAME | String | Yes | Name of the database
| JWT_PRIVATE_KEY_FILE | File path | Yes | PEM encoded ecdsa private key
| JWT_PUBLIC_KEY_FILE | File path | Yes | PEM encoded ecdsa public key
| JWT_PRIVATE_KEY | String | Alternatives to JWT_PRIVATE_KEY_FILE | PEM encoded ecdsa private key
| JWT_PUBLIC_KEY | String | Alternatives to JWT_PUBLIC_KEY_FILE | PEM encoded ecdsa public key
| BCRYPT_COST | Integer | Yes | Number of key expansion rounds should be tuned to deployment hardware
| PEPPER_KEY_FILE | File path | Yes | Pre-hash secret to prevent off-line decoding
| PEPPER_KEY |  String | Alternatives to PEPPER_KEY_FILE | Pre-hash secret to prevent off-line decoding

## TODO

- [x] Adding users
- [x] Password hashing
- [x] Password checking
- [x] Generation
- [x] GWT verification
- [x] Adding messages
- [x] Deleting messages
- [x] Getting a specific message
- [x] Getting all messages, with user’s access rights
- [x] Input validation
- [x] Text interpretation (palindrome)
- [x] Best path testing
- [x] GitHub access
- [x] Docker deployment
- [ ] OpenAPI specification
- [X] Inline comments
- [ ] Diagrams
- [X] Robust error handling
- [ ] Logging
- [ ] Failure mode testing
- [ ] Unit testing
- [ ] Prometheus integration
