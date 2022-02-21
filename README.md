# Torq by LN.capital

Torq is a capital management tool for routing nodes on the lightning network.

This software is in active development and should be considered alpha 

## Install

To install the backend, simply run `go install ./cmd/torq`

### Create a database and user for Torq

```postgresql
CREATE DATABASE <torq_db>;
CREATE USER <torq_user> WITH ENCRYPTED PASSWORD '<torq_user_password>';
GRANT ALL PRIVILEGES ON DATABASE <torq_db> TO <torq_user>;
```

## Requirements

Torq uses a postgres database with the TimescaleDB plugin.
