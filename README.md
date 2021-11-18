# Hermes

[![Go](https://github.com/Daniel-W-Innes/hermes/actions/workflows/tests.yml/badge.svg?branch=main)](https://github.com/Daniel-W-Innes/hermes/actions/workflows/tests.yml)

Hermes is a simple messaging service. It allows users do all four key operations on messages; add, remove, get, and
edit. Using these operations messages can be sent by one individual and seen by whomever that individual selects. Hermes
is built on the Fiber framework.

## Architecture

The code for Hermes is split into three distinct groups: routes, models, and utilities. The routes are the majority of
the business logic and have functions called handlers that are responsible for each type of API call. For example, there
is a handler called addMessage in the message file that add a message to the system. Models are the stateful objects of
the system and have methods in order to act as DOAs. This makes the routes simpler, as instead of embedding SQL in the
business logic there is a layer of abstraction with methods like insert. Additionally, some of the models have special
methods, for example setPassword that handles serialization or deserialization actions on that model. Utilities are
repeatedly used functions that are not necessarily business logic. These usually are database or API functions, e.g.,
authentications and database connections.

## Building and Deployment

Hermes is fully Dockerized. In order to run it one must generate or check out the appropriate secrets. These secrets are
stored in files in the secrets folder, in order to be loaded into Docker secrets. It should be noted that this in not
the only way of loading Docker secrets it is just the simplest way and therefore is the way used for this project. After
all of the secrets are in place thae application can be started using the Docker up command. This will bring up and
configure PostgreSQL, along with Hermes. Hermes is configured to listen on port 8080 by default. There is a go module
file also included, in case anyone wants to run Hermes without Docker.

## Testing

For the sake of time, Hermes only has the bare minimum of tests. It has a best path test, for every single API endpoint
and specialized tests for the palindrome function. These tests can be run using go test and they are automatically run
using GitHub actions.

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
- [ ] Inline comments
- [ ] Diagrams
- [ ] Robust error handling
- [ ] Logging
- [ ] Failure mode testing
- [ ] Prometheus integration
