package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"oauth2/models"
	"oauth2svc/ldap"

	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/errors"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/server"
	"gopkg.in/oauth2.v3/store"
)

var clientStore *store.ClientStore
var base, bindDN, bindPassword, groupFilter, host, serverName, userFilter string
var port int
var useSSL bool
var skipTLS bool

func ldapVerified(username, password string) (err error) {

	client := &ldap.LDAPClient{
		Base:         base,
		Host:         host,
		Port:         port,
		UseSSL:       useSSL,
		SkipTLS:      skipTLS,
		BindDN:       bindDN,
		BindPassword: bindPassword,
		UserFilter:   userFilter,
		GroupFilter:  groupFilter,
		Attributes:   []string{"givenName", "sn", "uid"},
		ServerName:   serverName,
	}
	defer client.Close()

	ok, user, err := client.Authenticate(username, password)
	if err != nil {
		log.Printf("Error authenticating user %s: %+v", username, err)
		return
	}
	if !ok {
		log.Printf("Authenticating failed for user %s", username)
		return
	}
	log.Printf("User: %+v", user)

	groups, err := client.GetGroupsOfUser(username)
	if err != nil {
		log.Printf("Error getting groups for user %s: %+v", username, err)
		return
	}
	log.Printf("Groups: %+v", groups)
	return
}

func init() {
	flag.StringVar(&base, "base", "dc=k8s,dc=com", "Base LDAP")
	flag.StringVar(&bindDN, "bind-dn", "cn=admin,dc=k8s,dc=com", "Bind DN")
	flag.StringVar(&bindPassword, "bind-pwd", "password", "Bind password")
	flag.StringVar(&groupFilter, "group-filter", "(memberUid=%s)", "Group filter")
	flag.StringVar(&host, "host", "10.20.0.19", "LDAP host")
	// flag.StringVar(&password, "password", "", "Password")
	// flag.StringVar(&username, "username", "", "Username")
	flag.IntVar(&port, "port", 389, "LDAP port")
	flag.StringVar(&userFilter, "user-filter", "(uid=%s)", "User filter")
	flag.StringVar(&serverName, "server-name", "", "Server name for SSL (if use-ssl is set)")
	flag.BoolVar(&useSSL, "use-ssl", false, "Use SSL")
	flag.BoolVar(&skipTLS, "skip-tls", true, "Skip TLS start")
}

func clientAuthorizedHandler(clientID string, grant oauth2.GrantType) (allowed bool, err error) {
	fmt.Printf(clientID)
	return true, nil
}

// Ldap verified first
func clientLDapHandler(r *http.Request) (clientID, clientSecret string, err error) {
	clientID = r.Form.Get("client_id")
	clientSecret = r.Form.Get("client_secret")
	if clientID == "" || clientSecret == "" {
		err = errors.ErrInvalidClient
	}
	//fmt.Println(clientID, clientSecret)
	_, cerr := clientStore.GetByID(clientID)
	if cerr != nil { //not exist
		err = ldapVerified(clientID, clientSecret)
		if err == nil {
			clientStore.Set(clientID, &models.Client{
				ID:     clientID,
				Secret: clientSecret,
				Domain: "http://localhost",
			})
			fmt.Println(clientID, "added")
		}

	}

	return
}

func main() {
	flag.Parse()
	manager := manage.NewDefaultManager()
	// token memory store
	manager.MustTokenStorage(store.NewMemoryTokenStore())

	// client memory store
	// clientStore := store.NewClientStore()
	clientStore = store.NewClientStore()
	// clientStore.Set("user1", &models.Client{
	// 	ID:     "user1",
	// 	Secret: "password",
	// 	Domain: "http://localhost",
	// })
	// clientStore.Set("user2", &models.Client{
	// 	ID:     "user2",
	// 	Secret: "password",
	// 	Domain: "http://localhost",
	// })
	manager.MapClientStorage(clientStore)

	srv := server.NewDefaultServer(manager)
	srv.SetClientAuthorizedHandler(clientAuthorizedHandler)
	srv.SetAllowGetAccessRequest(true)
	//srv.SetClientInfoHandler(server.ClientFormHandler)
	srv.SetClientInfoHandler(clientLDapHandler)

	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		log.Println("Internal Error:", err.Error())
		return
	})

	srv.SetResponseErrorHandler(func(re *errors.Response) {
		log.Println("Response Error:", re.Error.Error())
	})

	http.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		err := srv.HandleAuthorizeRequest(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})

	http.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		// clientID, clientSecret, err := srv.ClientInfoHandler(r)
		// if err != nil {
		// 	return
		// }
		// fmt.Println(clientID, clientSecret)
		srv.HandleTokenRequest(w, r)
	})

	log.Fatal(http.ListenAndServe(":9096", nil))
}
