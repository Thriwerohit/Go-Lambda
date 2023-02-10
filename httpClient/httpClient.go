package httpClient

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func ParseClient(method, path string, body *strings.Reader, v interface{}) (*http.Response, error) {
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
		MaxIdleConns:        10,
		IdleConnTimeout:     10 * time.Second,
	}
	http_client := &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}
	rel := &url.URL{Path: path}

	baseUrl, err := url.Parse(os.Getenv("mongoParse"))
	if err != nil {
		log.Fatal(err)
	}

	u := baseUrl.ResolveReference(rel)
	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Add("X-Parse-Application-Id", os.Getenv("mongoApplicationKey"))
	req.Header.Add("X-Parse-Master-Key", os.Getenv("mongoMasterKey"))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	resp, err := http_client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		if resp.StatusCode == 400 {
			type temp struct {
				Code  int    `json:"code"`
				Error string `json:"error"`
			}
			var a temp
			_ = json.NewDecoder(resp.Body).Decode(&a)
			fmt.Printf("response error %v", a)
			errR := errors.New(fmt.Sprint(a))
			return nil, errR
		}

		err = fmt.Errorf(fmt.Sprintf("response error from parse client - %v", resp))
		return nil, err
	}
	defer resp.Body.Close()
	errJson := json.NewDecoder(resp.Body).Decode(v)
	if errJson != nil {
		return nil, errJson
	}
	return resp, err
}

func PostParseClient(method, path string, body *strings.Reader, v interface{}) (*http.Response, error) {
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
		MaxIdleConns:        10,
		IdleConnTimeout:     10 * time.Second,
	}
	http_client := &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}
	rel := &url.URL{Path: path}

	baseUrl, err := url.Parse(os.Getenv("postParse"))
	if err != nil {
		log.Fatal(err)
	}

	u := baseUrl.ResolveReference(rel)
	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Add("X-Parse-Application-Id", os.Getenv("postApplicationKey"))
	req.Header.Add("X-Parse-Master-Key", os.Getenv("postMasterKey"))
	req.Header.Add("X-Transactions-Server", os.Getenv("TransactionsServer"))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")

	resp, err := http_client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		if resp.StatusCode == 400 {
			type temp struct {
				Code  int    `json:"code"`
				Error string `json:"error"`
			}
			var a temp
			_ = json.NewDecoder(resp.Body).Decode(&a)
			fmt.Printf("response error %v", a)
			errR := errors.New(fmt.Sprint(a))
			return nil, errR
		}

		err = fmt.Errorf(fmt.Sprintf("response error from parse client - %v", resp))
		return nil, err
	}
	defer resp.Body.Close()
	errJson := json.NewDecoder(resp.Body).Decode(v)
	if errJson != nil {
		return nil, errJson
	}
	return resp, err
}
