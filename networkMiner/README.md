<!-- ToC start -->
# Table of Contents

1. [Pegnet Network Coordinator](#pegnet-network-coordinator)
1. [The Polling Part --> Network Coordinator](#the-polling-part--->-network-coordinator)
1. [The PoW Part --> Network Miner](#the-pow-part--->-network-miner)
   1. [Start of mining](#start-of-mining)
   1. [The stop of mining](#the-stop-of-mining)
1. [Other Network Mining Features](#other-network-mining-features)
   1. [Network Mining Security](#network-mining-security)
1. [Running the Network Mining Setup](#running-the-network-mining-setup)


<!-- ToC end -->


# Pegnet Network Coordinator

I am going to assume you read the documentation on mining, and have a general understanding of how a miner works. To be brief though a miner consists of a few parts:

1. Polling
 - The miner needs access to a factomd and pricing data sources to poll for the oracle price record (OPR) data, and submit them every block.

2. The hashing PoW
 - Given the polling data, the miner needs to repeatly hash the opr, and when the polling indicates the block is about to close, the miner should submit the best X hashes it has collected. (this is the simple algo atm)

This means, for every miner you setup, you horizontally scale your polling. If you have 10 miners, that is a lot. So instead we split the miner into 2 parts. 1 part polling, and 1 part PoW.

# The Polling Part --> Network Coordinator

What is the network coordinator?

The network coodinator is a daemon that polls factomd and price data sources and forwards the information to the network miners. The data that is sent to miners:

- Full OPRs that contain all the price information, block height, etc
- Factomd events (every minute and block)
    - One of these events tells the miners to stop mining for a block

The coordinator can also receive information from miners:

- Entries that contain the OPR + nonce. The coordinator will create the commit/reveal (paying for it) and submit the entry to Factom.
- Statistics to enable the control panel to display the miner's hashrate and some other information
- Some extra control related information like password challenges (will be described below)


# The PoW Part --> Network Miner

The network miner does not talk to a factomd, a data source, or anything else. It will ONLY talk to a coordinator, meaning you can place these behind LANs and they do not need internet access. They only need access to the network coordinator.

The network miner will listen for events to start and stop mining, submitting their best work at the end of a mining period.

## Start of mining

When a network miner receives the signal to start mining, the miner will replace the **Identity** and **coinbase payout** in the OPR with their own. This means individual miners can different payout addresses, and more importantly the identity for different miners **MUST** be different. I repeat, the identity between miners MUST be different. If they are the same, your identical miners will hash the exact same search space, wasting hashpower.

The miners will continue to hash until the stop signal comes in

## The stop of mining

When the coordinator tells the miner to stop, the miner will submit it's best X records to the coordinator. All the records will be written. In the current implementation, the network coordinator does not filter the entries that get written to the blockchain. This is a TODO.

So if you have 5 network miners, all submitting their best 3 OPRs by difficulty, then the network coordinator will submit 15 entries per block. There will be a change coming to allow a network coordinator to aggregate by payout address.

# Other Network Mining Features

## Network Mining Security

To ensure random network miners do not connect to your network coordinator, the coordinator and miner can enforce a basic authentication scheme. By enabling the authentication (on by default), the miner and the coordinator need to both contain the same shared secret. The coordinator then upon receiving a connection request, will issue a challenge to the miner, where the miner needs to hash the challenge along with the password, and return it. This basic challenge means the password is never relayed across a network, and packet sniffing would not allow any 3rd party to join, as the challenge changes with each connection request.

# Running the Network Mining Setup

If you read the mining documentation, then you are halfway there to running a network mining setup. The configuration of a network coordinator is the exact same as a regular miner. You will need to read those docs to setup an ECaddress and orcale price locations. It will just not do the PoW mining. The additional configurations include the shared secret located in the config file, and the ability to change the listening port. To launch:

```bash
# caddr will change the port it is hosted on. This can also be changed in the config file
pegnet netcoordinator --caddr :1234
```

Once the netcoordinator is running, you can now run netminers to communicate with the coordinator. The configuration needed is:

- `MiningCoordinatorHost` in the config or `--caddr` for the ip:port of the coordinator.
- `CoordinatorSecret` should match the coordinator
- `UseCoordinatorAuthentication` should be true
- `NumberOfMiners` or `--miners` should not exceed your core count.
- `RecordsPerBlock` or `--top` determines how many entries per block to submit
- `IdentityChain` Make this unique PER netminer
- `FCTAddress` and `CoinbaseAddress` should be set to an address you control

```bash
# For a 4 core machine and submit 5 entries per block
pegnet netminer --caddr localhost:1234 --miners 4 --top 5
```
