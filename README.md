# GoGameServer
A simple message driven game server in go.

### Requirement

* gnet
* RabbitMQ
* redis
* etcd

### Build

### Run
* Linux
* Mac
* Windows

### Docker

Build image:

```bash
docker build -t gogameserver .
```

Start all services with docker compose:

```bash
docker-compose up
```

The default configuration file is located at `bin/config/config.toml` and is copied into the image as `/app/config.toml`. You can provide additional command line parameters to specify which service to run, for example:

```bash
docker run gogameserver run game 0
```

### JetBrains Open Source licenses

`GoGameServer` had been being developed with `GoLand` IDE under the **free JetBrains Open Source license(s)** granted by JetBrains.

<a href="https://www.jetbrains.com/?from=gnet" target="_blank"><img src="https://resources.jetbrains.com/storage/products/company/brand/logos/jb_beam.png" width="250" align="middle"/></a>

