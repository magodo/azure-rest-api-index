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

## How are the Swaggers collected?

The very first thing is to collect all the *valid* swagger files. The tool will walk the *<specs rootdir>/specification* folder recursively, and look for *readme.md* file, which is maintained by the service team respectively to record all the *valid* swagger files for each version/package. Note that during this walk, the data plane folders (i.e. *data-plane*) or example folder (i.e. *examples*) are skipped.

Once all the Swaggers are collected, the index can be built based on that. However, for the same combination of `(RP, Version, Method, RT [, Action])`, there can be multiple swaggers have the operation defined. This is mostly due to in earlier days, some of the Swagger that has cross RP reference tended to copy the depending swagger over, making duplications among the repo. The tool uses the following process to deduplicate this:

1. (Auto) If the RP name detected in the API path exactly matches one of the candidate swagger file's file paths, then regard that file is the original definition of this API path, and eliminating the other candidates.
2. (Manual) For the remaining ones, we maintained a file to pick the correct file, at: *./azidx/dedup.json*.

## Index Format

Following is an explaination about the format of the generated index file from the `build` command:

```json
{
    "commit": "<commit_id>",
    "resource_providers": {
        "<rp_name>": {
            "<api_version>": {
                "<operation>": {
                    "<resource_type>": {
                        "actions": {
                            "<action_name>": {
                                "<api_path_pattern>": "<json_reference>",
                                ...
                            }
                        },
                        "operation_refs": {
                            "<api_path_pattern>": "<json_reference>",
                            ...
                        }
                    },
                    ...
                },
                ...
            },
            ...
        },
        ...
    }
}
```

- `commit_id`: From which Git commit of Azure/azure-rest-api-specs this file is generated.
- `rp_name`: RP name in upper case (e.g. `MICROSOFT.FOO`). Especially, it can be `*`, which indicates the most relavent RP name is a parameter in the API path.
- `api_version`: The api version (e.g. `2020-01-01`)
- `operation`: The operation in upper case (e.g. `GET`)
- `resource_type`: The resource type in upper case (e.g. `VIRTUALNETWORKS/SUBNETS`)
- `action_name`: (Optional) The name of the action for resource type in upper case.

    This can be the *list* of a child resource of the resource type in scope. E.g. 

          "/VIRTUALNETWORKS": {
            "actions": {
              "SUBNETS": {...}
            }
          }
    
    This represents the `GET` against the API path `.../virtualNetworks/<vnet_name>/subnets`.

    This can also be the *POST* action for the resource type in scope. E.g.

        "/SERVICE/GATEWAYS": {
            "actions": {
              "GENERATETOKEN": {...}
            }
        }
    
    This represents the `POST` against the API path `.../services/<service_name>/gateways/<gateway_name>/generateToken`.

- `operation_refs`: (Optional) The regular API path of the resource type in scope.
- `api_path_pattern`: The API path pattern defined in the Swagger.  Especially, the path segment is marked as either `{}` or `{*}`, where the latter one represents the path segment is decorated with the [`x-ms-skip-url-encoding`](https://azure.github.io/autorest/extensions/#x-ms-skip-url-encoding).

    Note that there can be more than one combination of `api_path_pattern: json_reference`, the reason is that the operation can exist under different scope, e.g. under a resource group, a subscription, or/and a tenant.

- `json_reference`: The [JSON schema reference](https://json-schema.org/draft/2020-12/json-schema-core#name-schema-references) to the Swagger definition of the current operation.
