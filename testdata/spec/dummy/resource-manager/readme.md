# Dummy

> see https://aka.ms/autorest

This is the AutoRest configuration file for Dummy.

---

## Getting Started

To build the SDK for Dummy, simply [Install AutoRest](https://aka.ms/autorest/install) and in this folder, run:

> `autorest`

To see additional help and options, run:

> `autorest --help`

---

## Configuration

### Basic Information

These are the global settings for the Dummy API.

``` yaml
openapi-type: arm
tag: package-dummy-2023-05
opt-in-extensible-enums: true
```

### Tag: package-2023-05-preview

These settings apply only when `--tag=package-2023-05-preview` is specified on the command line.

```yaml $(tag) == 'package-2023-05-preview'
input-file:
  - Microsoft.Dummy/preview/2023-05-01-preview/foo.json
```

### Tag: package-2023-05

These settings apply only when `--tag=package-2023-05` is specified on the command line.

```yaml $(tag) == 'package-2023-05'
input-file:
  - Microsoft.Dummy/stable/2023-05-15/foo.json
```
