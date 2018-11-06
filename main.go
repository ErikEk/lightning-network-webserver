package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
	"os/user"
	"path/filepath"
	"strings"
	"strconv"
	"net/http"
	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/macaroons"
	"github.com/roasbeef/btcutil"
	//"golang.org/x/crypto/acme/autocert"
	//"golang.org/x/net/context"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	macaroon "gopkg.in/macaroon.v2"
	"html/template"
	//"google.golang.org/api/option"
	"github.com/gorilla/websocket"
)

const (
	defaultTLSCertFilename  = "tls.cert"
	defaultMacaroonFilename = "admin.macaroon"
)
var (
	tpl *template.Template
	tlsCert     string
	rpcMacaroon string
	rpcServer   = defaultRPCServer
	lndDir      = defaultLndDir
	listenPort  = defaultPort
	firebaseApp *firebase.App
	firebaseDb  *firestore.Client

	defaultLndDir       = btcutil.AppDataDir("lnd", false)
	defaultTLSCertPath  = filepath.Join(defaultLndDir, defaultTLSCertFilename)
	defaultMacaroonPath = filepath.Join(defaultLndDir, defaultMacaroonFilename)
	defaultRPCServer    = "localhost:10009"
	defaultPort         = 8080
	newinvoice 			= ""
)


type clientss struct {
	active	bool
	money 	int
}
var clients = map[*websocket.Conn]*clientss{}

var broadcast = make(chan Message)      // broadcast channel

// Define our message object
type Message struct {
	Email         string `json:"email"`
	Username      string `json:"username"`
	Message  	  string `json:"message"`
	Payed	 	  string `json:"payed"`
	Value    	  string `json:"value"`
    AskForInvoice string `json:"askforinvoice"`
}
// Configure the upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
func fatal(err error) {
	fmt.Fprintf(os.Stderr, "[LNSite] %v\n", err)
	os.Exit(1)
}

func getClient() (lnrpc.LightningClient, func()) {
	conn := getClientConn()

	cleanUp := func() {
		conn.Close()
	}
	return lnrpc.NewLightningClient(conn), cleanUp
}

// Taken from lnd's lncli command.
func getClientConn() *grpc.ClientConn {
	lndDir := cleanAndExpandPath(lndDir)
	if lndDir != defaultLndDir {
		// If a custom lnd directory was set, we'll also check if custom
		// paths for the TLS cert and macaroon file were set as well. If
		// not, we'll override their paths so they can be found within
		// the custom lnd directory set. This allows us to set a custom
		// lnd directory, along with custom paths to the TLS cert and
		// macaroon file.
		tlsCertPath := cleanAndExpandPath(tlsCert)
		if tlsCertPath == defaultTLSCertPath {
			tlsCert = filepath.Join(lndDir, defaultTLSCertFilename)
		}

		macPath := cleanAndExpandPath(rpcMacaroon)
		if macPath == defaultMacaroonPath {
			rpcMacaroon = filepath.Join(lndDir, defaultMacaroonFilename)
		}
	}

	// Load the specified TLS certificate and build transport credentials
	// with it.
	tlsCertPath := cleanAndExpandPath(tlsCert)
	creds, err := credentials.NewClientTLSFromFile(tlsCertPath, "")
	if err != nil {
		fatal(err)
	}

	// Create a dial options array.
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
	}

	// Load the specified macaroon file.
	macPath := cleanAndExpandPath(rpcMacaroon)
	macBytes, err := ioutil.ReadFile(macPath)
	if err != nil {
		fatal(err)
	}
	mac := &macaroon.Macaroon{}
	if err = mac.UnmarshalBinary(macBytes); err != nil {
		fatal(err)
	}

	macConstraints := []macaroons.Constraint{
		// We add a time-based constraint to prevent replay of the
		// macaroon. It's good for 60 seconds by default to make up for
		// any discrepancy between client and server clocks, but leaking
		// the macaroon before it becomes invalid makes it possible for
		// an attacker to reuse the macaroon. In addition, the validity
		// time of the macaroon is extended by the time the server clock
		// is behind the client clock, or shortened by the time the
		// server clock is ahead of the client clock (or invalid
		// altogether if, in the latter case, this time is more than 60
		// seconds).
		// TODO(aakselrod): add better anti-replay protection.
		macaroons.TimeoutConstraint(60),
	}

	// Apply constraints to the macaroon.
	constrainedMac, err := macaroons.AddConstraints(mac, macConstraints...)
	if err != nil {
		fatal(err)
	}

	// Now we append the macaroon credentials to the dial options.
	cred := macaroons.NewMacaroonCredential(constrainedMac)
	opts = append(opts, grpc.WithPerRPCCredentials(cred))

	conn, err := grpc.Dial(rpcServer, opts...)
	if err != nil {
		fatal(err)
	}

	return conn
}


func main() {

	tlsCertFlag := flag.String("tlsCert", defaultTLSCertPath, "path for the certificate used by the lnd server.")
	rpcMacaroonFlag := flag.String("macaroon", defaultMacaroonPath, " path for the macaroon.")
	rpcServerFlag := flag.String("rpcServer", defaultRPCServer, "rpc server to connect to.")
	listenPortFlag := flag.Int("port", defaultPort, "port on which to listen for connections.")
	httpsEnableFlag := flag.Bool("https", false, "enables https using autocert/letsencrypt.")
	httpEnableFlag := flag.Bool("http", true, "enables https using autocert/letsencrypt.")
	//firebaseCredsFlag := flag.String("firebaseCreds", "~/firebase.json", "serviceAccountKey.json for firebase.")
	flag.Parse()
	tlsCert = *tlsCertFlag
	rpcMacaroon = *rpcMacaroonFlag
	rpcServer = *rpcServerFlag
	listenPort = *listenPortFlag
	httpsEnabled := *httpsEnableFlag
	httpEnabled := *httpEnableFlag
	/*firebaseCredsFile := cleanAndExpandPath(*firebaseCredsFlag)
	opt := option.WithCredentialsFile(firebaseCredsFile)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		fatal(err)
	}
	firebaseApp = app
	firebaseDb, err = firebaseApp.Firestore(context.Background())
	if err != nil {
		fatal(err)
	}*/

	
	//watchPayments()

	
	if httpsEnabled {
	}
	/*
	if httpsEnabled {
		certManager := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist("chat-backend.rawtx.com"),
			Cache:      autocert.DirCache(filepath.Join(cleanAndExpandPath("~"), "certs")),
		}

		server := &http.Server{
			Addr: ":https",
			TLSConfig: &tls.Config{
				GetCertificate: certManager.GetCertificate,
			},
			Handler: api.MakeHandler(),
		}

		go http.ListenAndServe(":http", certManager.HTTPHandler(nil))
		log.Fatal(server.ListenAndServeTLS("", ""))
	} else {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", listenPort), api.MakeHandler()))
	}
	*/

	if httpEnabled {

		fs := http.FileServer(http.Dir("./public"))
		http.Handle("/public/", http.StripPrefix("/public/",fs))
		
		fs = http.FileServer(http.Dir("./js"))
		http.Handle("/js/", http.StripPrefix("/js", fs))

		fs = http.FileServer(http.Dir("./css"))
		http.Handle("/css/", http.StripPrefix("/css", fs))

		// Websockets
		http.HandleFunc("/ws", handleConnections)

		http.HandleFunc("/", getIndex)
		http.HandleFunc("/invoice", getInvoice)

		fileServer := http.FileServer(http.Dir("./images"))
		http.Handle("/images/", http.StripPrefix("/images", fileServer))

		go func() {
			for {
				time.Sleep(time.Second * 4)
				msg := Message{}
				
				for client := range clients {

					if(newinvoice != "") {
						payed, _ := checkPayments(newinvoice)
						if(payed == true) {
							clients[client].money = clients[client].money + 5
							msg = Message{Payed: "True", Value: strconv.Itoa(clients[client].money)}
							newinvoice = ""
						}
					}

					err := client.WriteJSON(msg)
					if err != nil {
						log.Printf("error: %v", err)
						client.Close()
						delete(clients, client)
					}
				}
			}
		}()

		go handleMessages()

		log.Println("Starting server on :8080")
		err := http.ListenAndServe(":8080", nil)
		log.Fatal(err)
	}
}
func handleMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast

		
		// Send it out to every client that is currently connected
		for client := range clients {
			
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()

	// Register our new client
	clients[ws] = &clientss{active: true, money:0}
	for {
		var msg Message
		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		if (msg.Message != ""){
			money = money - 1
		}
		if (msg.AskForInvoice == "5") {
			p,_ := loadInvoiceData(w,r,"To Lightning Chat", 5)
			fmt.Println(p.Invoice)
			msg.AskForInvoice = p.Invoice
			newinvoice = p.Invoice
		}

		// Send the newly received message to the broadcast channel
		broadcast <- msg
	}
}

func getIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Println("getIndex...")
	p, _ := loadIndexData(w,r)
	fmt.Println(p)
    t, _ := template.ParseFiles("./templates/testtemplate.html")
    t.Execute(w, p)
}
func getInvoice(w http.ResponseWriter, r *http.Request) {
	p, _ := loadInvoiceData(w,r,"To Lightning Chat", 5)
    t, _ := template.ParseFiles("templates/getInvoice.html")
    t.Execute(w, p)
}
func init() {
	tpl = template.Must(template.ParseGlob("templates/*.html"))
}
// cleanAndExpandPath expands environment variables and leading ~ in the
// passed path, cleans the result, and returns it.
func cleanAndExpandPath(path string) string {
	// Expand initial ~ to OS specific home directory.
	if strings.HasPrefix(path, "~") {
		var homeDir string
		user, err := user.Current()
		if err == nil {
			homeDir = user.HomeDir
		} else {
			homeDir = os.Getenv("HOME")
		}
		path = strings.Replace(path, "~", homeDir, 1)
	}
	return filepath.Clean(os.ExpandEnv(path))
}
