package server

import (
	"fmt"
	"github.com/amorbielyi/asyncSubdomainChecker/checker"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

/*
 Through "http:/localhost:8080/domain/status/[domainName]" route
 it sends GET-HTTP request to underlying 3-rd part API "https://sonar.omnisint.io/subdomains/[root]",
 in order to get response body which contains all subdomains of [domainName] domain",
 and after that in asynchronous way using goroutines it calls http.Get(someSubdomain) on each subdomain and
 checks its global availability (e.g whether we are able to reach the server by address or not)
 and sends JSON as response to the same route.
 JSON response is array of Results
*/

// I prefer to use sonar.omnisint as 3-rd party middleware API to achieve these goals
const middlewareApiRoute = "https://sonar.omnisint.io/subdomains"

const addr = "localhost" + ":" + port
const port = "8080"

//  Handler with logic that binds to "domain/status/:root" route
func getSubdomainStatus(c *gin.Context) {
	rootDomainParam := c.Param("root")
	log.Printf("checking subdomains reachebility for %s root domain", rootDomainParam)

	resp, err := http.Get(fmt.Sprintf("%s/%s", middlewareApiRoute, rootDomainParam))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "could not HTTP-GET " + middlewareApiRoute})
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic("could not read response body")
		}
		replacer := strings.NewReplacer("[", "", "]", "", "\"", "")

		payload := replacer.Replace(string(body))
		achecker := checker.NewAsyncSubdomainChecker(strings.Split(payload, ","))
		res := achecker.GetResults()
		fmt.Println("subdomains count: ", len(res))
		c.JSON(http.StatusOK, res)
	}
}

// Inits and launches GIN HTTP-Web-Server on localhost:8080
// to handle HTTP requests via REST API.
func Launch() {
	router := gin.Default()

	// GET-Route parametrized by 'root' param that hols root domain name,
	// usage: 'http://localhost:8080/domain/status/dropbox.com'.
	router.GET("domain/status/:root", getSubdomainStatus)

	// Run server and check errors
	if err := router.Run(addr); err != nil {
		log.Fatalf("error running web server %v", err)
	}
}

func init() {
	log.SetPrefix("server ")
}