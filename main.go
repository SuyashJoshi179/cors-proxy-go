package main

import (
	// "fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

var allowedUrls = []string{
	"http://localhost:3000",
	"https://localhost:3000",
	"https://crypt-web.web.app",
	"https://crypt-web.web.app/",
}

// function to check whether the backendServer parameter is a valid url or not
func isValidUrl(urlString string) bool {
    _, err := url.ParseRequestURI(urlString)
    return err == nil
}

// function to check whether the request is coming from an allowed URL or not
func isAllowedUrl(origin string) bool {
    for _, u := range allowedUrls {
        if u == origin {
            return true
        }
    }
    return false
}

func process(c *gin.Context) {
	// Extract the URL from request path
	backendServer := strings.TrimLeft(c.Param("proxyPath"), "/")

	// check if the URL is valid
	if backendServer == "" || !isValidUrl(backendServer) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid backend URL in the path"})
		return
	}

	// Check if the request is coming from an allowed Origin
	origin := c.Request.Header.Get("Origin")
	// fmt.Println("Origin: ", origin)
	if !isAllowedUrl(origin) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Unauthorized access"})
		return
	}
	
	// Create a new proxy
	remote, err := url.Parse(backendServer)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid backend URL in the path"})
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Header.Set("Origin", "")
		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
		req.URL.Path = remote.Path
	}

	// Add CORS headers to the response
	c.Header("Access-Control-Allow-Origin", "*")
	// c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
	// c.Header("Access-Control-Allow-Headers", "Content-Type")

	// checking for unexpected errors
	defer func() {
		if r := recover(); r != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected Error Occurred"})
		}
	}()
	
	// Serve the request via the proxy
	proxy.ServeHTTP(c.Writer, c.Request)
}

func main() {
	router := gin.Default()
	router.Any("/*proxyPath", process)

	router.Run("localhost:4000")
}
