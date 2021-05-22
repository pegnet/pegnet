package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

var APIHost = "localhost:8099"

// SendRequest will send a PostRequest object to the local API server and return the PostResponse
func SendRequest(req *PostRequest) (*PostResponse, error) {
	j, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	httpx := "http"
	server := APIHost

	re, err := http.NewRequest("POST", fmt.Sprintf("%s://%s/v1", httpx, server), bytes.NewBuffer(j))
	if err != nil {
		return nil, err
	}

	re.Header.Add("content-type", "application/json")
	resp, err := client.Do(re)
	if err != nil {
		errs := fmt.Sprintf("%s", err)
		if strings.Contains(errs, "\\x15\\x03\\x01\\x00\\x02\\x02\\x16") {
			err = fmt.Errorf("Factom-walletd API connection is encrypted. Please specify -wallettls=true and -walletcert=walletAPIpub.cert (%v)", err.Error())
		}
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("Wallet username/password incorrect.  Edit factomd.conf or\ncall factom-cli with -walletuser=<user> -walletpassword=<pass>")
	}

	r := &PostResponse{}
	if err := json.Unmarshal(body, r); err != nil {
		return nil, err
	}

	return r, nil
}
