package main

import (
	"bytes"
	"github.com/Mimerel/go-logger-client"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)


type configuration struct {
	Token string `yaml:"token,omitempty"`
	Elasticsearch Elasticsearch `yaml:"elasticSearch,omitempty"`
	Host string `yaml:"host,omitempty"`
	Port string `yaml:"port,omitempty"`
	Ignore []Period `yaml:"ignore,omitempty"`
	Logger logs.LogParams
}

type Period struct {
	From int `yaml:"from,omitempty"`
	To int `yaml:"to,omitempty"`
}

type Elasticsearch struct {
	Url string `yaml:"url,omitempty"`
}


func main() {
	config := readConfiguration()

	Port := config.Port
	http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		urlPath := r.URL.Path
		urlParams := strings.Split(urlPath, "/")
		if len(urlParams) == 4 {
			SendProwlNotification(w, r, urlParams, &config)
		} else {
			w.WriteHeader(500)
		}
	})
	http.ListenAndServe(":" + Port, nil)
}

/**
Reads configuration file
 */
func readConfiguration() (configuration) {
	pathToFile := os.Getenv("LOGGER_CONFIGURATION_FILE")
	if _, err := os.Stat("./configuration.yaml"); !os.IsNotExist(err) {
		pathToFile = "./configuration.yaml"
	} else if pathToFile == "" {
		pathToFile = "/home/pi/go/src/go-prowl/configuration.yaml"
	}
	yamlFile, err := ioutil.ReadFile(pathToFile)

	if err != nil {
		panic(err)
	}

	var config configuration

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		panic(err)
	} else {
		config.Logger = logs.New(config.Elasticsearch.Url, config.Host)
		config.Logger.Info("Configuration Loaded : %+v \n", config)
	}
	return config
}

/**
Sends Prowl notification
 */
func SendProwlNotification(w http.ResponseWriter, r *http.Request, urlParams []string, config *configuration) {
	AppName := urlParams[1]
	Event := urlParams[2]
	Description := urlParams[3]
	timeout := time.Duration(30 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	if sendNotification(config) {
		postingUrl := "https://api.prowlapp.com/publicapi/add?apikey=" + config.Token + "&application=" + AppName + "&event=" + Event + "&description=" + Description + "&priority=1"
		_, err := client.Get(postingUrl)
		if err != nil {
			w.WriteHeader(500)
		} else {
			if config.Elasticsearch.Url != "" {
				err := sendToElasticSearch(config, AppName, Event, Description)
				if err != nil {
					config.Logger.Error("Unable to store prowl event in elasticsearch")
				}
			}
			w.WriteHeader(200)
		}
	} else {
		w.WriteHeader(204)
	}
}

/**
Checks if the notification is in the authorized times
otherwise ignores the notification
 */
func sendNotification(config *configuration) (bool) {
	hour := time.Now().Hour() * 100
	now := hour + time.Now().Minute()
	for _, moment := range config.Ignore {
		if now >= moment.From && now <= moment.To {
			return false
		}
	}
	return true
}

/**
Sends notification to ElasticSearch for storing
 */
func sendToElasticSearch(config *configuration, AppName string, Event string, Description string) (err error) {
	body := createsBodyForElasticSearchCreation(config, AppName, Event, Description)
	timeout := time.Duration(30 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	postingUrl := config.Elasticsearch.Url + "/_bulk"
	config.Logger.Info("Starting to post body")

	resp, err := client.Post(postingUrl, "application/json" , bytes.NewBuffer([]byte(body)))
	if err != nil {
		config.Logger.Error("Failed to post request to elasticSearch %s ", err)
	}
	temp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		config.Logger.Error("Failed to read response from elasticSearch %s ", err)
	}
	config.Logger.Info("response Body %s ", temp)

	resp.Body.Close()
	config.Logger.Info("Event successfully sent to Elasticsearch ")

	return nil
}


func createsBodyForElasticSearchCreation(config *configuration, AppName string, Event string, Description string) (body string) {
	moment := time.Now().Unix()
	timestamp2 := time.Unix(moment, 0).Format(time.RFC3339)
	timestamp := strconv.FormatInt(moment, 10)
	body = body + "{ 'update': { '_id': '" + timestamp  + "_" + config.Host + "', '_type': 'events', '_index': 'prowl' }}\n"
	body = body + "{ 'doc': { "
	body = body + " 'application': '" + AppName + "'"
	body = body + ", 'event': '" +  Event + "'"
	body = body + ", 'description': '" + Description + "'"
	body = body + ", 'value': 1"
	body = body + ", 'timestamp': '" + timestamp + "'"
	body = body + ", 'timestamp2': '" + timestamp2 + "'"
	body = body + "}, 'doc_as_upsert' : true }\n"
	body = strings.Replace(body, "'", "\"", -1)
	return body
}