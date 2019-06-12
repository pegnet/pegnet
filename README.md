# OracleRecord
This is now the main repository for mining for the PegNet

So I started 13 miners on the PegNet TestNet.

We need to add some verification code to the miners, and a few other details, but otherwise anyone can mine on the TestNet. To do so, do these 10 things:

*Assumes Linux or use of a virtual machine running Linux*

1. Clone the OracleRecord repository https://github.com/pegnet/OracleRecord

2. Copy the defaultconfig.ini file from the OracleRecord repository to ~/.pegnet directory

3. Sign up for an API Key from https://currencylayer.com.

4. Replace the Oracle.APILayerKey key with your API Key

5. Edit the Miner.IdentityChain to add a unique name like:

6. IdentityChain=prototype,BigMiner

7. Right now you can donate to the cause, and leave the pPNT address alone. We will had a utility to generate your own PNT address from one of your FCT addresses in your wallet

8. Buy some Entry Credits and fund one of your EC addresses in your factom-walletd wallet

9. Download the image checked into the PegNetDistribution repository

10. Create a folder and Run it.
