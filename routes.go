package main

import (
	//"github.com/ant0ine/go-json-rest/rest"
	"github.com/lightningnetwork/lnd/lnrpc"
	"golang.org/x/net/context"
	qrcode "github.com/skip2/go-qrcode"
	"net/http"
	"fmt"
)
/*func showPubkey(w http.ResponseWriter, r *http.Request) {
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
}*/
type IndexPage struct {
    PubKey string
    NodeUri string
}
type getInvoicePage struct {
    Invoice string
}
func loadIndexData(w http.ResponseWriter, r *http.Request) (*IndexPage, error) {
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

    return &IndexPage{PubKey: res.GetIdentityPubkey(),NodeUri: res.GetUris()[0]}, nil
}
func loadInvoiceData(w http.ResponseWriter, r *http.Request,memo string,value int64) (*getInvoicePage, error) {
	c, clean := getClient()
	defer clean()
	res, err := c.AddInvoice(context.Background(), &lnrpc.Invoice{
		Memo:  memo,
		Value: value,
	})
	if err != nil {
		//w.WriteJson(map[string]string{"error": err.Error()})
		//return
	}
	invoicedata := res.PaymentRequest

	err = qrcode.WriteFile(invoicedata, qrcode.Medium, 256, "images/qr.png")

    return &getInvoicePage{Invoice: invoicedata}, nil
}