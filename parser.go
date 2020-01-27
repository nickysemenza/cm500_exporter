package main

import (
	"fmt"
	"strconv"
	"strings"
)

type Result struct {
	Upstream          []Upstream
	Downstream        []Downstream
	ProvisionedDSRate uint64
	ProvisionedUSRate uint64
	Init              Init
}
type Init struct {
	AcquireDsChanelStatusHz  string
	AcquireDsChanelComment   string
	ConnectivityStateStatus  string
	ConnectivityStateComment string
	BootStateStatus          string
	BootStateComment         string
	ConfigurationFileStatus  string
	ConfigurationFileComment string
	SecurityStatus           string
	SecurityComment          string
	CurrentSystemTime        string
	StartupFreq              string
	DownstreamBondedChannels string
	UpstreamBondedChannels   string
	SystemUpTime             string
	ExtUsPower               string
}
type Upstream struct {
	Channel     uint64
	Lock        string
	Type        string
	ChannelID   uint64
	SymbolRate  uint64 // Ksym/sec
	FrequencyHz uint64
	Power       float64
}
type Downstream struct {
	Channel        uint64
	Lock           string
	Modulation     string
	ChannelID      uint64
	FrequencyHz    uint64
	Power          float64
	SNR            float64
	Correctables   uint64
	UnCorrectables uint64
}

// parseStatusHTML parses the HTML status page.
// There are 5 distinct valuesets:
/*
1. Initial
  Acquire Downstream Channel (text) | Acquire Downstream Channel Comment (text) |
  Connectivity State (text) | Connectivity State Comment (text) |
  Boot State (text) | Boot State Comment (text) |
  Configuration File (text) | Configuration File Comment (text) |
  Security (text) | Security Comment (text) |
  Current System Time (text)
2. []Upstream
Channel (text) | Lock Status (text) | US Channel Type (text) | Channel ID (text) | Symbol Rate (text) | Frequency (text) | Power (text)

3. []Downstream
Channel (text) | Lock Status (text) | Modulation (text) | Channel ID (text) | Frequency (text) | Power (text) | SNR (text) | Correctables (text) | Uncorrectables (text)

4. Provisioning Rate
Is Genie (text) | DS Provisioned Rate (text) | US Provisioned Rate (text)

5. IP Provisioning Mode
Is Retail (bool) | IP Provisioning Mode (text) | MIB Value (text)
*/
func parseStatusHTML(html string) (*Result, error) {

	prefix := "    var tagValueList = '"
	split := strings.Split(strings.Replace(html, "\r\n", "\n", -1), "\n")
	var tagValueLists [][]string
	for _, line := range split {
		if strings.HasPrefix(line, prefix) {
			trimmed := strings.TrimLeft(strings.TrimRight(line, "|';"), prefix)
			split := strings.Split(trimmed, "|")
			tagValueLists = append(tagValueLists, split)
		}
	}
	if len(tagValueLists) != 5 {
		return nil, fmt.Errorf("expected len 5, got %d", len(tagValueLists))
	}

	var err error
	res := Result{}

	// parse upstream
	upstreamRows := splitSlice(tagValueLists[1][1:], 7)
	for _, r := range upstreamRows {
		u := Upstream{}
		u.Channel, err = strconv.ParseUint(r[0], 10, 64)
		if err != nil {
			return nil, err
		}
		u.Lock = r[1]
		u.Type = r[2]
		u.ChannelID, err = strconv.ParseUint(r[3], 10, 64)
		if err != nil {
			return nil, err
		}
		u.SymbolRate, err = strconv.ParseUint(r[4], 10, 64)
		if err != nil {
			return nil, err
		}
		u.FrequencyHz, err = strconv.ParseUint(strings.TrimRight(r[5], " Hz"), 10, 64)
		if err != nil {
			return nil, err
		}
		u.Power, err = strconv.ParseFloat(r[4], 64)
		if err != nil {
			return nil, err
		}

		res.Upstream = append(res.Upstream, u)
	}
	// parse INITIAL

	initValues := tagValueLists[0]
	res.Init.AcquireDsChanelStatusHz = initValues[0]
	res.Init.AcquireDsChanelComment = initValues[1]
	res.Init.ConnectivityStateStatus = initValues[2]
	res.Init.ConnectivityStateComment = initValues[3]
	res.Init.BootStateStatus = initValues[4]
	res.Init.BootStateComment = initValues[5]
	res.Init.ConfigurationFileStatus = initValues[6]
	res.Init.ConfigurationFileComment = initValues[7]
	res.Init.SecurityStatus = initValues[8]
	res.Init.SecurityComment = initValues[9]
	res.Init.CurrentSystemTime = initValues[10]
	res.Init.StartupFreq = initValues[11]
	res.Init.DownstreamBondedChannels = initValues[12]
	res.Init.UpstreamBondedChannels = initValues[13]
	res.Init.SystemUpTime = initValues[14]
	res.Init.ExtUsPower = initValues[15]

	// parse DOWNSTREAM
	downstreamRows := splitSlice(tagValueLists[2][1:], 9)
	for _, r := range downstreamRows {
		d := Downstream{}
		d.Channel, err = strconv.ParseUint(r[0], 10, 64)
		if err != nil {
			return nil, err
		}
		d.Lock = r[1]
		d.Modulation = r[2]
		d.ChannelID, err = strconv.ParseUint(r[3], 10, 64)
		if err != nil {
			return nil, err
		}

		d.FrequencyHz, err = strconv.ParseUint(strings.TrimRight(r[4], " Hz"), 10, 64)
		if err != nil {
			return nil, err
		}
		d.Power, err = strconv.ParseFloat(r[5], 64)
		if err != nil {
			return nil, err
		}
		d.SNR, err = strconv.ParseFloat(r[6], 64)
		if err != nil {
			return nil, err
		}
		d.Correctables, err = strconv.ParseUint(r[7], 10, 64)
		if err != nil {
			return nil, err
		}
		d.UnCorrectables, err = strconv.ParseUint(r[8], 10, 64)
		if err != nil {
			return nil, err
		}

		res.Downstream = append(res.Downstream, d)
	}
	// provisoned rate

	res.ProvisionedDSRate, err = strconv.ParseUint(tagValueLists[3][1], 10, 64)
	if err != nil {
		return nil, err
	}
	res.ProvisionedUSRate, err = strconv.ParseUint(tagValueLists[3][2], 10, 64)
	if err != nil {
		return nil, err
	}

	return &res, nil

}

func splitSlice(slice []string, batchSize int) [][]string {
	var result [][]string
	for i := 0; i < len(slice); i += batchSize {
		j := i + batchSize
		if j > len(slice) {
			j = len(slice)
		}

		result = append(result, slice[i:j])
	}
	return result
}
