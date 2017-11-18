# New Relic Infrastructure Integration for couchbase-plugin

Reports status and metrics for couchbase-plugin service

## Requirements

Couchbase REST Endpoint enabled

## Configuration

Edit the nr-couchbase-plugin-config.yml configuration file to provide a unique instance name, host and port of couchbase REST API and the bucket and nodes to monitor. Enter "all" as the value of bucket or node to specify all available buckets and nodes in the cluster.


## Installation

Install the couchbase plugin
```sh

cp -R bin /var/db/newrelic-infra/custom-integrations/

cp nr-couchbase-plugin-definition.yml /var/db/newrelic-infra/custom-integrations/

cp nr-couchbase-plugin-config.yml  /etc/newrelic-infra/integrations.d/

```

Restart the infrastructure agent
```sh
sudo systemctl stop newrelic-infra

sudo systemctl start newrelic-infra
```

## Usage

Test the plugin by executing it from the command line

```sh
Usage of ./bin/nr-couchbase-plugin:
  -host string
    	Hostname or IP of the Couchbase server (default "34.194.55.204")
  -port int
    	Port of the Couchbase server (default 8091)
  -ssl
    	Use SSL connection to Couchbase server
  -username string
    	Username for authenticating to Couchbase server
  -password string
    	Password for authenticating to the Couchbase server
  -bucket string
    	(OPTIONAL) If specified, only the specified bucket stats will be fetched (default "all")
  -node string
    	(OPTIONAL) If specified, only the specified node will be queried (default "all")
  -pretty
    	Print pretty formatted JSON.
  -verbose
    	Print more information to logs.
```

where {host} and {port} refer to the Couchbase Server host and port.

## Compatibility

* Supported OS: Linux 

