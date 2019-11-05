# Backup Tool
This tool is meant to facilitate the backing up and archiving of various types
of production data. Currently the offering of data sources and storage services
is fairly limited, but the plan is to make more available.

## Building
```
go build -o backup-tool
```

## Running
Depending on the input/output plugins you're using, you may have to specify some
environment variables.
```
export AZURE_STORAGE_ACCOUNT=mystorageaccount
export AZURE_STORAGE_CONTAINER=backup
export AZURE_STORAGE_KEY=thisshouldntbeinyourcommandhistory

export GRAFANA_URL=https://grafana.mydomain.com
export GRAFANA_TOKEN=thisalsoshouldntbeinyourcommandhistory

./backup-tool -storage.type=azure -datasource=grafana -output.prefix=grafana
```
