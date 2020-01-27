package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Client holds conenction information
type Client struct {
	Address  string
	Username string
	Password string
	latest   *Result
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

	// do it twice so that we get the XSRF_TOKEN from Set-Cookie
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

var (
	frequency = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "frequency_hz",
		Help: "Frequency in Hertz.",
	}, []string{"direction", "channel"})
	power = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "power",
		Help: "Power dBmV.",
	}, []string{"direction", "channel"})
	downstreamSNR = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ds_snr_db",
		Help: "Downstream Signal to Noise ratio in Decibels.",
	}, []string{"channel"})
)

func main() {

	c := Client{
		Address:  "192.168.100.1",
		Username: "admin",
		Password: "password",
	}
	go func() {
		for {
			time.Sleep(2 * time.Second)

			data, err := c.fetchStatusPage()
			if err != nil {
				log.Fatal(err)
			}
			res, err := parseStatusHTML(data)
			if err != nil {
				log.Fatal(err)
			}
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
		}
	}()

	http.HandleFunc("/", c.dump)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)

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
