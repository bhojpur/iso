# Bhojpur ISO - Image Sourcing Operations

The `Bhojpur ISO` is a secure software packaging, distribution, and management system
applied within [Bhojpur.NET Platform](https://github.com/bhojpur/platform) ecosystem
for dilivery of `applications` or `services`.

## Key Features

- ISO image burner
- Package Manager

## Installation

To install `Bhojpur ISO`, grab a release on the [Release page](https://github.com/bhojpur/iso/releases)
or to install it in your system:

```bash
$ curl https://get.bhojpur.net/iso/install.sh | sudo sh
```

To run `Bhojpur ISO` builder

```bash
$ isomake
```

To run `Bhojpur ISO` manager

```bash
$ isomgr search ...
$ isomgr install ..
$ isomgr --help
```

## Build Source Code

```bash
$ make build
```

Alternatively, you can issue the following command to build with `task` utility

```bash
$ task build-tools
```
