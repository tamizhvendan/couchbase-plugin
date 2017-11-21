package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	sdkArgs "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/infra-integrations-sdk/metric"
	"github.com/newrelic/infra-integrations-sdk/sdk"
)

type argumentList struct {
	sdkArgs.DefaultArgumentList
	Host     string `default:"localhost" help:"Hostname or IP of the Couchbase server"`
	Port     int    `default:"8091" help:"Port of the Couchbase server"`
	Username string `default:"" help:"Username for authenticating to Couchbase server"`
	Password string `default:"" help:"Password for authenticating to the Couchbase server"`
	SSL      bool   `default:"false" help:"Use SSL connection to Couchbase server"`
	Bucket   string `default:"all" help:"(OPTIONAL) If specified, only the specified bucket stats will be fetched"`
	Node     string `default:"all" help:"(OPTIONAL) If specified, only the specified node will be queried"`
}

type metricType int

type metricDef struct {
	metricT metricType
	metricN string
}

type statsEndpoint struct {
	uri    string
	bucket string
	node   string
}

const (
	integrationName               = "com.newrelic.couchbase-plugin"
	integrationVersion            = "0.1.0"
	gauge              metricType = iota
	delta
	rate
	attribute
)

var args argumentList

var transport = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}
var httpClient = &http.Client{Transport: transport, Timeout: 10 * time.Second}

var baseURL string

var username string

var password string

var listBuckets []string

var configuredMetrics = map[string]metricDef{
	"cmd_get":                         metricDef{gauge, ""},
	"cmd_set":                         metricDef{gauge, ""},
	"delete_hits":                     metricDef{gauge, ""},
	"ep_cache_miss_rate":              metricDef{gauge, ""},
	"couch_docs_fragmentation":        metricDef{gauge, ""},
	"couch_views_fragmentation":       metricDef{gauge, ""},
	"curr_connections":                metricDef{gauge, ""},
	"ep_dcp_replica_items_remaining":  metricDef{gauge, ""},
	"ep_dcp_2i_items_remaining":       metricDef{gauge, ""},
	"ep_dcp_views_items_remaining":    metricDef{gauge, ""},
	"ep_dcp_replica_backoff":          metricDef{gauge, ""},
	"ep_dcp_xdcr_backoff":             metricDef{gauge, ""},
	"vb_avg_total_queue_age":          metricDef{gauge, ""},
	"ep_oom_errors":                   metricDef{gauge, ""},
	"ep_tmp_oom_errors":               metricDef{gauge, ""},
	"vb_active_resident_items_ratio":  metricDef{gauge, ""},
	"vb_replica_resident_items_ratio": metricDef{gauge, ""},
	//percent_quota_utilization: mem_used / ep_mem_high_wat
	"mem_used":        metricDef{gauge, ""},
	"ep_mem_high_wat": metricDef{gauge, ""},
	//percent_metadata_utilization: ep_meta_data_memory / ep_mem_high_wat
	"ep_meta_data_memory": metricDef{gauge, ""},
	//disk_write_queue: ep_queue_size + ep_flusher_todo
	"ep_queue_size":   metricDef{gauge, ""},
	"ep_flusher_todo": metricDef{gauge, ""},
	//total_ops: cmd_get + cmd_set + incr_misses + incr_hits + decr_misses + decr_hits + delete_misses + delete_hits
	"incr_misses":   metricDef{gauge, ""},
	"incr_hits":     metricDef{gauge, ""},
	"decr_misses":   metricDef{gauge, ""},
	"decr_hits":     metricDef{gauge, ""},
	"delete_misses": metricDef{gauge, ""},
}

func main() {
	integration, err := sdk.NewIntegration(integrationName, integrationVersion, &args)
	fatalIfErr(err)

	if args.All || args.Inventory {
		fatalIfErr(populateInventory(integration.Inventory))
	}

	if args.All || args.Metrics {
		fatalIfErr(populateMetrics(integration))
	}
	fatalIfErr(integration.Publish())
}

func fatalIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func populateInventory(inventory sdk.Inventory) error {
	// Insert here the logic of your integration to get the inventory data
	// Ex: inventory.SetItem("softwareVersion", "value", "1.0.1")
	// --
	return nil
}

func populateMetrics(integration *sdk.Integration) error {
	protocol := "http://"
	if args.SSL {
		protocol = "https://"
	}
	baseURL = fmt.Sprintf("%s%s%s%d", protocol, args.Host, ":", args.Port)
	username = strings.TrimSpace(args.Username)
	password = strings.TrimSpace(args.Password)

	bucketArg := strings.TrimSpace(args.Bucket)
	if bucketArg == "all" {
		// get all bucket names
		req, err := http.NewRequest("GET", baseURL+"/pools/default/buckets", nil)
		if err != nil {
			return err
		}
		req.SetBasicAuth(username, password)
		bucketsResponse, err := httpClient.Do(req)
		if err != nil {
			return err
		}
		bucketsData, err := ioutil.ReadAll(bucketsResponse.Body)
		defer bucketsResponse.Body.Close()
		if err != nil {
			return err
		}

		log.Debug("Reading bucket names" + string(bucketsData))
		listBuckets = getAllBucketNames(bucketsData)
	} else {
		listBuckets = []string{bucketArg}
	}

	var statEndpoints []statsEndpoint
	nodeArg := strings.TrimSpace(args.Node)
	if nodeArg == "all" {
		for _, bucketName := range listBuckets {
			log.Debug("Reading nodes for bucket: " + bucketName)
			req, err := http.NewRequest("GET", baseURL+"/pools/default/buckets/"+bucketName+"/nodes", nil)
			if err != nil {
				return err
			}
			req.SetBasicAuth(username, password)

			bucketsByNodesResponse, err := httpClient.Do(req)
			if err != nil {
				return err
			}
			bucketsByNodesData, _ := ioutil.ReadAll(bucketsByNodesResponse.Body)
			defer bucketsByNodesResponse.Body.Close()
			getAllStatsEndpoints(bucketsByNodesData, bucketName, &statEndpoints)
		}
	} else {
		for _, bucketName := range listBuckets {
			log.Debug("Reading nodes for bucket: " + bucketName)
			statsURI := fmt.Sprintf("%s%s%s%s%s", "/pools/default/buckets/", bucketName, "/nodes/", nodeArg, "/stats")
			statEndpoints = append(statEndpoints, statsEndpoint{uri: statsURI, bucket: bucketName, node: nodeArg})
		}
	}

	for _, ep := range statEndpoints {
		log.Debug("Processing metrics at " + ep.uri)
		req, err := http.NewRequest("GET", baseURL+ep.uri, nil)
		if err != nil {
			return err
		}
		req.SetBasicAuth(username, password)

		statsResponse, err := httpClient.Do(req)
		if err != nil {
			return err
		}
		statsData, _ := ioutil.ReadAll(statsResponse.Body)
		defer statsResponse.Body.Close()
		populateStats(integration, ep.bucket, ep.node, statsData)
	}
	return nil
}

func getAllBucketNames(data []byte) []string {
	var bucketnames []string
	config := Config{
		Properties: []Property{
			{Path: "./name", Type: "[]op"},
		},
	}
	err := PickDeserializedUsingConfig(bytes.NewReader(data), config, "name", &bucketnames)
	if err != nil {
		log.Fatal(err)
		return []string{}
	}
	return bucketnames
}

func getAllStatsEndpoints(data []byte, bucketname string, ep *[]statsEndpoint) {

	statsUriAlias := "statsUris"
	hostnameAlias := "hostNames"
	config := Config{
		Properties: []Property{
			{Path: "servers/hostname", Type: "[]o", Alias: &hostnameAlias},
			{Path: "servers/stats/uri", Type: "[]op", Alias: &statsUriAlias},
		},
	}
	var result struct {
		Hostnames []string
		StatsUris []string
	}

	err := PickDeserializedUsingConfig(bytes.NewReader(data), config, "", &result)
	if err != nil {
		log.Fatal(err)
	}
	for i, _ := range result.Hostnames {
		*ep = append(*ep, statsEndpoint{uri: result.StatsUris[i], bucket: bucketname, node: result.Hostnames[i]})
	}
}

func populateStats(integration *sdk.Integration, bucketName string, hostName string, statsData []byte) {
	ms := integration.NewMetricSet("CouchbaseSample")
	ms.SetMetric("bucket", bucketName, metric.ATTRIBUTE)
	ms.SetMetric("node", hostName, metric.ATTRIBUTE)

	for metricName, metricDef := range configuredMetrics {
		var sumMetricSamples float64
		var countMetricSamples float64
		metrics := []float64{}
		metricPath := fmt.Sprintf("op/samples/%s", metricName)
		config := Config{
			Properties: []Property{
				{Path: metricPath, Type: "[f]"},
			},
		}
		err := PickDeserializedUsingConfig(bytes.NewReader(statsData), config, metricName, &metrics)
		if err != nil {
			log.Fatal(err)
		}
		for _, m := range metrics {
			sumMetricSamples = sumMetricSamples + m
			countMetricSamples++
		}
		metricValue := sumMetricSamples / countMetricSamples
		switch metricDef.metricT {
		case gauge:
			ms.SetMetric(metricName, metricValue, metric.GAUGE)
		case delta:
			ms.SetMetric(metricName, metricValue, metric.DELTA)
		case rate:
			ms.SetMetric(metricName, metricValue, metric.RATE)
		case attribute:
			ms.SetMetric(metricName, metricValue, metric.ATTRIBUTE)
		}
	}
}
