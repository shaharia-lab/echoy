# Echoy
Echoy - intelligent &amp; smart AI assistance for your daily life

## Installation

```shell
curl -fsSL https://raw.githubusercontent.com/shaharia-lab/echoy/main/setup.sh | bash
```

## Development

### Generating mocks

All interfaces have generated mocks in their own package using [Mockery V2](https://github.com/vektra/mockery).

```shell
go install github.com/vektra/mockery/v2@latest
```

To update or generate new mocks run:

```shell
mockery
```

To ignore a specific package or make customizations please refer to the `.mockery.yaml` config file and Mockery documentation.
Don't forget to commit the generated mocks to the repository.