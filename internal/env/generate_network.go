package env

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/proxy"
)

const (
	connectTimeout = 2 * time.Second
	httpTimeout    = 10 * time.Second

	userAgent = "chore"
)

type ipInfoResponse struct {
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Loc      string `json:"loc"`
	Org      string `json:"org"`
	Postal   string `json:"postal"`
	Timezone string `json:"timezone"`
}

type ifConfigResponse struct {
	IP string `json:"ip"`
}

var (
	netDialer = proxy.FromEnvironmentUsing(&net.Dialer{
		Timeout:       connectTimeout,
		FallbackDelay: -1,
		KeepAlive:     -1,
	}).(proxy.ContextDialer)
	ipInfoOrgFormat = regexp.MustCompile(`^AS(\d+)\s+(.*?)$`)

	HTTPClientV4 = &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
			DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
				return netDialer.DialContext(ctx, "tcp4", address)
			},
		},
		Timeout: httpTimeout,
	}
	HTTPClientV6 = &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
			DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
				return netDialer.DialContext(ctx, "tcp6", address)
			},
		},
		Timeout: httpTimeout,
	}
)

func doRequest(ctx context.Context, client *http.Client, url string, target interface{}) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Set("User-Agent", "chore")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("cannot access endpoint: %w", err)
	}

	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("unexpected response status code %d", resp.StatusCode)
	}

	reader := bufio.NewReader(resp.Body)
	decoder := json.NewDecoder(reader)

	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("cannot parse response: %w", err)
	}

	return nil
}

func GenerateNetwork(ctx context.Context, results chan<- string, wg *sync.WaitGroup) {
	wg.Add(1)

	go func() {
		defer wg.Done()

		if _, ok := os.LookupEnv(EnvNetworkIPv4); ok {
			return
		}

		resp := ipInfoResponse{}
		if err := doRequest(ctx, HTTPClientV4, "https://ipinfo.io/json", &resp); err != nil {
			log.Printf("cannot request network data: %v", err)

			return
		}

		if resp.IP != "" {
			sendValue(ctx, results, EnvNetworkIPv4, resp.IP)
		}

		if resp.Hostname != "" {
			sendValue(ctx, results, EnvNetworkHostname, resp.Hostname)
		}

		if resp.City != "" {
			sendValue(ctx, results, EnvNetworkCity, resp.City)
		}

		if resp.Region != "" {
			sendValue(ctx, results, EnvNetworkRegion, resp.Region)
		}

		if resp.Country != "" {
			sendValue(ctx, results, EnvNetworkCountry, resp.Country)
		}

		asnChunks := ipInfoOrgFormat.FindStringSubmatch(resp.Org)
		switch {
		case asnChunks == nil && resp.Org != "":
			sendValue(ctx, results, EnvNetworkOrganization, resp.Org)
		case asnChunks != nil:
			sendValue(ctx, results, EnvNetworkASN, asnChunks[1])
			sendValue(ctx, results, EnvNetworkOrganization, asnChunks[2])
		}

		if resp.Postal != "" {
			sendValue(ctx, results, EnvNetworkPostal, resp.Postal)
		}

		if resp.Timezone != "" {
			sendValue(ctx, results, EnvNetworkTimezone, resp.Timezone)
		}

		if lat, lon, ok := strings.Cut(resp.Loc, ","); ok {
			sendValue(ctx, results, EnvNetworkLatitude, lat)
			sendValue(ctx, results, EnvNetworkLongitude, lon)
		}
	}()
}

func GenerateNetworkIPv6(ctx context.Context, results chan<- string, wg *sync.WaitGroup) {
	wg.Add(1)

	go func() {
		defer wg.Done()

		if _, ok := os.LookupEnv(EnvNetworkIPv6); ok {
			return
		}

		resp := ifConfigResponse{}
		if err := doRequest(ctx, HTTPClientV6, "https://ifconfig.co", &resp); err != nil {
			log.Printf("cannot get IPv6 address: %v", err)

			return
		}

		sendValue(ctx, results, EnvNetworkIPv6, resp.IP)
	}()
}
