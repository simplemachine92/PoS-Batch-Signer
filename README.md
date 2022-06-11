# Proof Of Stake Pages EIP712 CLI Signer

### A tool written to batch sign EIP712 compliant messages necessary for **Soulbound** NFTs issued to Public Goods Fundoors participating in the Proof Of Stake (digital) Book Launch.

This tool uses a Firebase service-account to read/write EIP-712 Signatures to a Firebase Realtime DB
https://firebase.google.com/docs/database/admin/start

We compare a list of Pledge events emitted from our Smart Contract with our database in order to sign and store EIP712 messages to pledgees who have not yet received a message.

Feel free to modify for your own purposes. In this particular case, parser.js messagePacks and lzw encodes our TypedData for delivery to our front-end a'la [Signator.io](https://github.com/scaffold-eth/scaffold-eth/tree/signatorio)

Right now, due to muh' bear market constraints, this tool lacks the more granular control over signatures than planned, but I may add them later..

**In the case of PoS, I will provide both the .env and service-account-key.json for Vitaliks convenience**

He should only have to modify PRIVATE_KEY & PUBLIC_KEY in .env to operate..

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [Support](#support)
- [Contributing](#contributing)

## Installation

### Clone the repo

```sh
git clone https://github.com/simplemachine92/PoS-Batch-Signer.git
cd pos-batch-signer
```

### Install Node deps

```sh
npm i
```

### Copy example.env to .env and modify with your variables

```sh
cp example.env .env
```

### Copy service-account-example.json to service-account-key.json and modify with your service-worker..

Service Account Reference: [Firebase Docs](https://firebase.google.com/support/guides/service-accounts)

```sh
cp service-account-example.json service-account-key.json
```

Be sure to remove the comment at the top if replacing the vars and not using your own file.

## Usage

### This tool tested and working on go version go1.18.1 darwin/arm64

go version go1.18.1 should work regardless of architecture

### There are two commands for now..

sign

```sh
go run main.go sign -m='Your Message'
```

or listen

```sh
go run main.go listen -l='Your Message'
```

### **Don't remove the quotes!** For example..

```sh
go run main.go sign -m='howdy, fellers!'
```

### You should see similar logging if the command ran successfully:

```sh
Signing to all pending users with Msg: howdy, fellers!

Donation Events Total: 16

Unique Donation Events: 13

Unique Signatures (DB): 13

Sigs Generated This Run: 0
```

or for listen command:

```sh
Signed and stored a message for address 0xEB00BA1C44995119c55279d4F628ac19d4d35f7d

Signed so far while listening: 1
```

## A note on Messages

Messages are restricted to < 60 characters due to our on-chain svg's limitations.

Empty Message strings will also result in an error.

## Support

Please [open an issue](https://github.com/simplemachine92/PoS-Batch-Signer/issues/new) for support, or dm me [@simplemachine92](https://twitter.com/SimpleMachine92)
