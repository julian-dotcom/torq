# Contributing to torq

Thank you for considering to contribute to torq! The guidelines bellow are not set in stone. As the project grows so will this document. Feel free to propose changes to this document by opening a pull request.

## Setting up the developer environment

### Requirements

You'll need the following tools to run torq:

- git
- docker
- go
- node + npm
- make

#### Windows extras

- git bash
  - Installed with the unix tools is a usefull tool to run our end-to-end tests on Windows.
- Docker Desktop
  - Is also compatible with torq. Please also activate WSL2 in Docker desktop.

### Simnet Lightning Network and development database

You can create a virtual **btcd**, **lnd** and **database** as development environment\*\* for Alice, Carol and Bob with
one command, either with a simple Make command or Go.

- **Alternative 1 - Make Command:** `make create-dev-env`
- **Alternative 2 - Go command:** `go build ./virtual_network/torq_vn && go run ./virtual_network/torq_vn create --db true`

The tls.cert and admin.macaroon files will be exported for alice, bob and carol. These are available under `/virtual_network/generated_files`.

> ℹ️  If you get an error regarding timescaledb please run the command: `docker pull timescale/timescaledb:latest-pg14`


### Running the frontend

Go on the folder web `cd /web` and install the dependencies with the command `npm install --legacy-peer-deps`.
To create the required build folder into /web folder run the command: `npm run build`

You can now start the frontend with the command `npm start`. This will start the frontend on port 3000.
You will need to start the backend to log in. The password will be `password`, unless you changed it when you started the backend.

### Running the backend

In the projects root folder run:

```bash
go run ./cmd/torq/torq.go --torq.password password --db.name torq --db.port 5444 --db.password password start
```

### Adding nodes to Torq

When torq has started, go to the frontend, navigate to the Settings page, and then add bob as new node.

You can add alice or carol as well, but this is only recommended if you wish to develop features that are
impacted by having multiple nodes added.

**Bob connection details:**
- GRPC Address: `localhost:10009`
- TLS Certificate: `dev-bob-tls.cert`
- Macaroon: `dev-bob-admin.macaroon`

**Alice connection details:**
- GRPC Address: `localhost:10008`
- TLS Certificate: `dev-alice-tls.cert`
- Macaroon: `dev-alice-admin.macaroon`
-
**Carol connection details:**
- GRPC Address: `localhost:10010`
- TLS Certificate: `dev-carol-tls.cert`
- Macaroon: `dev-carol-admin.macaroon`

### Optional: Run torq compartments in isolation

This is not part of the normal development workflow. It's only if you wish to run parts of the environment in isolation.

To run the database `docker run -d --name torqdb -p ${dbPort}:5432 -e POSTGRES_PASSWORD=${dbPassword} timescale/timescaledb:latest-pg14`

To run the backend without LND/CLN event subscription on port 8080 `go run ./cmd/torq/torq.go --db.name ${dbName} --db.password ${dbPassword} --db.port ${dbPort} --torq.password ${torqPassword} --torq.no-sub start`

To run the frontend in dev mode on port 3000 you can use `cd web && npm start`


## Useful development commands

### Start, Stop and Purge the development environment

Once the virtual environment is created, it will already be running and ready to go. However, if you can start, stop and purge (delete) the
environment with the following commands:

Stop:
- `make stop-dev-env`
- `go build ./virtual_network/torq_vn && go run ./virtual_network/torq_vn stop --db true`

Start:
  - `make start-dev-env`
  - `go build ./virtual_network/torq_vn && go run ./virtual_network/torq_vn start --db true`

Purge (delete):
- `make purge-dev-env`
- `go build ./virtual_network/torq_vn && go run ./virtual_network/torq_vn purge --db true`

### Interacting with the virtual network nodes

You can interact with the nodes, to simplify this you can add functions to access the lncli and btcd by simply running the following commands
```bash
alice() { docker exec -it  dev-alice /bin/bash -c "lncli --macaroonpath="/root/.lnd/data/chain/bitcoin/simnet/admin.macaroon" --network=simnet $@"};
bob() { docker exec -it  dev-bob /bin/bash -c "lncli --macaroonpath="/root/.lnd/data/chain/bitcoin/simnet/admin.macaroon" --network=simnet $@"};
carol() { docker exec -it  dev-carol /bin/bash -c "lncli --macaroonpath="/root/.lnd/data/chain/bitcoin/simnet/admin.macaroon" --network=simnet $@"};
vbtcd() { docker exec -it  dev-btcd /bin/bash -c "btcctl --simnet --rpcuser=devuser --rpcpass=devpass --rpccert=/rpc/rpc.cert --rpcserver=localhost $@"};
```

This lets you run commands like `alice "addinvoice 2000"` or `bob "getinfo"`. **NB:** Remember to add quotes around the command after the function as shown in the examples!


### Creating database migrations

When Torq starts up it will automatically run any new migration files.

if you wish to create a new migration file you can use the following command:

```
migrate create -seq -ext psql -dir database/migrations add_enabled_deleted_to_local_node
```

Make sure to have [golang-migrate](https://github.com/golang-migrate/migrate/tree/v4.15.2/cmd/migrate) CLI installed.
You should not bother creating a rollback migration file. We will not be supporting that in this project.

> NB! The migration itself will run once torq get booted.

## How Can I Contribute?

### Reporting Bugs

If you find a **Closed** issue that seems like it is the same thing that you're experiencing, open a new issue and include a link to the original issue in the body of your new one.

### Suggesting Enhancements

Let us know what you are missing so we can improve the software for everybody!

### Your First Code Contribution

Unsure where to begin contributing to torq? You can start by looking through these `Good first issue` and `Help wanted` issues:

- [Good first issue][good first issue] - issues which should only require a few lines of code, and a test or two.
- [Help wanted issues][help wanted] - issues which should be a bit more involved than `Good first issue` issues.

[good first issue]: https://github.com/lncapital/torq/issues?q=is%3Aopen+is%3Aissue+label%3A%22good+first+issue%22
[help wanted]: https://github.com/lncapital/torq/issues?q=is%3Aopen+is%3Aissue+label%3A%22help+wanted%22

### Code style guides

We are aware we are currently violating some of the style guides, but we are working our way through the codebase to address this.

#### Filename and variable casing

| Item                                                                                                                           | Case       |
| ------------------------------------------------------------------------------------------------------------------------------ | ---------- |
| Database columns and tables<br>go folder and file names                                                                        | snake_case |
| Javascript folder names<br>Sass file names **\*.scss**<br>Typescript file names **\*.ts**<br>JSON Keys<br>JavaScript variables | camelCase  |
| Typescript file names using JSX **\*.tsx**                                                                                     | PascalCase |
| MARKDOWN.md file names                                                                                                         | UPPERCASE  |

#### Channel ID naming convention

Internally we prefer to use the core lightning short channel id format (`777777x666x1`) for representing channel ids. Naming of channel id variables and keys should be as follows:

| Name              | Refers to                                                 |
| ----------------- | --------------------------------------------------------- |
| shortChannelId    | Core lightning format short channel id (preferred format) |
| lndShortChannelId | LND's uint64 format                                       |
| lndChannelPoint   | LND's channel point format                                |

In the serverside code there are helper functions for converting between core lightning short channel ids and lnd short channel ids.

#### Automated Formatters

For the frontend code we use Prettier and for the backend code we use gofmt. We recommend installing plugins to your editor to run both of these formatters on save. We also have an .editorconfig file to specify tabbing and spacing and depending on editor you may need to install an editor config plugin to read it. VS Code for example doesn't automatically parse .editorconfig files without a plugin.

#### Linters

Frontend: eslint
Backend: golangci-lint

Editor plugins are available for both. Linting problems will fail the build.

## Running Tests

**Running the unit tests:**
```bash
make test
```

When running `make test` fails potentially the dev database was still running so a consecutive run should work if that was the case.

**Running the e2e tests:**
```bash
make test-e2e
```

To run our full end-to-end tests similar to our github actions pipeline the command (in git bash on Windows\*) is `make test && make test-e2e-debug`

**Testing just the frontend:**
```bash
make test-frontend
```

Testing just the backend:
```bash
make start-dev-db && make wait-db && make test-backend && make stop-dev-db
```

To run a specific backend test with verbose logging `make start-dev-db && make wait-db && go test -v -count=1 ./pkg/lnd -run TestSubscribeForwardingEvents && make stop-dev-db`


### Creating a pull request

When you reached a point where you want feedback. Then it's time to create your pull request. Make sure you pull request references a GitHub issue! When an issue does not exist then it's required for you to create one so it can be referenced.
