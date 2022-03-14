# Torq by LN.capital

Torq is a capital management tool for routing nodes on the lightning network.

This software is in active development and should be considered alpha 

## Install

To install the backend, simply run `go install ./cmd/torq`

## Requirements

Torq uses a postgres database with the TimescaleDB plugin.

## Run With Docker

Torq can be run from a prebuilt docker image:

### Run TimescaleDB container

Optional step if you already have a PostgreSQL database with TimescaleDB plugin. If not run one with the following command.

``` sh
docker run -d --name timescaledb -p 5432:5432 \
-v <AbsoluteDataPath>:/var/lib/postgresql/data \
-e POSTGRES_PASSWORD=<DBPassword> timescale/timescaledb:latest-pg14
```

### Create a database and user for Torq

```postgresql
CREATE DATABASE torq;
CREATE USER torq WITH ENCRYPTED PASSWORD '<DBPassword>';
GRANT ALL PRIVILEGES ON DATABASE torq TO torq;
```

### Run Torq

At present Torq only connects to a single LND node. To run, provide the IP, port, TLS cert and Macaroon of your LND Node as well as the database password set above. Database name and user are configurable but both default to `torq`.

``` sh
docker run -p 50050:50050 --rm  \
-v <AbsolutePathLNDTLSCert>:/app/tls.cert \
-v <AbsolutePathLNDMacaroon>:/app/readonly.macaroon \
lncapital/torq --lnd.macaroon /app/readonly.macaroon \
--lnd.node_address <IP:Port of LND Node e.g. 190.190.190.190:10009> --lnd.tls /app/tls.cert \
--db.password <DBPassword> --db.host host.docker.internal start
```
