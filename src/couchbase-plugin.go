package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/buger/jsonparser"
	sdkArgs "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/infra-integrations-sdk/metric"
	"github.com/newrelic/infra-integrations-sdk/sdk"
)

type argumentList struct {
	sdkArgs.DefaultArgumentList
	Host   string `default:"10.142.163.101" help:"Hostname or IP of the Couchbase Server"`
	Port   int    `default:"8091" help:"Port of the Couchbase Server"`
	SSL    bool   `default:"false" help:"Use SSL connection to Couchbase Server"`
	Bucket string `default:"all" help:"(OPTIONAL) If not specified, all buckets will be fetched. If specified, only the specified bucket stats will be fetched"`
	Node   string `default:"all" help:"(OPTIONAL) If specified, only the specified node will be queried"`
}

type metricType int

type metricDef struct {
	metricT metricType
	metricN string
}

type bucketNodeInstance struct {
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

var listBuckets []string

var configuredMetrics = map[string]metricDef{
	"curr_items": metricDef{gauge, ""},
	//RAM usage metrics
	"ep_mem_low_wat":     metricDef{gauge, ""},
	"ep_mem_high_wat":    metricDef{gauge, ""},
	"mem_used":           metricDef{gauge, ""},
	"ep_tmp_oom_errors":  metricDef{gauge, ""},
	"ep_oom_errors":      metricDef{gauge, ""},
	"ep_cache_miss_rate": metricDef{gauge, ""},
	//Disk usage metrics
	"ep_queue_size":   metricDef{gauge, ""},
	"ep_bg_fetched":   metricDef{gauge, ""},
	"ep_io_num_read":  metricDef{gauge, ""},
	"ep_io_num_write": metricDef{gauge, ""},
	//
	"cmd_get":                      metricDef{gauge, ""},
	"ep_kv_size":                   metricDef{gauge, ""},
	"ep_flusher_todo":              metricDef{gauge, ""},
	"ep_dcp_total_queue":           metricDef{gauge, ""},
	"vb_active_num":                metricDef{gauge, ""},
	"vb_replica_num":               metricDef{gauge, ""},
	"vb_active_perc_mem_resident":  metricDef{gauge, ""},
	"vb_replica_perc_mem_resident": metricDef{gauge, ""},
	"ops": metricDef{gauge, ""},
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
	if args.Bucket == "all" {
		// get all bucket names
		bucketsResponse, err2 := httpClient.Get(baseURL + "/pools/default/buckets")
		if err2 != nil {
			return err2
		}
		bucketsData, _ := ioutil.ReadAll(bucketsResponse.Body)
		log.Debug("Reading bucket names")
		listBuckets = getAllBucketNames(bucketsData)
	} else {
		listBuckets = []string{args.Bucket}
	}

	var listBucketNodes []bucketNodeInstance
	if args.Node == "all" {
		for _, bucketName := range listBuckets {
			log.Debug("Reading nodes for bucket: " + bucketName)
			bucketsByNodesResponse, err3 := httpClient.Get(baseURL + "/pools/default/buckets/" + bucketName + "/nodes")
			if err3 != nil {
				return err3
			}
			bucketsByNodesData, _ := ioutil.ReadAll(bucketsByNodesResponse.Body)
			getAllStatsEndpoints(bucketsByNodesData, bucketName, &listBucketNodes)
		}
	} else {
		for _, bucketName := range listBuckets {
			log.Debug("Reading nodes for bucket: " + bucketName)
			bucketnodeURI := fmt.Sprintf("%s%s%s%s%s", "/pools/default/buckets/", bucketName, "/nodes/", args.Node, "/stats")
			listBucketNodes = append(listBucketNodes, bucketNodeInstance{uri: bucketnodeURI, bucket: bucketName, node: args.Node})
		}
	}

	for _, bucketNode := range listBucketNodes {
		log.Debug("Processing metrics at " + bucketNode.uri)
		bucketStatsResponse, err4 := httpClient.Get(baseURL + bucketNode.uri)
		if err4 != nil {
			return err4
		}
		bucketStatsData, _ := ioutil.ReadAll(bucketStatsResponse.Body)
		populateBucketStats(integration, bucketNode.bucket, bucketNode.node, bucketStatsData)
	}
	return nil
}

func getAllBucketNames(data []byte) []string {
	var bucketnames []string
	jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		val, err := jsonparser.GetString(value, "name")
		if err == nil {
			bucketnames = append(bucketnames, val)
		}
	})
	return bucketnames
}

func getAllStatsEndpoints(data []byte, bucketname string, statendpoint *[]bucketNodeInstance) {
	jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			log.Fatal(err)
		}

		hostname, err := jsonparser.GetString(value, "hostname")
		if err != nil {
			log.Error("Unable to parse 'hostname' while getting stats endpoints for bucket[" + bucketname + "]")
		}

		statsuri, err := jsonparser.GetString(value, "stats", "uri")
		if err == nil {
			*statendpoint = append(*statendpoint, bucketNodeInstance{uri: statsuri, bucket: bucketname, node: hostname})
		} else {
			log.Error("Unable to parse 'stats/uri' while getting stats endpoints for bucket[" + bucketname + "]")
		}

	}, "servers")
}

func populateBucketStats(integration *sdk.Integration, bucketName string, hostName string, data []byte) {
	ms := integration.NewMetricSet("CouchbaseSample")
	ms.SetMetric("bucket", bucketName, metric.ATTRIBUTE)
	ms.SetMetric("node", hostName, metric.ATTRIBUTE)

	jsonparser.ObjectEach(data, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		metricName, _ := jsonparser.ParseString(key)
		var metricValue float64
		if metricDef, ok := configuredMetrics[metricName]; ok {
			jsonparser.ArrayEach(value, func(sampleValueBytes []byte, sampleValueType jsonparser.ValueType, sampleOffset int, err error) {
				if err == nil {
					//TODO: Casting type needs to be detemined
					sampleValue, err2 := jsonparser.ParseFloat(sampleValueBytes)
					if err2 == nil {
						metricValue = sampleValue
					}
					if sampleValueType != jsonparser.Number {
						log.Error("Unexpected ValueType for metric: " + metricName)
					}
				} else {
					log.Error("Error iterating sample for metric: ", metricName, err.Error())
				}
			})
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
		return nil
	}, "op", "samples")
}

// Below code is not used
var paths = [][]string{
	[]string{"uptime"},
	[]string{"hostname"},
	[]string{"interestingStats", "curr_items"},
	[]string{"interestingStats", "mem_used"},
	[]string{"interestingStats", "ep_bg_fetched"},
	[]string{"interestingStats", "cmd_get"},
	[]string{"status"},
}

func populateNodeMetrics(integration *sdk.Integration, data []byte) {
	ms := integration.NewMetricSet("CouchbaseSample")
	jsonparser.ArrayEach(data, func(nodeData []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			log.Fatal(err)
		}
		jsonparser.EachKey(nodeData, func(index int, metricData []byte, dataType jsonparser.ValueType, err error) {
			if err != nil {
				log.Fatal(err)
			}
			switch index {
			case 0:
				metricValue, _ := jsonparser.ParseInt(metricData)
				ms.SetMetric("uptime", metricValue, metric.GAUGE)
			case 1:
				metricValue, _ := jsonparser.ParseString(metricData)
				ms.SetMetric("hostname", metricValue, metric.GAUGE)
			case 2:
				metricValue, _ := jsonparser.ParseInt(metricData)
				ms.SetMetric("interestingStats.curr_items", metricValue, metric.GAUGE)
			case 3:
				metricValue, _ := jsonparser.ParseString(metricData)
				ms.SetMetric("interestingStats.mem_used", metricValue, metric.GAUGE)
			}
		}, paths...)

	}, "nodes")
}
