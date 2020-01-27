package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

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
