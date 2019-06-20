# PegNet
This is now the main repository for mining for the PegNet

The TestNet has started writing records in mid June.

This is a work in progress.  The TestNet is certain to reset the actual rewards a few times as we develop more code and adjust the approach.  None the less, it is a great help to have multiple people do mining.  If you would like to help, you can run a miner.  To do so, do these 10 things:

*Assumes Linux or use of a virtual machine running Linux*

1. Clone the OracleRecord repository https://github.com/pegnet/PegNet

2. Copy the defaultconfig.ini file from the OracleRecord repository to ~/.pegnet directory

3. Sign up for an API Key from https://currencylayer.com.

4. Replace the Oracle.APILayerKey key with your API Key

5. Edit the Miner.IdentityChain to add a unique name like:

6. IdentityChain=prototype,BigMiner

7. Right now you can donate to the cause, and leave the pPNT address alone. We will had a utility to generate your own PNT address from one of your FCT addresses in your wallet

7.5 Run a local factomd instance, and run factom-walletd.  This will give you a factom node to run against, and a wallet to write entries with.  There is an Open Node configuration, and we will add instructions for that soon.

8. Buy some Entry Credits and fund one of your EC addresses in your factom-walletd wallet

9. Download the image checked into the PegNetDistribution repository

10. Create a folder and Run it.
