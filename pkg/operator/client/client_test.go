package client

import (
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	apis "github.com/fabedge/fabedge/pkg/apis/v1alpha1"
	"github.com/fabedge/fabedge/pkg/operator/apiserver"
	certutil "github.com/fabedge/fabedge/pkg/util/cert"
)

func TestGetCertificate(t *testing.T) {
	mux, url, teardown := newServer()
	defer teardown()

	certManager, _ := newCertManager()
	expectedCert := certManager.GetCACert()

	var req *http.Request
	mux.HandleFunc(apiserver.URLGetCA, func(w http.ResponseWriter, r *http.Request) {
		req = r
		w.Write(certManager.GetCACertPEM())
	})

	cert2, err := GetCertificate(url)

	g := NewGomegaWithT(t)
	g.Expect(err).Should(BeNil())
	g.Expect(req.Method).Should(Equal(http.MethodGet))
	g.Expect(*cert2.Raw).Should(Equal(*expectedCert))
}

func TestSignCertByToken(t *testing.T) {
	g := NewGomegaWithT(t)
	certManager, certPool := newCertManager()
	mux, url, teardown := newServer()
	defer teardown()

	var requestContent []byte
	var receivedToken string
	var req *http.Request
	mux.HandleFunc(apiserver.URLSignCERT, func(w http.ResponseWriter, r *http.Request) {
		req = r
		receivedToken = r.Header.Get(apiserver.HeaderAuthorization)[7:]
		requestContent, _ = ioutil.ReadAll(r.Body)

		csr, _ := certutil.DecodePEM(requestContent)
		certDER, _ := certManager.SignCert(csr)

		w.WriteHeader(http.StatusOK)
		w.Write(certutil.EncodeCertPEM(certDER))
	})

	keyDER, csr, _ := certutil.NewCertRequest(certutil.Request{CommonName: "test"})
	privateKey, _ := x509.ParsePKCS1PrivateKey(keyDER)
	csrPEM := certutil.EncodeCertRequestPEM(csr)

	token := "123456"
	cert, err := SignCertByToken(url, token, csr, certPool)

	g.Expect(err).Should(BeNil())
	g.Expect(req.Method).Should(Equal(http.MethodPost))
	g.Expect(cert.Raw.Subject.CommonName).Should(Equal("test"))
	g.Expect(cert.Raw.PublicKey).Should(Equal(privateKey.Public()))
	g.Expect(receivedToken).Should(Equal(token))
	g.Expect(requestContent).Should(Equal(csrPEM))
}

func TestClient_SignCert(t *testing.T) {
	g := NewGomegaWithT(t)
	certManager, _ := newCertManager()
	mux, url, teardown := newServer()
	defer teardown()

	var requestContent []byte
	var req *http.Request
	mux.HandleFunc(apiserver.URLSignCERT, func(w http.ResponseWriter, r *http.Request) {
		req = r
		requestContent, _ = ioutil.ReadAll(r.Body)

		csr, _ := certutil.DecodePEM(requestContent)
		certDER, _ := certManager.SignCert(csr)

		w.WriteHeader(http.StatusOK)
		w.Write(certutil.EncodeCertPEM(certDER))
	})

	cli, err := NewClient(url, nil)
	g.Expect(err).Should(BeNil())

	keyDER, csr, _ := certutil.NewCertRequest(certutil.Request{CommonName: "test"})
	privateKey, _ := x509.ParsePKCS1PrivateKey(keyDER)
	csrPEM := certutil.EncodeCertRequestPEM(csr)

	cert, err := cli.SignCert(csr)
	g.Expect(err).Should(BeNil())
	g.Expect(req.Method).Should(Equal(http.MethodPost))
	g.Expect(cert.Raw.Subject.CommonName).Should(Equal("test"))
	g.Expect(cert.Raw.PublicKey).Should(Equal(privateKey.Public()))
	g.Expect(requestContent).Should(Equal(csrPEM))
}

func TestClient_UpdateEndpoints(t *testing.T) {
	g := NewGomegaWithT(t)
	mux, url, teardown := newServer()
	defer teardown()

	var receivedEndpoints []apis.Endpoint
	var req *http.Request
	mux.HandleFunc(apiserver.URLUpdateEndpoints, func(w http.ResponseWriter, r *http.Request) {
		req = r
		content, _ := ioutil.ReadAll(r.Body)
		_ = json.Unmarshal(content, &receivedEndpoints)

		w.WriteHeader(http.StatusNoContent)
	})

	cli, err := NewClient(url, nil)
	g.Expect(err).Should(BeNil())

	endpoints := []apis.Endpoint{
		{
			Name:            "connector",
			PublicAddresses: []string{"connector"},
			Subnets:         []string{"2.2.0.0/24"},
			NodeSubnets:     []string{"10.10.10.1/32"},
		},
	}
	g.Expect(cli.UpdateEndpoints(endpoints)).Should(Succeed())
	g.Expect(req.Method).Should(Equal(http.MethodPut))
	g.Expect(receivedEndpoints).Should(Equal(endpoints))
}

func newServer() (mux *http.ServeMux, url string, close func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)
	return mux, server.URL, server.Close
}

func newCertManager() (certutil.Manager, *x509.CertPool) {
	certDER, keyDER, _ := certutil.NewSelfSignedCA(certutil.Config{CommonName: "CA"})
	manager, _ := certutil.NewManger(certDER, keyDER, time.Hour)

	pool := x509.NewCertPool()
	pool.AddCert(manager.GetCACert())

	return manager, pool
}