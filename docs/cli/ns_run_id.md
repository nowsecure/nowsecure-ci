## ns run id

Run an assessment for a pre-existing app by specifying app-id

```
ns run id [app-id] [flags]
```

### Examples

```
# Common flags
ns run id [app-id] \
  --group-ref YOUR_GROUP_UUID \
  --analysis-type static \
  --poll-for-minutes 30

# Run an assessment without waiting for results
ns run id [app-id] \
  --group-ref YOUR_GROUP_UUID \
  --poll-for-minutes 0

# Run a full (dynamic and static) assessment
ns run id [app-id] \
  --analysis-type full \
  --group-ref YOUR_GROUP_UUID \
  --poll-for-minutes 60

# Run an assessment with a score threshold
ns run id [app-id] \
  --analysis-type static \
  --minimum-score 70 \
  --poll-for-minutes 60 \
  --group-ref YOUR_GROUP_UUID

```

### Options

```
  -h, --help   help for id
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

