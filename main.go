package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Client holds conenction information
type Client struct {
	Address  string
	Username string
	Password string
	latest   *Result
	Interval time.Duration
}

func (c *Client) fetchStatusPage() (string, error) {
	cookieJar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: cookieJar,
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/DocsisStatus.htm", c.Address), nil)

	if err != nil {
		fmt.Println(err)
	}
	req.SetBasicAuth(c.Username, c.Password)

	// do it twice so that we get the XSRF_TOKEN from Set-Cookie..
	// TOOD: fix this
	client.Do(req)
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request failed, got %d (%v), %v", res.StatusCode, string(body), res.Header)
	}

	return string(body), nil
}

func main() {

	viper.SetDefault("address", "192.168.100.1")
	viper.SetDefault("username", "admin")
	viper.SetDefault("password", "password")
	viper.SetDefault("port", 4023)
	viper.SetDefault("interval", "30s")
	viper.AutomaticEnv()

	port := viper.GetInt("port")

	c := Client{
		Address:  viper.GetString("address"),
		Username: viper.GetString("username"),
		Password: viper.GetString("password"),
		Interval: viper.GetDuration("interval"),
	}
	go func() {
		for {
			data, err := c.fetchStatusPage()
			if err != nil {
				log.Printf("failed to fetch status page: %v", err)
				continue
			}
			res, err := parseStatusHTML(data)
			if err != nil {
				log.Printf("failed to parse status page: %v", err)
				continue
			}
			log.Println("fetched")
			c.latest = res
			for _, d := range res.Downstream {
				frequency.WithLabelValues("ds", strconv.FormatUint(d.Channel, 10)).Set(float64(d.FrequencyHz))
				power.WithLabelValues("ds", strconv.FormatUint(d.Channel, 10)).Set(float64(d.Power))
				downstreamSNR.WithLabelValues(strconv.FormatUint(d.Channel, 10)).Set(float64(d.SNR))
			}
			for _, u := range res.Upstream {
				frequency.WithLabelValues("us", strconv.FormatUint(u.Channel, 10)).Set(float64(u.FrequencyHz))
				power.WithLabelValues("us", strconv.FormatUint(u.Channel, 10)).Set(float64(u.Power))
			}
			time.Sleep(c.Interval)
		}
	}()

	http.HandleFunc("/", c.dump)
	http.Handle("/metrics", promhttp.Handler())
	bind := fmt.Sprintf(":%d", port)
	log.Printf("listening on %v, parsing %s every %s", bind, c.Address, c.Interval)
	http.ListenAndServe(bind, nil)

}
func (c *Client) dump(w http.ResponseWriter, r *http.Request) {

	js, err := json.Marshal(c.latest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
