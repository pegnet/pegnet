#!/usr/bin/env bash
#
# Default is to fund a simulation EC address for testing
# Replace the EC address and the FCT address with your own to initialize a PegNet
# on the Factom MainNet
EC="EC3TsJHUs8bzbbVnratBafub6toRYdgzgbR7kWwCW4tqbmyySRmg"
FCT="FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q"

# Imports addresses for the Local Testnet.  These addresses are no good for the mainnet, but
# for ease of developing on local simulated networks, we add them to your wallet here.
factom-cli importaddress Fs3E9gV6DXsYzf7Fqx1fVBQPQXV695eP3k5XbmHEZVRLkMdD9qCK
factom-cli importaddress Es2XT3jSxi1xqrDvS5JERM3W3jh1awRHuyoahn3hbQLyfEi1jvbq

# Fund the EC address
# Creates and sends a FCT tx that adds 100 FCT worth of ECs to the given EC address
factom-cli rmtx tx
factom-cli newtx tx
factom-cli addtxinput tx $FCT 100
factom-cli addtxecoutput tx $EC 100
factom-cli addtxfee tx $FCT
factom-cli signtx tx
factom-cli sendtx tx

factom-cli listaddresses
