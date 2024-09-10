package proxy

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fobilow/ham/helper"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

var loginPage = helper.GetEnv("LOGIN_PAGE", "login.html")
var webRoot = helper.GetEnv("WEB_ROOT", "./public")
var appPort = helper.GetEnv("PROXY_PORT", "8082")
var apiEndpoint = helper.GetEnv("API_ENDPOINT", "http://localhost:8080")
var apiProxyPrefix = helper.GetEnv("API_PROXY_PREFIX", "/api/")

var appSession = map[string]*Session{}

type Session struct {
	AccessToken string    `json:"access_token,omitempty"`
	Expiry      time.Time `json:"expiry"`
}

func (s *Session) IsInvalid() bool {
	return time.Now().After(s.Expiry)
}

func GetSession(token string) *Session {
	s, ok := appSession[token]
	if !ok {
		return &Session{}
	}

	return s
}

func Run() {
	fmt.Printf("Run Parameters:\n API_ENDPOINT: %s\n LOGIN_PAGE: %s\n WEB_ROOT: %s\n PROXY_PORT: %s\n API_PROXY_PREFIX: %s\n",
		apiEndpoint, loginPage, webRoot, appPort, apiProxyPrefix)
	if _, err := os.Stat(webRoot); err != nil {
		log.Fatal("web root does not exist. ham-proxy cannot start without a web root")
	}

	router := gin.Default()
	router.Use(gzip.Gzip(gzip.BestCompression))
	router.Any(apiProxyPrefix+"*path", func(c *gin.Context) {
		handleApiRequest(c)
	})
	router.NoRoute(func(c *gin.Context) {
		handleWebRequest(c)
	})

	log.Fatal(router.Run(fmt.Sprintf(":%s", appPort)))
}

func handleApiRequest(c *gin.Context) {
	log.Println("handling API Request...")
	target, err := url.Parse(apiEndpoint)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Director = func(req *http.Request) {
		req.Header.Del("Accept-Encoding") // prevent double compression
	}
	proxy.ModifyResponse = func(res *http.Response) error {
		if token := res.Header.Get("X-HAM-PROXY-TOKEN"); len(token) > 0 {
			fmt.Println("proxy token in response", token)
			session := strings.Split(token, "~")
			expiry, _ := strconv.Atoi(session[1])

			cookie := http.Cookie{
				Name:     "access_token",
				Value:    session[0],
				Path:     "/",
				Domain:   "",
				Expires:  time.Now().Add(time.Hour),
				Secure:   false,
				HttpOnly: false,
			}
			res.Header.Add("Set-Cookie", cookie.String())

			appSession[session[0]] = &Session{
				AccessToken: session[0],
				Expiry:      time.Unix(int64(expiry), 0),
			}
		} else {
			tokenCookie, err := c.Request.Cookie("access_token")
			if err != nil {
				c.Redirect(302, loginPage)
				return nil
			}
			if res.StatusCode == http.StatusUnauthorized {
				delete(appSession, tokenCookie.Value)
				c.Redirect(302, loginPage)
				return nil
			}
		}

		return nil
	}
	uri, _ := url.ParseRequestURI(strings.Replace(c.Request.RequestURI, apiProxyPrefix, "/", 1))
	c.Request.URL.Scheme = target.Scheme
	c.Request.URL.Host = target.Host
	c.Request.URL.Path = uri.Path
	c.Request.URL.RawQuery = uri.RawQuery
	log.Println("final request URL", c.Request.URL.String())
	proxy.ServeHTTP(c.Writer, c.Request)
	return
}

func handleWebRequest(c *gin.Context) {
	log.Println("handling WEB Request...")
	dir, file := path.Split(c.Request.RequestURI)
	file = strings.Split(file, "?")[0]
	if file == "" {
		file = "index.html"
	}
	ext := filepath.Ext(file)
	if c.Request.RequestURI == "/" {
		ext = ".html"
		file = "index.html"
	}

	file = webRoot + path.Join(dir, file)

	switch ext {
	case ".html":
		b, err := os.ReadFile(file)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		authPage := strings.Contains(string(b), `data-ham-proxy="requires-authentication"`)
		if authPage {
			// check if user is logged in
			tokenCookie, err := c.Request.Cookie("access_token")
			if err != nil {
				c.Redirect(302, loginPage)
				return
			}
			if GetSession(tokenCookie.Value).IsInvalid() {
				c.Redirect(302, loginPage)
				return
			}
		}
	}

	log.Println("File:", file)
	c.File(file)
}