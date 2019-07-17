<p align="center">
  <img src="https://pegnet.org/assets/img/logo.png"/>
</p>

----

[![Build Status](https://travis-ci.com/pegnet/pegnet.svg?branch=develop)](https://travis-ci.com/pegnet/pegnet)
[![Discord](https://img.shields.io/discord/550312670528798755.svg?label=&logo=discord&logoColor=ffffff&color=7389D8&labelColor=6A7EC2)](https://discord.gg/V6T7mCW)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/pegnet/pegnet/blob/master/LICENSE)



## A Network of Pegged Tokens

This is the main repository for the PegNet application.

Pegged tokens reflect real market assets such as currencies, precious metals, cryptocurrencies, commodities, etc. The conversion rates on PegNet are determined by a decentralized set of miners who submit values based on current market data. These values are then graded and recorded in the Factom blockchain. 

The draft proposal paper is available [here](https://docs.google.com/document/d/1yv1UaOXjJLEYOvPUT_a8RowRqPX_ofBTJuPHmq6mQGQ).

For any questions, troubleshooting or further information head to [discord](https://discord.gg/V6T7mCW).

All mining is currently done on Testnet.

## Mining

#### Requirements

* Pegnet binary from the [releases page](https://github.com/pegnet/pegnet/releases)
* Factom binaries from their [distribution page](https://github.com/FactomProject/distribution/releases)
* Factom Address that is funded or Entry Credits which can be purchased directly [here](https://shop.factom.com/) or [here](https://ec.de-facto.pro/).

#### Setup

Create a `.pegnet` folder inside your home directory. Copy the `defaultconfig.ini` [file](https://raw.githubusercontent.com/pegnet/pegnet/master/defaultconfig.ini) there. 

On Windows this is your `%USERPROFILE%` folder

Linux example:
```bash
mkdir ~/.pegnet
wget https://raw.githubusercontent.com/pegnet/pegnet/master/defaultconfig.ini -P ~/.pegnet/
```

* Sign up for an API Key from https://currencylayer.com, replace APILayerKey in the config with your own

* Replace either ECAddress or FCTAddress with your own
* Modify the IdentityChain name to one of your choosing.
* Have a factomd node running on mainnet.
* Have factom-walletd open
* Start Pegnet

On first startup there will be some delay while the bytemap is shuffled. It will only start mining at the beginning of a block also.

## Developer Sandbox Setup

Docker guide can be found [here](https://github.com/pegnet/pegnet/blob/develop/Docker.md) for an automated solution.

### Manual Environment Setup

Factom binaries

https://github.com/FactomProject/distribution/releases


```bash
# In first terminal
factomd --count=5 --net=alot+ --blktime=120 --enablenet=false --network=LOCAL # change blktime to whatever suits

# Second Terminal
factom-walletd

# Third Terminal
fa='factom-cli importaddress Fs3E9gV6DXsYzf7Fqx1fVBQPQXV695eP3k5XbmHEZVRLkMdD9qCK'
ec='factom-cli newecaddress'
factom-cli listaddresses # Verify addresses
factom-cli buyec $fa $ec 100000
factom-cli balance $ec # Verify Balance

# Fork Repo on github, clone your fork
git clone https://github.com/<USER>/pegnet

# Add main pegnet repo as a remote to keep in sync
cd pegnet
git remote add upstream https://github.com/pegnet/pegnet

git pull upstream develop

# Initialize the pegnet chain
cd initialization
go build
./initialization
```

### Contributing 

* Run a testnet node
* Create an issue
* Write a feature branch on your own fork and submit a PR. Most follow the convention of being named after a particular issue  eg PNT-42_Nifty_Fix 

