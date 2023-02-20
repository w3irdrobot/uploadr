# Uploadr

Uploadr is a simple server aimed at enabling [Nostr](https://github.com/nostr-protocol/nostr) users to host their own images in a simple way. Along with a simple integration into a client, a user can have an image uploaded to their Uploadr server and later displayed, all without using centralized services such as [nostr.build](https://nostr.build/) and [inosta.cc](https://inosta.cc/) (No shade thrown at either project. They are both great and [Inosta is OSS](https://github.com/johnongit/INOSTA-Nostr-Img-Service)). The idea is that a client could allow a user to set the location of their Uploadr instance, and from the on images would be uploaded their and the returned URL would be inserted into the broadcasted note.

## Deployment

Prebuilt binaries are available [in the GitHub releases](https://github.com/KoalaSat/nostros/releases/latest). The suggested way to get up and running is to copy the [example configuration](./config.example.toml) and change it as is needed for the specific deployment.

## Development

The service is written in simple Golang. There isn't much to it really. A domain will be required to be configured using the command line or in a TOML config file that the service is told about using the command line as well.

```shell
go run . --domain http://localhost:8080
```

or

```shell
cp config.example.toml config.toml
go run . --config config.toml
```


This will start up a service on port `:8080` by default.
