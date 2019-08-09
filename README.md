# Sensu Serverspec Event Handler
TravisCI: [![TravisCI Build Status](https://travis-ci.org/nikkixdev/sensu-serverspec-events.svg?branch=master)](https://travis-ci.org/nikkixdev/sensu-serverspec-events)

The Sensu Serverspec Event Handler is a [Sensu Event Handler][3] that parses
[Serverspec][2] JSON output and creates new Sensu events for each test.

## Installation

Download the latest version of the sensu-serverspec-events from [releases][4],
or create an executable script from this source.

From the local path of the sensu-serverspec-events repository:
```
go build -o /usr/local/bin/sensu-serverspec-events main.go
```

## Configuration

Example Sensu Go handler definition:

```yaml
type: Handler
api_version: core/v2
metadata:
  name: serverspec-events
  namespace: default
spec:
  command: sensu-serverspec-events -n serverspec --handlers pagerduty
  env_vars:
  - SENSU_API_TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NjUzMTE4NjYsImp0aSI6ImViYWM5OTk0NTI5MzYyNzIwMjc2NTY3NzU5OGZmZjdkIiwic3ViIjoiYWRtaW4iLCJncm91cHMiOlsiY2x1c3Rlci1hZG1pbnMiLCJzeXN0ZW06dXNlcnMiXSwicHJvdmlkZXIiOnsicHJvdmlkZXJfaWQiOiJiYXNpYyIsInVzZXJfaWQiOiJhZG1pbiJ9fQ.VE9c6CGYZTRR9e6fyez75n8EHgHn94z_Sk-h6iqZ8jQ
  type: pipe
```

Example Sensu Go check definition:

```yaml
type: CheckConfig
api_version: core/v2
metadata:
  name: serverspec-run
  namespace: default
spec:
  command: serverspec-run.sh SPEC_OPTS="--format json" | tail -n +3
  handlers:
  - serverspec-events
  interval: 60
  publish: true
  subscriptions:
  - test
```


## Usage Examples

Help:
```
a serverspec json handler to use with sensu

Usage:
  sensu-serverspec-events [flags]

Flags:
      --handlers strings   sensu handlers that the new serverspec events will be handled by
  -h, --help               help for sensu-serverspec-events
  -n, --namespace string   sensu namespace that the new serverspec events will be created in (default "default")
  -t, --token string       sensu api token, (default is the value of the SENSU_API_TOKEN environment variable)
  -u, --url string         sensu api url (default "http://127.0.0.1:8080")
```

[1]: https://github.com/sensu/sensu-go
[2]: https://serverspec.org/
[3]: https://docs.sensu.io/sensu-go/latest/reference/handlers/
[4]: https://github.com/nikkixdev/sensu-serverspec-events/releases