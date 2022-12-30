package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
)

func main() {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "receipt", "receipt.jpg"))
	h.Set("Content-Type", "image/jpeg")

	fw, err := writer.CreatePart(h)
	if err != nil {
		panic(err)
	}

	file, err := os.Open("/Users/manzanit0/Documents/aeron-receipt.jpg")
	if err != nil {
		panic(err)
	}

	data, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	_, err = fw.Write(data)
	if err != nil {
		panic(err)
	}

	// Finish constructing the multipart request body
	err = writer.Close()
	if err != nil {
		panic(err)
	}

	url := "https://invx-production.up.railway.app/api/receipts/parse"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body.Bytes()))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Content-Length", fmt.Sprintf("%d", body.Len()))
	req.Header.Set("Authorization", "mcduck")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()
	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(res)
	fmt.Println(string(respBody))
}
