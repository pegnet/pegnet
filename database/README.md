# DATABASE
So what we need is a simple key/value store database with buckets.  This will be 
used to keep up with balances at addresses from mining rewards, and may be 
extended to include balances at addresses after conversions and transactions

The database ensures that we don't have to reprocess the entire PegNet chains
everytime someone launches the PegNet.