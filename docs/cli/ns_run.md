## ns run

Run an assessment for a given application

### Options

```
      --analysis-type string   One of: full, static, sbom (default "full")
      --artifacts-dir string   directory in which to put artifacts (default "$PWD")
  -h, --help                   help for run
      --minimum-score int      score threshold below which we exit code 1
      --poll-for-minutes int   polling max duration (default 60)
      --save-findings          fetch all findings associated with an assessment and write to $PWD/findings.json
```

### Options inherited from parent commands

```
      --api-host string         REST API base url (default "https://lab-api.nowsecure.com")
      --ci-environment string   appended to the user_agent header
  -c, --config string           config file path
      --group-ref string        group uuid with which to run assessments
      --log-level string        logging level (default "info")
  -o, --output string           write  output to <file> instead of stdout.
      --output-format string    write  output in specified format. (default "json")
      --token string            auth token for REST API
      --ui-host string          UI base url (default "https://app.nowsecure.com")
  -v, --verbose                 enable verbose logging (same as --log-level debug)
```

### SEE ALSO

* [ns](ns.md)	 - NowSecure command line tool to interact with NowSecure Platform
* [ns run file](ns_run_file.md)	 - Upload and run an assessment for a specified binary file
* [ns run id](ns_run_id.md)	 - Run an assessment for a pre-existing app by specifying app-id
* [ns run package](ns_run_package.md)	 - Run an assessment for a pre-existing app by specifying package and platform

