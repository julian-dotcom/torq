![Torq - Banner](./docs/images/readme-banner.png)

# Torq

Torq is a node management tool for large lightning network nodes.

![All Tests](https://github.com/lncapital/torq/actions/workflows/test-on-push.yml/badge.svg)

## Quick start

To install Torq run:

```sh
sh -c "$(curl -fsSL https://torq.sh)"
```

You do not need sudo/root to run this and you can check the contents of the installation script here: https://torq.sh

## Permissions

Since Torq is built to manage your node, it needs most/all permissions to be fully functional. However, if you want to
be extra careful you can disable some permissions that are not strictly needed.

**NB: You sometimes need to restart Torq after updating the macaroon.
Wait until the save button in the UI is green before restarting Torq.**

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
