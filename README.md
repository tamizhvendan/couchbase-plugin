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
./bin/nr-couchbase-plugin -hostname {host} -port {port} -bucket all -node all
```

where {host} and {port} refer to the Couchbase Server host and port.

## Compatibility

* Supported OS: Linux 

