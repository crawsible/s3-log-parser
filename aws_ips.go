package main

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type AWSPublicIPs struct {
	PublicIPs []struct {
		CIDR string `json:"ip_prefix"`
	} `json:"prefixes"`
}

var reCIDR *regexp.Regexp = regexp.MustCompile(`(\d+\.\d+\.\d+\.\d+)/(\d+)`)

func GetAWSPublicIPRanges() ([][2]uint32, error) {
	IPRanges := [][2]uint32{}
	r, err := http.Get("https://ip-ranges.amazonaws.com/ip-ranges.json")
	if err != nil {
		return IPRanges, err
	}
	defer r.Body.Close()

	resp := AWSPublicIPs{}
	err = json.NewDecoder(r.Body).Decode(&resp)
	if err != nil {
		return IPRanges, err
	}

	for _, IP := range resp.PublicIPs {
		IPRange, err := getRangeForCIDR(IP.CIDR)
		if err != nil {
			return IPRanges, err
		}

		IPRanges = append(IPRanges, IPRange)
	}

	return IPRanges, nil
}

type InvalidCIDRError struct{}

func (e InvalidCIDRError) Error() string {
	return "InvalidCIDRError: CIDR is invalid"
}

func getRangeForCIDR(CIDR string) ([2]uint32, error) {
	IPRange := [2]uint32{}
	matches := reCIDR.FindStringSubmatch(CIDR)
	if len(matches) < 3 {
		return IPRange, InvalidCIDRError{}
	}

	var err error
	IPRange[0], err = getIPValue(matches[1])
	if err != nil {
		return IPRange, err
	}

	var CIDRLength int
	CIDRLength, err = strconv.Atoi(matches[2])
	if err != nil {
		return IPRange, err
	}

	IPRange[1] = IPRange[0] + (1 << uint(32-CIDRLength))
	return IPRange, nil
}

func getIPValue(IP string) (uint32, error) {
	IPValue := uint32(0)
	for i, IPByte := range strings.Split(IP, ".") {
		byteVal, err := strconv.Atoi(IPByte)
		if err != nil {
			return IPValue, err
		}

		IPValue += (uint32(byteVal) * (1 << uint32(8*(3-i))))
	}

	return IPValue, nil
}
