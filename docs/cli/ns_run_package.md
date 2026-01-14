## ns run package

Run an assessment for a pre-existing app by specifying package and platform

```
ns run package [package-name] [flags]
```

### Examples

```
# Recommended Flags
ns run package [package-name] \
  --android \
  --group-ref YOUR_GROUP_UUID \
  --analysis-type static \
  --poll-for-minutes 30

# Run an Assessment Without Waiting for Results
ns run package [package-name] \
  --android \
  --group-ref YOUR_GROUP_UUID \
  --poll-for-minutes 0

# Run a Full (Dynamic and Static) Assessment
ns run package [package-name] \
  --android \
  --analysis-type full \
  --group-ref YOUR_GROUP_UUID \
  --poll-for-minutes 60

# Run an Assessment With a Score Threshold
ns run package [package-name] \
  --android \
  --analysis-type static \
  --minimum-score 70 \
  --poll-for-minutes 60 \
  --group-ref YOUR_GROUP_UUID

```

### Options

```
      --android   app is for android platform
  -h, --help      help for package
      --ios       app is for ios platform
```

### Options inherited from parent commands

```
      --analysis-type string    One of: full, static, sbom (default "full")
      --api-host string         REST API base url (default "https://lab-api.nowsecure.com")
      --artifacts-dir string    directory in which to put artifacts (default "$PWD")
      --ci-environment string   appended to the user_agent header
  -c, --config string           config file path
      --group-ref string        group uuid with which to run assessments
      --log-level string        logging level (default "info")
      --minimum-score int       score threshold below which we exit code 1
  -o, --output string           write  output to <file> instead of stdout.
      --output-format string    write  output in specified format. (default "json")
      --poll-for-minutes int    polling max duration (default 60)
      --save-findings           fetch all findings associated with an assessment and write to $PWD/findings.json
      --token string            auth token for REST API
      --ui-host string          UI base url (default "https://app.nowsecure.com")
  -v, --verbose                 enable verbose logging (same as --log-level debug)
```

### SEE ALSO

* [ns run](ns_run.md)	 - Run an assessment for a given application

