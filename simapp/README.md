---
sidebar_position: 1
---

# `SimApp`

`SimApp` is an application built using the Cosmos SDK and Symbiotic specific Staking and Slashing modules.

## Running SimApp with Symbiotic Relay 

1. Need go 1.24+
2. All the following commands are to be run in the root of repo.
2. Run `make build`.
3. Run `./build/simd testnet gen-keys -v 5` this generates 5 validator keys and stores them in `.testnets/keys` file.
4. The Key file contains the node consensus private keys (ed25519), these keys should be added to the validator info in the relay key registry with the key tag - 43
5. You can also use the `MockRelayClient` implementation for testing purposes which doesn't need any relay contracts or sidecar setup. You can specify all the epoch validator keys as a json file, see example in `valkeys_example.json` which uses keys from `keys_example` file. You can also use these files directly with your setup by replacing your `keys` with ones in the example to test the chain.
6. Once you have the keys registered either with the original relay contracts or with the mock relay file run `./build/simd testnet setup --priv-keys-file=.testnets/keys --chain-id=chain-xyz`
7. The testnet setup command will generate all the home dirs for your testnet nodes with the consensus private keys previously generated.
8. Once the setup is ready run the following command, replace the home dir for each node with its respective generated dir path :
```
# if you are using mock client with valkeys.json file in .testnets/valkeys.json
SYMBIOTIC_KEY_FILE=.testnets/valkeys.json ./build/simd start --home=.testnets/chain-xyz/node0/simd
# or if you want to run with real relay sidecar (make sure to connect each cosmos node and relay sidecar 1:1)
SYMBIOTIC_RELAY_RPC=localhost:8080 ./build/simd start --home=.testnets/chain-xyz/node0/simd
```
9. Repeat the above command for each node with its respective home dir.
10. You can validate the validator set updates by running :
```
./build/simd q consensus comet validator-set
```
11. You should notice as soon as the relay client updates the validator set the validator set displayed by the above command also updates. Cosmos chain by default will check for updates from relay every 10 cosmos blocks.
12. For checking slashing module in action, ensure to have at least 4 validators registered, and then you can stop 1 validator. After 50 block signature misses of the validator you should see an event from cosmos chain slashing the validator.
13. To watch for slash events use this :
```
wscat -c ws://127.0.0.1:26657/websocket # or any websocket client like postman

# and run the following message to receive slash event blocks

{ "jsonrpc": "2.0", "method": "subscribe", "id": 1, "params": { "query": "tm.event='NewBlock' AND slash.address EXISTS" } }
```
14. The slashing event will look something like this :
```json
{
    "type": "slash",
    "attributes": [
        {
            "key": "address",
            "value": "cosmosvalcons1h8lgteknhy0qn3w7h7j4kj6elr7mg9zegk9djt",
            "index": true
        },
        {
            "key": "power",
            "value": "10000",
            "index": true
        },
        {
            "key": "reason",
            "value": "missing_signature",
            "index": true
        },
        {
            "key": "jailed",
            "value": "cosmosvalcons1h8lgteknhy0qn3w7h7j4kj6elr7mg9zegk9djt",
            "index": true
        },
        {
            "key": "slash_request_hash",
            "value": "0xbf821c8892c225af145fe89ac0e042c9cb844d8cbb97ef3bb4af73d99969750b",
            "index": true
        },
        {
            "key": "mode",
            "value": "BeginBlock",
            "index": true
        }
    ]
}
```
The `slash_request_hash` is the relay signature request id for the slash signature. External services will have to monitor this event and get the aggregated proof using this hash and submit it to relay contract for slashing the validator.

# Changes made to cosmos-sdk

This demo application is built on top of cosmos v0.53.4, but makes changes to its original `x/slashing` and `x/staking` modules to effectively disable those and have the symbiotic replacements `x/symslashing` and `x/symstaking` take over the responsibilities.
The modules that still depend on the original `x/slashing` and `x/staking` modules may not be compatible with the symbiotic replacements out of the box.