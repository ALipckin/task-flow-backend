package controllers

import (
	"TaskRestApiService/initializers"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func ProxyRequest(c *gin.Context, targetURL string) {
	log.Printf("Proxying request to %s", targetURL)
	initializers.LogToKafka("info", "ProxyRequest", "Proxy", targetURL)
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		initializers.LogToKafka("error", "ProxyRequest", "Failed to proxy", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid target URL"})
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: parsedURL.Scheme,
		Host:   parsedURL.Host,
	})

	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = parsedURL.Scheme
		req.URL.Host = parsedURL.Host
		req.Host = parsedURL.Host

		req.URL.Path = parsedURL.Path

		req.URL.RawQuery = parsedURL.RawQuery
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}
