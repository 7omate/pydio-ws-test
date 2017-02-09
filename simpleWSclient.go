package main

import (
	"flag"
	"time"
	"io/ioutil"
	"strconv"
	"encoding/json"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/hmac"
	"golang.org/x/net/websocket"
	"math/big"
	"crypto/rand"
	"strings"
	"log"
	"encoding/hex"
	"net/http"
)

func main() {
	// sample input
	var wsurl string = "ws://127.0.0.1:8090/ws?auth_hash=7fe7b47eee0ff9702a497ecea244896c43527561:c49d2f13bee944934623bf4c6047c04a68306d3758eedfe9bf00636be8f8eccc&auth_token=8GezB4bD0eNKY0sNRw7oIByi"
	var user, password, server string
	flag.StringVar(&wsurl, "wsURL", wsurl, "ws://.../ws?auth_has=...")
	flag.StringVar(&server, "server", server, "Server URL")
	flag.StringVar(&user, "user", user, "username")
	flag.StringVar(&password, "password", password, "password")
	flag.Parse()
	if user != "" && password != ""{
		config := get_server_data(server, user, password)
		log.Println("Server: ", server)
		log.Println(" WS URL from config: ", config.ws_url())
		tokens := basic_authenticate(server, user, password)
		ws_auth := gen_ws_authenticate_url(server, user, password, tokens)
		wsurl = config.ws_url() + "?" + ws_auth
	}
	log.Println(" Trying to connect to " + wsurl)
	ws, err := websocket.Dial(wsurl, "", server)
	if err != nil {
		log.Println(err)
		log.Fatal(" Failed to connect.")
	}
    mess := "register:1"
    log.Println(" Sending " + mess)
	websocket.Message.Send(ws, mess)
    log.Println(" [SUCCESS] Waiting for incomming messages")
	var data []byte
	for {
        err = websocket.Message.Receive(ws, &data)
        if err != nil {
            log.Println(err)
            break
        }
        log.Println(string(data[:]))
	}
}

func gen_ws_authenticate_url(server string, user string, password string, tokens map[string]interface{}) string {
	// generate a token
	endpoint := "/api/pydio/ws_authenticate"
	nonce := get_nonce()
	msg := endpoint + ":" + nonce + ":" + tokens["p"].(string)
	the_hash := Hmac256(msg, tokens["t"].(string))
	auth_hash := nonce + ":" + the_hash
	mess := "auth_hash=" + auth_hash + "&auth_token=" + tokens["t"].(string)
	return mess
}

func Hmac256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

func basic_authenticate(server, user, password string)map[string]interface{}{
	endpoint := "/api/pydio/keystore_generate_auth_token/websocket_test"
	var tokens map[string]interface{}
	client := &http.Client{Timeout: 20 *time.Second}
	uri := join(server, endpoint)
	log.Println(" GET ", uri)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		log.Println(err)
	}
	req.SetBasicAuth(user, password)
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(data, &tokens)
	if err != nil {
		log.Println(err)
	}
	//log.Println(tokens)
	return tokens
}

func get_server_data(server string, user string, password string) BoosterConfig{
	// get server infos
	endpoint := "api/pydio/state/plugins?format=json"
	client := &http.Client{Timeout: 20 *time.Second}
	uri := join(server, endpoint)
	//log.Println(uri)
	req, err := http.NewRequest("GET", uri, nil)
	req.SetBasicAuth(user, password)
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	config := BoosterConfig{}
	defer resp.Body.Close()
	var server_info map[string]interface{}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	err = json.Unmarshal(data, &server_info)
	if err != nil {
		log.Println(err)
		return config
	}
	plugins := server_info["plugins"].(map[string]interface{})
	for _, v := range plugins["ajxp_plugin"].([]interface{}) {
		plugin := v.(map[string]interface{})
		is_core_mq := false
		for k, w := range plugin {
			if k == "@id" && w == "core.mq"{
				is_core_mq = true
				break // correct plugin found
			}
		}
		if is_core_mq {
			for k, w := range plugin {
				if k == "plugin_configs"{
					server_settings := w.(map[string]interface{})
					for _, vv := range server_settings {
						params := vv.([]interface{})
						for _, val := range params {
							values := val.(map[string]interface{})
							//log.Println(values["@name"].(string) + " " + values["$"].(string))
							switch values["@name"] {
							case "WS_ACTIVE":
								config.WS_ACTIVE = paramToBool(values["$"].(string))
							case "BOOSTER_MAIN_HOST":
								config.BOOSTER_MAIN_HOST = strings.Trim(values["$"].(string), "\"")
							case "BOOSTER_MAIN_PORT":
								config.BOOSTER_MAIN_PORT = strings.Trim(values["$"].(string), "\"")
							case "BOOSTER_MAIN_SECURE":
								config.BOOSTER_MAIN_SECURE = paramToBool(values["$"].(string))
							case "WS_HOST":
								config.WS_HOST = strings.Trim(values["$"].(string), "\"")
							case "WS_PORT":
								config.WS_PORT = strings.Trim(values["$"].(string), "\"")
							case "WS_SECURE":
								config.WS_SECURE = paramToBool(values["$"].(string))
							case "WS_PATH":
								config.WS_PATH = strings.Trim(values["$"].(string), "\"")
							case "UPLOAD_HOST":
								config.UPLOAD_HOST = strings.Trim(values["$"].(string), "\"")
							case "UPLOAD_PORT":
								config.UPLOAD_PORT = strings.Trim(values["$"].(string), "\"")
							case "UPLOAD_SECURE":
								config.UPLOAD_SECURE = paramToBool(values["$"].(string))
							case "UPLOAD_ACTIVE":
								config.UPLOAD_ACTIVE = paramToBool(values["$"].(string))
							case "UPLOAD_PATH":
								config.UPLOAD_PATH = strings.Trim(values["$"].(string), "\"")
							// TODO: BOOSTER_WS_ADVANCED
							// TODO: BOOSTER_UPLOAD_ADVANCED
							}
						}
					}
				}
			}
		}
	}
	return config
}

func paramToBool(p string) bool{
	// convert string to bool
	val, err := strconv.ParseBool(p)
	if err != nil {
		log.Println("Problem converting " + p + " to boolean.")
	}
	return val
}

/* Booster config type */
type BoosterConfig struct {
	WS_ACTIVE bool
	BOOSTER_MAIN_HOST string
	BOOSTER_MAIN_PORT string // int ?
	BOOSTER_MAIN_SECURE bool
	//BOOSTER_WS_ADVANCED map[string]
	WS_HOST string
	WS_PORT string // int ?
	WS_SECURE bool
	WS_PATH string
	// BOOSTER_UPLOAD_ADVANCED map[string]
	UPLOAD_HOST string
	UPLOAD_PORT string // int?
	UPLOAD_SECURE bool
	UPLOAD_PATH string
	UPLOAD_ACTIVE bool
}

func (c BoosterConfig) ws_url() string{
	url := ""
	if c.WS_ACTIVE {
		if c.WS_SECURE {
			url += "wss://"
		} else {
			url += "ws://"
		}
		url += c.BOOSTER_MAIN_HOST
		if c.BOOSTER_MAIN_PORT != ""{
			url += ":"
			url += c.BOOSTER_MAIN_PORT
		}
		url += "/"
		url += c.WS_PATH
	}
	return url
}

func get_nonce() string{
	// TODO: maybe handle the error
	random_number,_ := rand.Int(rand.Reader, big.NewInt(1000000))
	b := sha1.Sum( random_number.Bytes() )
	//log.Println(hex.EncodeToString( b[:] ))
	return hex.EncodeToString(b[:])
}

func join(a, b string) string{
	return strings.TrimSuffix(a, "/") + "/" + strings.TrimPrefix(b, "/")
}

