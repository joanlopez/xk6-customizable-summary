# xk6-custosummary

**`xk6-custosummary`** is an [extension for k6](https://k6.io/docs/extensions).
It adds _extraofficial_ support for a "customizable" summary.

## Getting started

Using the `xk6-custosummary` extension involves building a k6 binary incorporating it.
A detailed guide on how to do this using a [Docker](https://www.docker.com/) or [Go](https://go.dev/) environment
is available in the [extension's documentation](https://grafana.com/docs/k6/latest/extensions/build-k6-binary-using-go/).

In the current state, building directly from the source code using Go could be helpful. We list below the suggested steps:

### Prepare the local environment

1. Make sure `git` and `go` are available commands.
2. Install [xk6](https://github.com/grafana/xk6#local-installation) as suggested in 
the [local installation](https://github.com/grafana/xk6#local-installation) documentation's section.
3. Clone the `xk6-custosummary` repository and move inside the project's folder.

### Build the binary

1. Build a k6 binary incorporating the `xk6-custosummary` extension
```bash
xk6 build --with github.com/joanlopez/xk6-custosummary=.
```

2. Run a test script with the newly built binary
```bash
./k6 run script.js
```

## Usage

Once [built](#getting-started) into a k6 executable using [xk6](https://github.com/grafana/xk6),
the extension can be imported by load test scripts as the `k6/x/custosummary` JavaScript module.

## Support

Please, note that this extension is not officially supported by Grafana Labs/k6 core team.