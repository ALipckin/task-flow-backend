package controllers

import (
	"TaskRestApiService/logger"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func ProxyRequest(c *gin.Context, targetURL string) {
	logger.Log(logger.LevelInfo, "ProxyRequest", gin.H{"targetURL": targetURL})
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		logger.Log(logger.LevelError, "Failed to proxy", gin.H{"error": err.Error()})
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
