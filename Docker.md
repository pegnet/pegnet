# PegNet Docker

The pegnet docker will help you get up an running for development in no time.

## Prerequisites
To be able to run PegNet you must have Docker and docker-compose, docker versions change depending of your setup.
- Docker and docker-compose installed.
- Follow instructions at https://docs.docker.com/compose/install/
- factom-cli installed in your host.

## Initial Steps
From wherever you have cloned this repo

- Create your config file from `defaultconfig.ini` using

```bash
cp defaultconfig.ini myconfig.ini
```

- After you have copied and setup your config file, run

```bash
docker-compose up
```

docker-compose will create and run 3 services:
- factomd (factomd_container)
- walletd (walletd_container)
- pegnet (pegnet_container)

You should have factomd and walletd running at this point.

## Initial Funding
Now that we have factomd and walletd running, we need to buy some Entry Credits and fund the default EC address, run

```bash
./initialization/fundEC.sh
```

This will fund the EC address *(requires factom-cli)*

---

Make sure the chains are created, run

```bash
docker-compose run --rm pegnet go run initialization/main.go
```

# Running the PegNet
Now that we have everything setup, run

```bash
docker-compose run --rm pegnet go run pegnetMining/Miner.go -log debug
```

# Other usefull commands
- run `docker-compose up -d` (running in deamon mode)
- run `docker-compose build --no-cache pegnet` (builds the penget image) 
- run `docker-compose run --rm pegnet bash` (run a temp pegnet container with bash)

# Destroy (reset all data)
- run `docker-compose down`
- run `docker volume rm factomd_volume_volume` (resets factomd data)
- run `docker volume rm pegnet_walletd_volume` (resets walletd data)