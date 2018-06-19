package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/lightningnetwork/lnd/lnrpc"
	"golang.org/x/net/context"
	qrcode "github.com/skip2/go-qrcode"
	"net/http"
	"fmt"
	//"io/ioutil"
	"reflect"
)

func getInvoice22(w rest.ResponseWriter, r *rest.Request) {
	c, clean := getClient()
	defer clean()

	memo := r.PathParam("memo")
	res, err := c.AddInvoice(context.Background(), &lnrpc.Invoice{
		Memo:  memo,
		Value: 100,
	})
	if err != nil {
		w.WriteJson(map[string]string{"error": err.Error()})
		return
	}
	w.WriteJson(map[string]string{"pay_req": res.PaymentRequest})
}

func getPubkey(w rest.ResponseWriter, r *rest.Request) {
	c, clean := getClient()
	defer clean()

	res, err := c.GetInfo(context.Background(), &lnrpc.GetInfoRequest{})
	if err != nil {
		w.WriteJson(map[string]string{"error": err.Error()})
		return
	}
	j := map[string]string{"pubkey": res.GetIdentityPubkey()}
	if len(res.GetUris()) > 0 {
		j["uri"] = res.GetUris()[0]
	}
	w.WriteJson(j)
	//var png []byte
	//png, err = qrcode.Encode("https://example.org", qrcode.Medium, 256)
  	//png, err := qrcode.Encode("https://example.org", qrcode.Medium, 256)
  	//w.WriteJson("uri:")
	err = qrcode.WriteFile(res.GetUris()[0], qrcode.Medium, 256, "qr.png")
}
func TodoIndex(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/html; charset=UTF-8")
    w.WriteHeader(http.StatusOK)
    //if err := json.NewEncoder(w).Encode(todos); err != nil {
    //    panic(err)
    //}
}
func showPubkey(w http.ResponseWriter, r *http.Request) {
	c, clean := getClient()
	defer clean()

	res, err := c.GetInfo(context.Background(), &lnrpc.GetInfoRequest{})
	if err != nil {
		//fmt.Fprintf(w, err.Error())
		return
	}
	//j := map[string]string{"pubkey": res.GetIdentityPubkey()}
	//if len(res.GetUris()) > 0 {
	//	j = res.GetUris()[0]
	//}

	//w.Header().Set("Content-Type", "application/html; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	//file, err := os.Open("files/not-found.html")
	http.ServeFile(w, r, "index.html")
	fmt.Fprintf(w, "<h1>%s</h1><h1>%s</h1>", res.GetIdentityPubkey(), res.GetUris()[0])
}
type Page struct {
    Title string
    Body  []byte
    Imageurl string
}
type IndexPage struct {
    PubKey string
    NodeUri string
}
type getInvoicePage struct {
    Invoice string
}
func loadIndexData(title string,w http.ResponseWriter, r *http.Request) (*IndexPage, error) {
	c, clean := getClient()
	defer clean()

	res, err := c.GetInfo(context.Background(), &lnrpc.GetInfoRequest{})
	if err != nil {
		fmt.Fprintf(w, err.Error())
		//return
	}

    /*if false {
 	   filename := title + ".txt"
    	body, err := ioutil.ReadFile(filename)
	    imageur := "images/qr.png"
	    if err != nil {
	        return nil, err
	    }
	    //return &Page{Title: title, Body: body, Imageurl: imageur}, nil
	}
	*/
	nodeuri := res.GetUris()[0]
	fmt.Println(nodeuri)

    return &IndexPage{PubKey: res.GetIdentityPubkey(),NodeUri: res.GetUris()[0]}, nil
}
func loadInvoiceData(title string,w http.ResponseWriter, r *http.Request) (*getInvoicePage, error) {
	c, clean := getClient()
	defer clean()
	res, err := c.AddInvoice(context.Background(), &lnrpc.Invoice{
		Memo:  "sda",
		Value: 100,
	})
	if err != nil {
		//w.WriteJson(map[string]string{"error": err.Error()})
		//return
	}
	invoicedata := res.PaymentRequest
	invoicedata_type := reflect.TypeOf(res.PaymentRequest).Kind()
	fmt.Println(invoicedata,invoicedata_type)

    return &getInvoicePage{Invoice: res.PaymentRequest}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
    title := r.URL.Path[len("/view/"):]
    fmt.Println(title,"---")
    /*
    p, _ := loadPage(title)
    t, _ := template.ParseFiles("index.html")
    t.Execute(w, p)
    */
}
