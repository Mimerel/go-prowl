package main

import (
	"fmt"
	"github.com/op/go-logging"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)


type configuration struct {
	Token string `yaml:"token,omitempty"`
	Port string `yaml:"port,omitempty"`
	Ignore []Period `yaml:"ignore,omitempty"`
}

type Period struct {
	From int `yaml:"from,omitempty"`
	To int `yaml:"to,omitempty"`
}



var log = logging.MustGetLogger("default")

var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{color:reset} %{message}`,
)


func main() {
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)
	backendLeveled := logging.AddModuleLevel(backend)
	backendLeveled.SetLevel(logging.NOTICE, "")
	logging.SetBackend(backendLeveled, backendFormatter)

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
		fmt.Printf("Configuration Loaded : %+v \n", config)
	}
	return config
}

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
			w.WriteHeader(200)
		}
	} else {
		w.WriteHeader(204)
	}
}

func sendNotification(config *configuration) (bool) {
	hour := time.Now().Hour() * 100
	now := hour + time.Now().Minute()
	for _, moment := range config.Ignore {
		if now >= moment.From && now < moment.To {
			return false
		}
	}
	return true
}