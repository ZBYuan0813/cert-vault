package pkg

import (
	"crypto/tls"
	"encoding/json"
	vault "github.com/hashicorp/vault/api"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type Role struct {
	RoleName      string     `json:"role"`
	RoleData       RoleData  `json:"data"`
}

type RoleData struct {
	Allowed_Domains   []string  `json:"allowed_domains"`
	Allow_subdomains  bool      `json:"allow_subdomains"`
	Allow_Any_Name    bool      `json:"allow_any_name"`
	Organization      string    `json:"organization"`
	Ou                string    `json:"ou"`
	Max_TTL           string    `json:"max_ttl"`
}

type Cert struct {
	RoleName         string    `json:"role"`
	CertData         CertData  `json:"cert"`
}

type CertData struct {
	CommonName       string `json:"common_name"`
}

type RevokeData struct{
	SerialNumber  string `json:"serial_number"`
}

func init(){
	if os.Getenv("VAULT_ADDR") == "" {
		os.Setenv("VAULT_ADDR", "http://vault.company.com.cn")
	}
	if os.Getenv("VAULT_TOKEN") == "" {
		os.Setenv("VAULT_TOKEN", "s.rPB1z1oTI5y4Ax2xcAhnFbFQ")
	}

}


func CreateVaultConfig() *vault.Client{
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	config := vault.Config{
		Address:   os.Getenv("VAULT_ADDR"),
		HttpClient: &http.Client{Transport: tr},
		Timeout:    3 * time.Second,
	}

	cv,err := vault.NewClient(&config)
	if err != nil{
		log.Println("NewClient err")
		return nil
	}
	cv.SetToken(os.Getenv("VAULT_TOKEN"))
	return cv
}

func CreateRole(path string, role string, body []byte,client * vault.Client) int{
	requestPath := path + role
	req := client.NewRequest("POST",requestPath)
	req.BodyBytes = body
	res, err := client.RawRequest(req)
	if err != nil{
		log.Println(err)
	}
	defer res.Body.Close()
	return res.StatusCode
}

func CreateCert(path string,role string ,body []byte,client *vault.Client) (ca map[string][]byte){
	requestPath := path + role
	req := client.NewRequest("POST",requestPath)
	req.BodyBytes = body
	res, _ := client.RawRequest(req)

	if res.StatusCode >= 400 {
		reason,_ := ioutil.ReadAll(res.Body)
		log.Printf("Create CA failed %s", reason)
		return nil
	}

	var v map[string]interface{}
	dec := json.NewDecoder(res.Body)
	for err := dec.Decode(&v); err != nil && err != io.EOF; {
		log.Printf("Res body decode Error: %s \n" , err.Error())
		return nil
	}
	log.Println(res.StatusCode)
	var out = make(map[string][]byte)
	tmp := v["data"].(map[string]interface{})
	//fmt.Println(tmp)
	for k,val := range tmp{
		if k == "private_key" || k == "certificate" || k == "issuing_ca" || k=="serial_number"{
			out[k] = []byte(val.(string))
		}
	}
	return out
}

func RevokeCert(path string, body []byte,client * vault.Client) int{
	// path: v1/pki/revoke
	requestPath := path
	req := client.NewRequest("POST",requestPath)
	req.BodyBytes = body
	res, err := client.RawRequest(req)
	if err != nil{
		log.Println(err)
	}
	defer res.Body.Close()
	return res.StatusCode
}














