package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

func modifyResponse() func(*http.Response) error {
    return func(resp *http.Response) error {
        resp.Header.Set("Access-Control-Allow-Origin", "*")
		fmt.Print(resp.Status)
        return nil
    }
}

func process(c *gin.Context) {
	remote, err := url.Parse(c.Param("proxyPath")[1:])
	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
		req.URL.Path = remote.Path
	}
	proxy.ModifyResponse = modifyResponse()

	proxy.ServeHTTP(c.Writer, c.Request)
}

func main() {
	router := gin.Default()
	router.Any("/*proxyPath", process)

	router.Run("localhost:4000")
}
