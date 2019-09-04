# PegNet Docker

The pegnet docker will help you get up and running for development in no time.

## Prerequisites
To be able to run PegNet you must have Docker and docker-compose, docker versions change depending on your setup.
- Docker and docker-compose installed.
- Follow instructions at https://docs.docker.com/compose/install/
- factom-cli installed in your host.

## Initial Steps
From wherever you have cloned this repo

- Create your config file from `config/defaultconfig.ini` using

```bash
cp config/defaultconfig.ini config/pegnetconfig.ini
```

- Edit the following values on `pegnetconfig.ini`
```ini
FactomdLocation="factomd:8088"
WalletdLocation="walletd:8089"
```

- After you have copied and setup your config file, run

```bash
docker-compose up
```

docker-compose will create and run 3 services:
- factomd (factomd_container)
- walletd (walletd_container)
- pegnet (pegnet_container)

You should have factomd and walletd running at this point. You will have to start the miner separately (see below).

## Initial Funding
Now that we have factomd and walletd running, we need to buy some Entry Credits and fund the default EC address, run (not in docker)

```bash
./initialization/fundEC.sh
```

This will fund the EC address *(requires factom-cli)*

---

After making sure the chains are created, run

```bash
docker-compose run --rm pegnet go run initialization/main.go
```

# Running a PegNet node
Now that we have everything set up, we can run a basic validator node:

```bash
docker-compose run --rm pegnet go run pegnet.go --log=debug
```

# Running a PegNet node with miners
Or a node with a set number of miners:

```bash
docker-compose run --rm pegnet go run pegnet.go --log=debug --miners=4
```

# Other useful commands
- run `docker-compose up -d` (running in daemon mode)
- run `docker-compose build --no-cache pegnet` (builds the pegnet image) 
- run `docker-compose run --rm pegnet bash` (run a temp pegnet container with bash)

# Destroy (reset all data)
- run `docker-compose down`
- run `docker volume rm factomd_volume_volume` (resets factomd data)
- run `docker volume rm pegnet_walletd_volume` (resets walletd data)
