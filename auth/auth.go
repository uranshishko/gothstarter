package auth

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/uranshishko/gothstarter/common"
)

var SessionName = "_rw_session"

const (
	UserEndpoint     = "https://graph.microsoft.com/v1.0/me"
	AuthFlowEndpoint = "https://login.microsoftonline.com/%s/oauth2/v2.0/%s"
)

var Client *msalClient

type msalClient struct {
	tenantId     string
	clientId     string
	clientSecret string
	callBackUrl  string
	Store        sessions.Store
}

type AuthResponse struct {
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	ExtExpiresIn int    `json:"ext_expires_in"`
	AccessToken  string `json:"access_token"`
}

type User struct {
	Id          string `json:"id"`
	DisplayName string `json:"displayName"`
	Mail        string `json:"mail"`
	FirstName   string `json:"givenName"`
	LastName    string `json:"surname"`
}

func (u *User) Marshal() string {
	b, _ := json.Marshal(u)
	return string(b)
}

type AuthError struct {
	Err            string `json:"error"`
	ErrDescription string `json:"error_description"`
}

func (e AuthError) Error() string {
	return fmt.Sprintf("%s: %s", e.Err, e.ErrDescription)
}

func NewAuthError(err, desc string) AuthError {
	return AuthError{err, desc}
}

func NewMsalClient(tenantId, clientId, clientSecret, callBackUrl string) {
	Client = &msalClient{
		tenantId:     tenantId,
		clientId:     clientId,
		clientSecret: clientSecret,
		callBackUrl:  callBackUrl,
	}
}

func (c *msalClient) codeFlowURL(silently bool) *url.URL {
	u, _ := url.Parse(fmt.Sprintf(AuthFlowEndpoint, c.tenantId, "authorize"))
	q := u.Query()

	q.Add("client_id", c.clientId)
	q.Add("scope", "https://graph.microsoft.com/user.readwrite")
	q.Add("response_type", "code")
	q.Add("response_mode", "query")
	q.Add("redirect_uri", c.callBackUrl)
	if silently {
		q.Add("prompt", "none")
	} else {
		q.Add("prompt", "select_account")
	}

	u.RawQuery = q.Encode()

	return u
}

func (c *msalClient) requestAccessToken(code string) ([]byte, error) {
	u, _ := url.Parse(fmt.Sprintf(AuthFlowEndpoint, c.tenantId, "token"))
	data := &url.Values{}

	data.Set("client_id", c.clientId)
	data.Set("scope", "https://graph.microsoft.com/user.readwrite")
	data.Set("code", code)
	data.Set("redirect_uri", c.callBackUrl)
	data.Set("grant_type", "authorization_code")
	data.Set("client_secret", c.clientSecret)

	authRes, err := common.MakeRequest("POST", u.String(), []byte(data.Encode()), func(o *common.RequestOpts) { o.Headers["Content-Type"] = "application/x-www-form-urlencoded" })
	if err != nil {
		return nil, err
	}

	if authRes.StatusCode >= 200 && authRes.StatusCode < 300 {
		return authRes.ReadBody()
	}

	body, err := authRes.ReadBody()
	if err != nil {
		return nil, err
	}

	var authErr AuthError
	err = json.Unmarshal(body, &authErr)
	if err != nil {
		return nil, err
	}

	return nil, authErr
}

func (c *msalClient) BeginAuth(hc common.HandlerContext) {
	u := c.codeFlowURL(false)
	hc.Redirect(302, u.String())
}

func (c *msalClient) BeginSilentAuth(hc common.HandlerContext) {
	u := c.codeFlowURL(true)
	hc.Redirect(302, u.String())
}

func (c *msalClient) CompleteAuth(hc common.HandlerContext) (User, error) {
	var user User

	r := hc.Request()
	w := hc.Response()

	s, _ := c.Store.Get(r, SessionName)

	if user, err := fetchUser(s); err == nil {
		return user, nil
	}

	q := r.URL.Query()
	code := q.Get("code")
	e := q.Get("error")
	errDesc := q.Get("error_description")

	if e != "" {
		return user, NewAuthError(e, errDesc)
	}

	authRes, err := c.requestAccessToken(code)
	if err != nil {
		return user, err
	}

	var authToken AuthResponse
	err = json.Unmarshal(authRes, &authToken)
	if err != nil {
		return user, err
	}

	err = setSessionValue(s, "AccessToken", authToken.AccessToken)
	if err != nil {
		return user, err
	}

	user, err = fetchUser(s)
	if err != nil {
		return user, err
	}

	err = setSessionValue(s, "user", user.Marshal())
	if err != nil {
		return user, err
	}

	err = s.Save(r, w)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (c *msalClient) Logout(hc common.HandlerContext) error {
	r := hc.Request()
	w := hc.Response()

	s, err := c.Store.Get(r, SessionName)
	if err != nil {
		return err
	}

	s.Options.MaxAge = -1
	s.Values = make(map[interface{}]interface{})
	err = s.Save(r, w)
	if err != nil {
		return fmt.Errorf("could not delete user session")
	}

	return nil
}

func fetchUser(s *sessions.Session) (User, error) {
	var user User

	token, err := getSessionValue(s, "AccessToken")
	if err != nil {
		return user, err
	}

	userRes, err := common.MakeRequest("GET", UserEndpoint, nil, func(o *common.RequestOpts) { o.Headers["Authorization"] = fmt.Sprintf("Bearer %s", token) })
	if err != nil {
		return user, err
	}

	body, err := userRes.ReadBody()
	if err != nil {
		return user, err
	}

	err = json.Unmarshal(body, &user)
	if err != nil {
		return user, err
	}

	return user, nil
}

func (c *msalClient) StoreInSession(hc common.HandlerContext, key string, value string) error {
	r := hc.Request()
	w := hc.Response()

	s, _ := c.Store.Get(r, SessionName)
	err := setSessionValue(s, key, value)
	if err != nil {
		return err
	}

	return s.Save(r, w)
}

func (c *msalClient) GetFromSession(hc common.HandlerContext, key string) (string, error) {
	r := hc.Request()

	s, _ := c.Store.Get(r, SessionName)
	value, err := getSessionValue(s, key)
	if err != nil {
		return "", err
	}

	return value, nil
}

func setSessionValue(s *sessions.Session, key string, value string) error {
	var b bytes.Buffer

	gz := gzip.NewWriter(&b)
	if _, err := gz.Write([]byte(value)); err != nil {
		return err
	}

	if err := gz.Flush(); err != nil {
		return err
	}

	if err := gz.Close(); err != nil {
		return err
	}

	s.Values[key] = b.String()

	return nil
}

func getSessionValue(s *sessions.Session, key string) (string, error) {
	value := s.Values[key]
	if value == nil {
		return "", fmt.Errorf("could not find session value")
	}

	rdata := strings.NewReader(value.(string))
	gr, err := gzip.NewReader(rdata)
	if err != nil {
		return "", err
	}

	str, err := io.ReadAll(gr)
	if err != nil {
		return "", err
	}

	return string(str), nil
}
