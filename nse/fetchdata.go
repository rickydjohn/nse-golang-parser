package fetchnse

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/nse-go/models"
)

type nsejson struct {
	Records struct {
		Data []models.Share `json:"data"`
	} `json:"records"`
}

func gunzipWrite(w io.Writer, data []byte) error {
	// Write gzipped data to the client
	gr, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	defer gr.Close()
	data, err = ioutil.ReadAll(gr)
	if err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			fmt.Printf("%v\n%v\n", err, data)
			return err
		}

	}
	w.Write(data)
	// _, err = io.Copy(w, gr)
	return nil
}

func (np *Nseparser) getfromnse(d fetchshare) ([]byte, error) {
	np.lock.RLock()
	h := np.p.Headers
	np.lock.RUnlock()
	c := http.Client{}
	c.Timeout = time.Duration(np.p.Timeout) * time.Minute
	req, err := http.NewRequest("GET", d.url, nil)
	if err != nil {
		return nil, err
	}
	for i, v := range h {
		req.Header.Add(i, v)
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	np.log.Info(d.url, " ", resp.StatusCode, " length: ", resp.ContentLength)

	for i, v := range resp.Header {
		np.p.Headers[i] = v[0]
	}

	bt, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	if err := gunzipWrite(&buffer, bt); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
