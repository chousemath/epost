package main

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"fmt"

	"golang.org/x/text/encoding/korean"

	"golang.org/x/text/transform"
)

type ePostRequest struct {
	Query        string // detailed address (e.g. 뒷골2로 47-20)
	CountPerPage uint   // must be within 20~50
	CurrentPage  uint   // must be greater than or equal to 1
}

type ePostResult struct {
	XMLName  xml.Name `xml:"post"`
	PageInfo pageInfo `xml:"pageinfo"`
	ItemList itemList `xml:"itemlist"`
}

type pageInfo struct {
	XMLName      xml.Name `xml:"pageinfo"`
	TotalCount   uint     `xml:"totalCount"`
	TotalPage    uint     `xml:"totalPage"`
	CountPerPage uint     `xml:"countPerPage"`
	CurrentPage  uint     `xml:"currentPage"`
}

type itemList struct {
	XMLName xml.Name `xml:"itemlist"`
	Items   []item   `xml:"item"`
}

type item struct {
	XMLName      xml.Name `xml:"item"`
	PostalCode   string   `xml:"postcd"`
	Address      string   `xml:"address"`
	AddressJibun string   `xml:"addrjibun"`
}

var regKey = os.Getenv("EPOST_REG_KEY")

const baseURL = "http://biz.epost.go.kr/KpostPortal/openapi"

func (e ePostRequest) getPostalCodes() error {
	var bufs bytes.Buffer

	wr := transform.NewWriter(&bufs, korean.EUCKR.NewEncoder())
	wr.Write([]byte(e.Query))
	defer wr.Close()

	var sb strings.Builder
	convVal := bufs.String()
	for i := 0; i < len(convVal); i++ {
		sb.WriteString("%")
		sb.WriteString(fmt.Sprintf("%02X", convVal[i]))
	}
	res, err := http.Get(fmt.Sprintf(
		"%s?regkey=%s&target=postNew&query=%s&countPerPage=%d&currentPage=%d",
		baseURL,
		regKey,
		sb.String(),
		e.CountPerPage,
		e.CurrentPage,
	))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)

	epRes := new(ePostResult)
	err = xml.Unmarshal(data, epRes)
	if err != nil {
		return err
	}

	fmt.Println("TotalCount:", epRes.PageInfo.TotalCount)
	fmt.Println("TotalPage:", epRes.PageInfo.TotalPage)
	fmt.Println("CountPerPage:", epRes.PageInfo.CountPerPage)
	fmt.Println("CurrentPage:", epRes.PageInfo.CurrentPage)

	for _, itm := range epRes.ItemList.Items {
		fmt.Println("PostalCode:", itm.PostalCode)
		fmt.Println("Address:", itm.Address)
		fmt.Println("AddressJibun:", itm.AddressJibun)
	}

	return nil
}

func main() {
	e := ePostRequest{
		Query:        "뒷골2로 47-20",
		CountPerPage: 50,
		CurrentPage:  1,
	}
	if err := e.getPostalCodes(); err != nil {
		fmt.Println(err)
	}
}
