package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Suffix     string            `yaml:"suffix"`
	Shortmarks map[string]string `yaml:"shortmarks"`
}

func main() {
	err := mainE()
	if err != nil {
		log.Fatal(err)
	}
}

func mainE() error {
	f := flag.NewFlagSet("shortmarks", flag.ExitOnError)
	configFileFlag := f.String("c", "", "yaml config file")
	portFlag := f.String("p", "", "listening port")
	err := f.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	configFile := os.Getenv("SHORTMARKS_CONFIG")
	if configFileFlag != nil && *configFileFlag != "" {
		configFile = *configFileFlag
	}
	if configFile == "" {
		log.Fatal("config file is required")
	}

	config := &Config{}

	content, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal([]byte(content), config)
	if err != nil {
		return err
	}

	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  30 * time.Second,
		Handler:      http.HandlerFunc(handler(config)),
	}

	port := "8080"
	if v := os.Getenv("PORT"); v != "" {
		port = v
	}
	if portFlag != nil && *portFlag != "" {
		port = *portFlag
	}

	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	log.Fatal(srv.Serve(ln))

	return nil
}

func handler(config *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if config.Suffix == "" || len(config.Shortmarks) == 0 {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		var short string

		hostname := strings.SplitN(req.Host, ":", 2)[0]
		if strings.HasSuffix(hostname, config.Suffix) {
			short = strings.TrimSuffix(hostname, config.Suffix)
		}

		target, ok := config.Shortmarks[short]
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		if req.URL.Path != "/" {
			target += req.URL.Path
		}
		if req.URL.RawQuery != "" {
			target += "?" + req.URL.RawQuery
		}

		w.Header().Set("Connection", "close")
		http.Redirect(w, req, target, http.StatusTemporaryRedirect)
	}
}
