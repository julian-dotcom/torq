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

Optional step if you already have a PostgreSQL database with TimescaleDB plugin. If not create one with the following command.

``` sh
docker run -d --name timescaledb -p 5432:5432 \
-v <AbsoluteDataPath>:/var/lib/postgresql/data \
-e POSTGRES_PASSWORD=<DBPassword> timescale/timescaledb:latest-pg14
```

### Create a database and user for Torq

Shell into the timescale container to run `psql`.

``` sh
docker exec -it timescaledb bash
psql -U postgres
```

Inside the postgres interactive terminal run the following three SQL commands to create a database and user.

```postgresql
CREATE DATABASE torq;
CREATE USER torq WITH ENCRYPTED PASSWORD '<DBPassword>';
GRANT ALL PRIVILEGES ON DATABASE torq TO torq;
```

After creating the database, exit psql and the TimescaleDB container by hitting `CTRL d` twice.

### Run Torq

At present Torq only connects to a single LND node. To run Torq provide the IP, Port, TLS cert and Macaroon of your LND Node as well as the database password set above. Database name and user are configurable but both default to `torq`.

``` sh
docker run -p 8080:8080 --rm  \
-v <AbsolutePathLNDTLSCert>:/app/tls.cert \
-v <AbsolutePathLNDMacaroon>:/app/readonly.macaroon \
lncapital/torq --lnd.macaroon /app/readonly.macaroon \
--lnd.node_address <IP:Port of LND Node e.g. 190.190.190.190:10009> --lnd.tls /app/tls.cert \
--db.password <DBPassword> --db.host host.docker.internal start
```
