# azure-rest-api-index

A tool/library to index https://github.com/Azure/azure-rest-api-specs, which is able to map a live request to its definition in the swagger spec.

## Usage

First, you'll need to build the index by running:

```shell
azure-rest-api-index build -o index.json <specs rootdir>/specification
```

The *dedup.json* above is a file used for resolving duplicated swagger definitions, which is maintained by the repo.

After the index is built, you can then lookup for any live request by the `lookup` subcommand. E.g. to look up a `GET` of a resource group, you can do:

```shell
azure-rest-api-index lookup -index index.json -method=GET -url "https://management.azure.com/subscriptions/sub1/resourceGroups/rg1?api-version=2022-09-01"
```
