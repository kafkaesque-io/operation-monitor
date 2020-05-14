package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"time"
	"unicode"

	"github.com/ghodss/yaml"
)

// PrometheusCfg configures Premetheus set up
type PrometheusCfg struct {
	Port          string `json:"port"`
	ExposeMetrics bool   `json:"exposeMetrics"`
}

// SlackCfg is slack configuration
type SlackCfg struct {
	AlertURL string `json:"alertUrl"`
}

// OpsGenieCfg is opsGenie configuration
type OpsGenieCfg struct {
	HeartBeatURL    string `json:"heartbeatUrl"`
	HeartbeatKey    string `json:"heartbeatKey"`
	AlertKey        string `json:"alertKey"`
	IntervalSeconds int    `json:"intervalSeconds"`
}

// AnalyticsCfg is analytics usage and statistucs tracking configuraiton
type AnalyticsCfg struct {
	APIKey            string `json:"apiKey"`
	IngestionURL      string `json:"ingestionUrl"`
	InsightsWriteKey  string `json:"insightsWriteKey"`
	InsightsAccountID string `json:"insightsAccountId"`
}

// SiteCfg configures general website
type SiteCfg struct {
	Headers         map[string]string `json:"headers"`
	URL             string            `json:"url"`
	Name            string            `json:"name"`
	IntervalSeconds int               `json:"intervalSeconds"`
	ResponseSeconds int               `json:"responseSeconds"`
	StatusCode      int               `json:"statusCode"`
	Retries         int               `json:"retries"`
	AlertPolicy     AlertPolicyCfg    `json:"alertPolicy"`
}

// SitesCfg configures a list of website`
type SitesCfg struct {
	Sites []SiteCfg `json:"sites"`
}

// OpsClusterCfg is each cluster's configuration
type OpsClusterCfg struct {
	Name        string         `json:"name"`
	URL         string         `json:"url"`
	AlertPolicy AlertPolicyCfg `json:"alertPolicy"`
}

// PulsarOpsCfg is for monitor a list of Pulsar cluster
type PulsarOpsCfg struct {
	MasterToken     string          `json:"masterToken"`
	Clusters        []OpsClusterCfg `json:"clusters"`
	IntervalSeconds int             `json:"intervalSeconds"`
}

// TopicCfg is topic configuration
type TopicCfg struct {
	Name            string         `json:"name"`
	Token           string         `json:"token"`
	LatencyBudgetMs int            `json:"latencyBudgetMs"`
	PulsarURL       string         `json:"pulsarUrl"`
	TopicName       string         `json:"topicName"`
	OutputTopic     string         `json:"outputTopic"`
	IntervalSeconds int            `json:"intervalSeconds"`
	ExpectedMsg     string         `json:"expectedMsg"`
	PayloadSizes    []string       `json:"payloadSizes"`
	NumOfMessages   int            `json:"numberOfMessages"`
	AlertPolicy     AlertPolicyCfg `json:"AlertPolicy"`
}

// WsConfig is configuration to monitor WebSocket pub sub latency
type WsConfig struct {
	Name            string         `json:"name"`
	Token           string         `json:"token"`
	Cluster         string         `json:"cluster"` // can be used for alert de-dupe
	LatencyBudgetMs int            `json:"latencyBudgetMs"`
	ProducerURL     string         `json:"producerUrl"`
	ConsumerURL     string         `json:"consumerUrl"`
	TopicName       string         `json:"topicName"`
	IntervalSeconds int            `json:"intervalSeconds"`
	Scheme          string         `json:"scheme"`
	Port            string         `json:"port"`
	Subscription    string         `json:"subscription"`
	URLQueryParams  string         `json:"urlQueryParams"`
	AlertPolicy     AlertPolicyCfg `json:"AlertPolicy"`
}

// PulsarPerfCfg is configuration to monitor Pulsar pub sub latency
type PulsarPerfCfg struct {
	Token           string     `json:"token"`
	TrustStore      string     `json:"trustStore"`
	IntervalSeconds int        `json:"intervalSeconds"`
	TopicCfgs       []TopicCfg `json:"topicCfgs"`
}

// PulsarFuncCfg specifies the configuration of each function trigger
type PulsarFuncCfg struct {
	Name        string         `json:"name"` // save me some work not to parse the url
	TriggerURL  string         `json:"triggerUrl"`
	AlertPolicy AlertPolicyCfg `json:"alertPolicy"`
}

// PulsarFunctionsCfg configures a list of trigger functions in each cluster
type PulsarFunctionsCfg struct {
	MasterToken     string          `json:"masterToken"` //we can use the PulsarOpsConfig's master token
	Clusters        []PulsarFuncCfg `json:"clusters"`
	IntervalSeconds int             `json:"intervalSeconds"`
}

// Configuration - this server's configuration
type Configuration struct {
	Name                  string             `json:"name"`
	AnalyticsConfig       AnalyticsCfg       `json:"analyticsConfig"`
	PrometheusConfig      PrometheusCfg      `json:"prometheusConfig"`
	SlackConfig           SlackCfg           `json:"slackConfig"`
	OpsGenieConfig        OpsGenieCfg        `json:"opsGenieConfig"`
	PulsarOpsConfig       PulsarOpsCfg       `json:"pulsarOpsConfig"`
	PulsarPerfConfig      PulsarPerfCfg      `json:"pulsarPerfConfig"`
	PulsarFunctionsConfig PulsarFunctionsCfg `json:"pulsarFunctionsConfig"`
	PulsarTopicConfig     []TopicCfg         `json:"pulsarTopicConfig"`
	SitesConfig           SitesCfg           `json:"sitesConfig"`
	WebSocketConfig       []WsConfig         `json:"webSocketConfig"`
}

// AlertPolicyCfg is a set of criteria to evaluation triggers for incident alert
type AlertPolicyCfg struct {
	// first evalation for a single count
	Ceiling int `json:"ceiling"`
	// Second evaluation for moving window
	MovingWindowSeconds   int `json:"movingWindowSeconds"`
	CeilingInMovingWindow int `json:"ceilingInMovingWindow"`
}

// Config - this server's configuration instance
var Config Configuration

// ReadConfigFile reads configuration file.
func ReadConfigFile(configFile string) {

	fileBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Printf("failed to load configuration file %s", configFile)
		panic(err)
	}

	if hasJSONPrefix(fileBytes) {
		err = json.Unmarshal(fileBytes, &Config)
		if err != nil {
			panic(err)
		}
	} else {
		err = yaml.Unmarshal(fileBytes, &Config)
		if err != nil {
			panic(err)
		}
	}

	log.Println(Config)
}

var jsonPrefix = []byte("{")

func hasJSONPrefix(buf []byte) bool {
	return hasPrefix(buf, jsonPrefix)
}

// Return true if the first non-whitespace bytes in buf is prefix.
func hasPrefix(buf []byte, prefix []byte) bool {
	trim := bytes.TrimLeftFunc(buf, unicode.IsSpace)
	return bytes.HasPrefix(trim, prefix)
}

//GetConfig returns a reference to the Configuration
func GetConfig() *Configuration {
	return &Config
}

//
type monitorFunc func()

// RunInterval runs interval
func RunInterval(fn monitorFunc, interval time.Duration) {
	go func() {
		fn()
		for {
			select {
			case <-time.Tick(interval):
				fn()
			}
		}

	}()
}
