package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"

	"github.com/leviathan1995/spleen/client/util"
)

func main() {
	var conf string
	var config map[string]interface{}
	flag.StringVar(&conf, "c", ".client.json", "The client configuration.")
	flag.Parse()

	bytes, err := ioutil.ReadFile(conf)
	if err != nil {
		log.Fatalf("Reading %s failed.", conf)
	}

	if err := json.Unmarshal(bytes, &config); err != nil {
		log.Fatalf("Parsing %s failed.", conf)
	}

	serverIP := config["server_ip"].(string)
	serverPort := int(config["server_port"].(float64))
	c := client.NewClient(serverIP, serverPort)
	c.Run()
}
