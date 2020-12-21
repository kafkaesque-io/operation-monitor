package cfg

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/kafkaesque-io/pulsar-monitor/src/util"

	"github.com/apex/log"
	"github.com/ghodss/yaml"
)

// PrometheusCfg configures Premetheus set up
type PrometheusCfg struct {
	Port                  string `json:"port"`
	ExposeMetrics         bool   `json:"exposeMetrics"`
	PrometheusProxyURL    string `json:"prometheusProxyURL"`
	PrometheusProxyAPIKey string `json:"prometheusProxyAPIKey"`
}

// SlackCfg is slack configuration
type SlackCfg struct {
	AlertURL string `json:"alertUrl"`
	Verbose  bool   `json:"verbose"`
}

// OpsGenieCfg is opsGenie configuration
type OpsGenieCfg struct {
	HeartBeatURL    string `json:"heartbeatUrl"`
	HeartbeatKey    string `json:"heartbeatKey"`
	AlertKey        string `json:"alertKey"`
	IntervalSeconds int    `json:"intervalSeconds"`
}

// AnalyticsCfg is analytics usage and statistucs tracking configuration
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
	StatusCodeExpr  string            `json:"statusCodeExpr"`
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

// PulsarAdminRESTCfg is for monitor a list of Pulsar cluster
type PulsarAdminRESTCfg struct {
	Token           string          `json:"Token"`
	Clusters        []OpsClusterCfg `json:"clusters"`
	IntervalSeconds int             `json:"intervalSeconds"`
}

// TopicCfg is topic configuration
type TopicCfg struct {
	Name               string         `json:"name"`
	Token              string         `json:"token"`
	TrustStore         string         `json:"trustStore"`
	NumberOfPartitions int            `json:"numberOfPartitions"`
	LatencyBudgetMs    int            `json:"latencyBudgetMs"`
	PulsarURL          string         `json:"pulsarUrl"`
	AdminURL           string         `json:"adminUrl"`
	TopicName          string         `json:"topicName"`
	OutputTopic        string         `json:"outputTopic"`
	IntervalSeconds    int            `json:"intervalSeconds"`
	ExpectedMsg        string         `json:"expectedMsg"`
	PayloadSizes       []string       `json:"payloadSizes"`
	NumOfMessages      int            `json:"numberOfMessages"`
	AlertPolicy        AlertPolicyCfg `json:"AlertPolicy"`
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

// K8sClusterCfg is configuration to monitor kubernete cluster
// only to be enabled in-cluster monitoring
type K8sClusterCfg struct {
	Enabled         bool           `json:"enabled"`
	PulsarNamespace string         `json:"pulsarNamespace"`
	KubeConfigDir   string         `json:"kubeConfigDir"`
	AlertPolicy     AlertPolicyCfg `json:"AlertPolicy"`
}

// BrokersCfg monitors all brokers in the cluster
type BrokersCfg struct {
	InClusterRESTURL string         `json:"inclusterRestURL"`
	IntervalSeconds  int            `json:"intervalSeconds"`
	AlertPolicy      AlertPolicyCfg `json:"AlertPolicy"`
}

// TenantUsageCfg tenant usage reporting and monitoring
type TenantUsageCfg struct {
	OutBytesLimit        uint64 `json:"outBytesLimit"`
	AlertIntervalMinutes int    `json:"alertIntervalMinutes"`
}

// Configuration - this server's configuration
type Configuration struct {
	// Name is the Pulsar cluster name, it is mandatory
	Name string `json:"name"`
	// ClusterName is the Pulsar cluster name if the Name cannot be used as the Pulsar cluster name, optional
	ClusterName string `json:"clusterName"`
	// TokenFilePath is the file path to Pulsar JWT. It takes precedence of the token attribute.
	TokenFilePath string `json:"tokenFilePath"`
	// Token is a Pulsar JWT can be used for both client client or http admin client
	Token             string             `json:"token"`
	BrokersConfig     BrokersCfg         `json:"brokersConfig"`
	TrustStore        string             `json:"trustStore"`
	K8sConfig         K8sClusterCfg      `json:"k8sConfig"`
	AnalyticsConfig   AnalyticsCfg       `json:"analyticsConfig"`
	PrometheusConfig  PrometheusCfg      `json:"prometheusConfig"`
	SlackConfig       SlackCfg           `json:"slackConfig"`
	OpsGenieConfig    OpsGenieCfg        `json:"opsGenieConfig"`
	PulsarAdminConfig PulsarAdminRESTCfg `json:"pulsarAdminRestConfig"`
	PulsarTopicConfig []TopicCfg         `json:"pulsarTopicConfig"`
	SitesConfig       SitesCfg           `json:"sitesConfig"`
	WebSocketConfig   []WsConfig         `json:"webSocketConfig"`
	TenantUsageConfig TenantUsageCfg     `json:"tenantUsageConfig"`
}

// AlertPolicyCfg is a set of criteria to evaluation triggers for incident alert
type AlertPolicyCfg struct {
	// first evaluation to count continuous failure
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
		log.Errorf("failed to load configuration file %s", configFile)
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

	if len(Config.Name) < 1 {
		panic("a valid `name` in Configuration must be specified")
	}

	// reconcile the JWT
	if len(Config.TokenFilePath) > 1 {
		tokenBytes, err := ioutil.ReadFile(Config.TokenFilePath)
		if err != nil {
			log.Errorf("failed to read Pulsar JWT from a file %s", Config.TokenFilePath)
		} else {
			log.Infof("read Pulsar token from the file %s", Config.TokenFilePath)
			Config.Token = string(tokenBytes)
		}
	}
	Config.Token = strings.TrimSuffix(util.AssignString(Config.Token, os.Getenv("PulsarToken")), "\n")

	log.Infof("config %v", Config)
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
		ticker := time.NewTicker(interval)
		fn()
		for {
			select {
			case <-ticker.C:
				fn()
			}
		}

	}()
}
