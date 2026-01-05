# nowsecure-ci

A command-line tool for integrating NowSecure security assessments into your CI/CD pipeline. This tool enables automated mobile application security testing for both Android and iOS applications.

> [!NOTE]  
> If you're looking for ready-made CI/CD integrations, look at one of the following:
>
> - [Azure](https://github.com/nowsecure/nowsecure-azure-ci-extension/)
> - [GitHub Actions](https://github.com/nowsecure/nowsecure-action/)
> - [Gitlab Components](https://gitlab.com/nowsecure.com/nowsecure-ci-component)
> - [CircleCI](https://github.com/nowsecure/nowsecure-circle-ci-orb)
> - [Jenkins Plugin](https://github.com/jenkinsci/nowsecure-ci-assessments-plugin)

## Installation

### From Source

```bash
git clone https://github.com/nowsecure/nowsecure-ci.git
cd nowsecure-ci
go build -o ns
```

### Using Go Install

```bash
go install github.com/nowsecure/nowsecure-ci@latest
```

## Prerequisites

Before using this tool, you need:

1. A token from your NowSecure platform instance. More information on this can be found in the [NowSecure Support Portal](https://support.nowsecure.com/hc/en-us/articles/7499657262093-Creating-a-NowSecure-Platform-API-Bearer-Token).
2. Your organization's group UUID from the NowSecure Platform. More information on this can be found in the
  [NowSecure Support Portal](https://support.nowsecure.com/hc/en-us/articles/38057956447757-Retrieve-Reference-and-ID-Numbers-for-API-Use-Task-ID-Group-App-and-Assessment-Ref).

## Configuration

The tool can be configured using command-line flags, environment variables, or a configuration file.

### Environment Variables

All flags can be set via environment variables with the `NS_` prefix:

```bash
export NS_TOKEN="your-api-token"
export NS_GROUP_REF="your-group-uuid"
```

### Configuration File

Create a `.ns-ci.yaml` file in your project root or home directory:

```yaml
token: your-api-token
group_ref: your-group-uuid
```

### Command-Line Flags

Flags can be provided explicitly as part of the CLI command itself

``` bash
ns run file ./path/to/app.apk \
  --group-ref YOUR_GROUP_UUID 
```

## Usage

The tool provides three methods to run security assessments:

### Run Assessment by Uploading a Binary File

Upload and analyze a mobile application binary (APK or IPA file):

```bash
ns run file ./path/to/app.apk \
  --group-ref YOUR_GROUP_UUID \
  --analysis-type full \
  --poll-for-minutes 60 \
  --minimum-score 70 \
  --save-findings \
  --artifacts-dir ./artifacts
```

### Run Assessment by Package Name

Trigger an assessment for an existing application using its package name and platform:

```bash
ns run package com.example.myapp \
  --android \
  --group-ref YOUR_GROUP_UUID \
  --analysis-type full \
  --poll-for-minutes 60 \
  --minimum-score 75
```

### Run Assessment by Application ID

Run an assessment using a pre-existing application's UUID:

```bash
ns run id aaaaaaaa-1111-bbbb-2222-cccccccccccc \
  --group-ref YOUR_GROUP_UUID \
  --analysis-type full \
  --poll-for-minutes 60 \
  --minimum-score 80 \
  --save-findings
```
