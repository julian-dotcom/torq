![Torq - Banner](./docs/images/readme-banner.png)

# Torq

![All Tests](https://github.com/lncapital/torq/actions/workflows/test-on-push.yml/badge.svg)

Torq is an advanced node management software that helps lightning node operators analyze and automate their nodes. It is designed to handle large nodes with over 1000 channels, and it offers a range of features to simplify your node management tasks, including:

* Analyze, connect and manage all your nodes from one place!
* Access a complete overview of all channels instantly.
* Build advanced automation workflows to automate Rebalancing, Channel Policy, Tagging and eventually any node action.
* Review forwarding history, both current and at any point in history.
* Customize and save table views. Containing only selected columns, advanced sorting and high fidelity filters.
* Export table data as CSV. Finally get all your forwarding or channel data as CSV files.
* Enjoy advanced charts to visualize your node's performance and make informed decisions.

Whether you're running a small or a large node, Torq can help you optimize its performance and streamline your node management process. Give it a try and see how it can simplify your node management tasks.

![torq-automation-preview](https://user-images.githubusercontent.com/647617/223672620-dcc047f3-ebbe-4087-8da8-9a103d8b9570.png)


## Quick start

### configuration

Torq supports a TOML configuration file. The docker compose install script auto generates this file.
You can find an example configuration file at [example-torq.conf](./docker/example-torq.conf)

### docker compose
To install Torq via docker compose:

```sh
sh -c "$(curl -fsSL https://torq.sh)"
```
You do not need sudo/root to run this and you can check the contents of the installation script here: https://torq.sh

### podman
To run the database via host network:

```sh
podman run -d --name torqdb --network=host -v torq_db:/var/lib/postgresql/data -e POSTGRES_PASSWORD="<YourPostgresPasswordHere>" timescale/timescaledb:latest-pg14
```

To run Torq via via host network:

First create your TOML configuration file and store it in `~/.torq/torq.conf`

```sh
podman run -d --name torq --network=host -v ~/.torq/torq.conf:/home/torq/torq.conf lncapital/torq:latest --config=/home/torq/torq.conf start
```

### kubernetes

We shared templates for CRDs in folder [kubernetes](./kubernetes).
This folder also has it's own readme.

### network

Be aware then when you try Torq on testnet, simnet or some other type of network that you use the network switch when trying to browse the web interface.
Then network switch is the globe icon in the top left corner, next to the Torq logo.

## How to Videos

[You can find the full list of video guides here.](https://docs.ln.capital/en/collections/3817618-torq-video-tutorials)

### How to create custom Channel Views

[![Torq Forwarding Views YouTube Guide](https://img.youtube.com/vi/5ZfgflfOFwQ/maxresdefault.jpg)](https://www.youtube.com/watch?v=5ZfgflfOFwQ)

### How to use Automation Workflows

[![Torq Workflow Automation YouTube Guide](https://img.youtube.com/vi/Go4uJoMhwrE/maxresdefault.jpg)](https://www.youtube.com/watch?v=Go4uJoMhwrE)

### How to use the Forwards Tab

[![Torq Forwarding Views YouTube Guide](https://img.youtube.com/vi/ZTetH8_jbgk/maxresdefault.jpg)](https://www.youtube.com/watch?v=ZTetH8_jbgk)


## Permissions

Since Torq is built to manage your node, it needs most/all permissions to be fully functional. However, if you want to
be extra careful you can disable some permissions that are not strictly needed.

Torq does not for now need the ability to create new macaroon or stop the LND daemon,

    lncli bakemacaroon \
        invoices:read \
        invoices:write \
        onchain:read \
        onchain:write \
        offchain:read \
        offchain:write \
        address:read \
        address:write \
        message:read \
        message:write \
        peers:read \
        peers:write \
        info:read \
        uri:/lnrpc.Lightning/UpdateChannelPolicy \
        --save_to=torq.macaroon

Here is an example of a macaroon that can be used if you want to prevent all actions that sends funds from your node:

    lncli bakemacaroon \
        invoices:read \
        invoices:write \
        onchain:read \
        offchain:read \
        address:read \
        address:write \
        message:read \
        message:write \
        peers:read \
        peers:write \
        info:read \
        uri:/lnrpc.Lightning/UpdateChannelPolicy \
        --save_to=torq.macaroon

## Help and feedback

Join our [Telegram group](https://t.me/joinchat/V-Dks6zjBK4xZWY0) if you need help getting started.
Feel free to ping us in the telegram group if you have any feature request or feedback.  We would also love to hear your ideas for features or any other feedback you might have.
