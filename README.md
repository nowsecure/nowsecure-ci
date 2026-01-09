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
2. A valid group UUID from the NowSecure Platform. More information on this can be found in the
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

- `ns run file`
- `ns run package`
- `ns run id`

### Available Parameters

#### Required Parameters

- `--group-ref` - A valid group reference from NowSecure Platform
- `--token` - Authentication token for the NowSecure Platform API

#### API Configuration

- `--api-host` - REST API base URL (default: `https://lab-api.nowsecure.com`)
  - Use this to point to a different NowSecure endpoint if you are accessing a single tenant instance
  
- `--ui-host` - UI base URL (default: `https://app.nowsecure.com`)
  - Use this to point to a different NowSecure instance is you are accessing a single tenant instance

#### Analysis Type

- `--analysis-type` - Type of assessment to run (default: `full`)
  - `full` - Complete security assessment including dynamic and static analysis
  - `static` - Static analysis only (requires `--android` or `--ios` platform flag)
  - `sbom` - Software Bill of Materials generation

#### Platform Selection (for `run package` and `run static`)

- `--android` - Specify that the application platform is Android
- `--ios` - Specify that the application platform is iOS

**Note:** These flags are mutually exclusive. You must provide exactly one when using `run package` or when running static analysis.

#### Polling and Results

- `--poll-for-minutes` - Maximum duration in minutes to poll for assessment results (default: `60`)
  - Set to `0` to trigger the assessment without waiting for results
  - Required to be greater than `0` when using `--save-findings`

- `--minimum-score` - Minimum acceptable security score threshold (default: `0`)
  - If the assessment score falls below this value, the command exits with code 1
  - Score range is 0-100

#### Artifacts and Findings

- `--save-findings` - Fetch and save all findings from the assessment (default: `false`)
  - Findings are written to `findings.json` in the artifacts directory
  - Requires `--poll-for-minutes` to be greater than 0

- `--artifacts-dir` - Directory path where artifacts should be saved (default: current working directory)
  - Used in conjunction with `--save-findings`

### Usage Examples

#### Run Assessment by Uploading a Binary File

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

#### Run Assessment by Package Name

Trigger an assessment for an existing application using its package name and platform:

```bash
ns run package com.example.myapp \
  --android \
  --group-ref YOUR_GROUP_UUID \
  --analysis-type full \
  --poll-for-minutes 60 \
  --minimum-score 75
```

**Note:** When using `run package`, you must specify either `--android` or `--ios` to indicate the platform.

#### Run Assessment by Application ID

Run an assessment using a pre-existing application's UUID:

```bash
ns run id aaaaaaaa-1111-bbbb-2222-cccccccccccc \
  --group-ref YOUR_GROUP_UUID \
  --analysis-type full \
  --poll-for-minutes 60 \
  --minimum-score 80 \
  --save-findings
```

#### Static Analysis for iOS

```bash
ns run package com.example.myapp \
  --ios \
  --analysis-type static \
  --group-ref YOUR_GROUP_UUID
```

#### Static Analysis for Android

```bash
ns run package com.example.myapp \
  --android \
  --analysis-type static \
  --group-ref YOUR_GROUP_UUID
```

#### SBOM Generation

```bash
ns run file ./path/to/app.apk \
  --analysis-type sbom \
  --group-ref YOUR_GROUP_UUID
```

#### Trigger Without Waiting for Results

```bash
ns run file ./path/to/app.ipa \
  --group-ref YOUR_GROUP_UUID \
  --poll-for-minutes 0
```
