![Torq - Banner](./docs/images/readme-banner.png)

# Torq

![All Tests](https://github.com/lncapital/torq/actions/workflows/test-on-push.yml/badge.svg)

Torq is an advanced node management software that helps lightning node operators analyze and automate their nodes. It is designed to handle large nodes with over 1000 channels, and it offers a range of features to simplify your node management tasks, including:

* Simultaneously connect to and analyze multiple nodes
* Access a complete overview of all channels in one place
* Build advanced automation workflows to automate any node action
* Review forwarding history, both current and historical
* Easily filter and sort data with high fidelity
* Store commonly used filter configurations for quick access to your preferred table views.
* Enjoy advanced charts to visualize your node's performance and make informed decisions.

Whether you're running a small or a large node, Torq can help you optimize its performance and streamline your node management process. Give it a try and see how it can simplify your node management tasks.

![torq-automation-preview](https://user-images.githubusercontent.com/647617/223672620-dcc047f3-ebbe-4087-8da8-9a103d8b9570.png)


## Quick start

To install Torq run:

```sh
sh -c "$(curl -fsSL https://torq.sh)"
```

You do not need sudo/root to run this and you can check the contents of the installation script here: https://torq.sh

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
