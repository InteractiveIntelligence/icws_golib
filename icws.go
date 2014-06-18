//the icws_golib package wraps CIC ICWS fuctionality in a GO library.
//To get started, call icws_golib.NewIcws() to instansiate a new Icws
//struct.  Then call icws.Login(...) and continue on from there.
package icws_golib

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"log"
)

type Icws struct {
	CurrentToken, CurrentCookie, CurrentSession, CurrentServer string
}

//Creates a new ICWS struct
func NewIcws() (icws *Icws) {
	icws = &Icws{}
	return
}


func (i *Icws) loginWithData(applicationName, server, username, password string, loginData map[string]string) (err error) {

	i.CurrentSession = "";
    i.CurrentServer = server;
	body, err, cookie := i.httpPost("connection", loginData)

    i.CurrentCookie = cookie;

	if err == nil {
		var returnData map[string]string
		json.Unmarshal(body, &returnData)
		i.CurrentToken = returnData["csrfToken"]
		i.CurrentSession = returnData["sessionId"]
        i.CurrentServer = server;

	} else {
		log.Printf("ERROR: %s\n", err.Error())
	}
	return
}

//Logs into a CIC server.  Server should be a url e.g. https://MyServer:8019
func (i *Icws) Login(applicationName, server, username, password string) (err error) {

    var loginData = map[string]string{
        "__type":          "urn:inin.com:connection:icAuthConnectionRequestSettings",
        "applicationName": applicationName,
        "userID":          username,
        "password":        password,
    }

    return i.loginWithData(applicationName, server, username, password, loginData);
}

//Logs into a CIC server for a MarketPlace application using the app's custom license.  Server should be a url e.g. https://MyServer:8019
func (i *Icws) LoginMarketPlaceApp(applicationName, server, username, password, marketplaceLicense, markeplaceAppKey string) (err error) {

    var loginData = map[string]string{
        "__type":          "urn:inin.com:connection:icAuthConnectionRequestSettings",
        "applicationName": applicationName,
        "userID":          username,
        "password":        password,
    }

    return i.loginWithData(applicationName, server, username, password, loginData);
}


func (i *Icws) httpDelete(url string) ( err error) {

	req, err := i.httpRequest("DELETE", url, nil)

	if err != nil {
		return
	}

	response, err := i.httpClient().Do(req)
	if err != nil {
		return
	}
	defer response.Body.Close()
	if response.StatusCode == 401 {
		err = errors.New("authorization expired")
		return
	}
	body, _ := ioutil.ReadAll(response.Body)
	if response.StatusCode/100 != 2 {
		err = errors.New(createErrorMessage(response.StatusCode, body))

		return
	}

	return
}

func (i *Icws) httpGet(url string) (body []byte, err error) {

	req, err := i.httpRequest("GET", url, nil)

	if err != nil {
		return
	}

	response, err := i.httpClient().Do(req)
	if err != nil {
		return
	}
	defer response.Body.Close()
	if response.StatusCode == 401 {
		err = errors.New("authorization expired, please run `cic login`")
		return
	}
	body, err = ioutil.ReadAll(response.Body)
	if response.StatusCode/100 != 2 {
		err = errors.New(createErrorMessage(response.StatusCode, body))

		return
	}

	return
}



func (i *Icws) httpPost(url string, attrs map[string]string) (body []byte, err error, cookie string) {

	rbody, _ := json.Marshal(attrs)
	req, err := i.httpRequest("POST", url, bytes.NewReader(rbody))
	if err != nil {
		return
	}

	response, err := i.httpClient().Do(req)
	if err != nil {
		return

	}
	defer response.Body.Close()
	if response.StatusCode == 401 {
		err = errors.New("authorization expired")
		return
	}
	body, err = ioutil.ReadAll(response.Body)

	if response.StatusCode/100 != 2 {
		err = errors.New(createErrorMessage(response.StatusCode, body))

		return
	}


	if response.Header["Set-Cookie"] != nil {
		cookie = response.Header["Set-Cookie"][0]
	}
	return
}

func (i *Icws) httpPut(url string, attrs map[string]string) (body []byte, err error) {

	rbody, _ := json.Marshal(attrs)
	req, err := i.httpRequest("PUT", url, bytes.NewReader(rbody))
	if err != nil {
		return
	}

	response, err := i.httpClient().Do(req)
	if err != nil {
		return

	}
	defer response.Body.Close()
	if response.StatusCode == 401 {
		err = errors.New("authorization expired")
		return
	}
	body, err = ioutil.ReadAll(response.Body)

	if response.StatusCode/100 != 2 {
		err = errors.New(createErrorMessage(response.StatusCode, body))

		return
	}

	return
}

func (i *Icws) httpClient() (client *http.Client) {

	client = &http.Client{}
	return
}

func (i *Icws) httpRequest(method, url string, body io.Reader) (request *http.Request, err error) {
	request, err = http.NewRequest(method, i.CurrentServer + "/icws/" + i.CurrentSession + url, body)
	if err != nil {
		return
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Accept-Language", "en-us")

	if len(i.CurrentCookie) > 0 {
		request.Header.Add("Cookie", i.CurrentCookie)
	} else {
		return
	}

	if len(i.CurrentToken) > 0 {
		request.Header.Add("ININ-ICWS-CSRF-Token", i.CurrentToken)
	} else {
		return
	}

	// request.Header.Add("User-Agent", fmt.Sprintf("cic cli (%s-%s)", runtime.GOOS, runtime.GOARCH))
	return
}

func createErrorMessage(statusCode int, body []byte) string {

	var errorDescription string

	switch statusCode {
	case 400:
		errorDescription = "Bad Request (400)"
	case 401:
		errorDescription = "Unauthorized (401)"
	case 403:
		errorDescription = "Forbidden (403)"
	case 404:
		errorDescription = "Not Found (404)"
	case 410:
		errorDescription = "Gone (410)"
	case 500:
		errorDescription = "Internal Server Error (500)"
	}

	var message map[string]interface{}
	json.Unmarshal(body, &message)

    if message["errorId"] != nil{
	       errorDescription += ": " + message["errorId"].(string);
    }

    if(message["message"] != nil){
        errorDescription += " " + message["message"].(string)
    }
    return errorDescription;
}