package api

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/examples/helloworld/helloworld"
)

const (
	Address = "localhost:56789"
)

const (
	healthCheckPath     = "/health"
	defaultTickInterval = time.Millisecond * 5
)

var (
	Certificate = []byte(`-----BEGIN CERTIFICATE-----
MIIC/jCCAeagAwIBAgIJAPGpe7jZ+B9JMA0GCSqGSIb3DQEBCwUAMBQxEjAQBgNV
BAMMCWxvY2FsaG9zdDAeFw0xOTAxMDYwMzM5MDZaFw0yOTAxMDMwMzM5MDZaMBQx
EjAQBgNVBAMMCWxvY2FsaG9zdDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoC
ggEBAMdmso6OeDGKvUvdW7n3pkht2G/QlsIf3VUcVGpM6gItoTgHmj/8I9LXAiKf
o1KVfQ7BfokI+DbKlUHfQ5Pa8thHIdOFV8XGRfJFfHxoZMGBVCwr/4NMBjBollfi
rBzqkdKsL9bQH8bmEmqJY2J5MUzg2gQaxhMCXWxDflFHpw7zmjHKaP13LSl0G3hF
zyxYPFb58k/3D2GjYy+zULXDLT9DCV9BfEXveKNJaDUy2OfmoDRDx/BGXxIdipzn
LgN4wr4gvSRPq3Jo/ElOcM7wguMgBRpCH/uKVZqLggHc9q2P+iCOZ6lqfpNNnfI8
LXqrwCFMv1K3Hx6Y4K2MwVDoVf8CAwEAAaNTMFEwHQYDVR0OBBYEFM3D5zeZf58b
UCtlLPsC3bqiNokXMB8GA1UdIwQYMBaAFM3D5zeZf58bUCtlLPsC3bqiNokXMA8G
A1UdEwEB/wQFMAMBAf8wDQYJKoZIhvcNAQELBQADggEBAEtdLAuc/RryzzGMjLub
LHjAvBp9b16Fd7k/SJbI2qq4KIOh2PUCcA944/nCrecv0csucMujE5ZnuiTnI/up
OoBAHunlxDAJfs5k62m4LSOLEZM6JcbepNIjQV/cEcf7mBzLw9BWA6KU/LenoLaM
79mDNgJniECr8r+I5hMqBLFoZqUiRrkLOgvrNmVwA3qGUPK3ZhpBEuXxrKBjFl7K
MOIqal+RwsT20zIjbKnBiFppnqIW1ntCUk9fRkd7ZHptpJnBDaqyFGV45lFen3Ai
4HBz397Xk1Tsw9wl4pmG0cedcN2GMk7LaMXnItaECmACSVTJlqPdQT8AvsZ53kgi
w9o=
-----END CERTIFICATE-----`)
	Key = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEAx2ayjo54MYq9S91bufemSG3Yb9CWwh/dVRxUakzqAi2hOAea
P/wj0tcCIp+jUpV9DsF+iQj4NsqVQd9Dk9ry2Ech04VXxcZF8kV8fGhkwYFULCv/
g0wGMGiWV+KsHOqR0qwv1tAfxuYSaoljYnkxTODaBBrGEwJdbEN+UUenDvOaMcpo
/XctKXQbeEXPLFg8VvnyT/cPYaNjL7NQtcMtP0MJX0F8Re94o0loNTLY5+agNEPH
8EZfEh2KnOcuA3jCviC9JE+rcmj8SU5wzvCC4yAFGkIf+4pVmouCAdz2rY/6II5n
qWp+k02d8jwteqvAIUy/UrcfHpjgrYzBUOhV/wIDAQABAoIBAQCZj859GN0ZkjY8
AapNappFd0rSuboQoAeNLzcXckpZCRj6lGhHVH+mNO0xCu31gKiBv6QaFq1JTPRr
eWyKpniU9Rro0e0Jo6tka/z1tlO57kaLigrJ67dsem8mGavgzQkmTHK/JSMDw1V1
dH70bE76XMOpm5DlPNIDuWrDX8IZMUcbPqTMvrRUphqgBYOV9x32Jnvc36BzTw88
corWZ7RINXSlgvz6K27HtVHbJL7orJv8v2hNMue5chtbZy2oag8mHQgNFSxn6/0T
wQ32Ms87X5xZTkpGv+u3kAFCv0u1YirLxnxVpgIN7eBG/RN3/fh1xJkDmjv/qwqs
6trRvcOZAoGBAO4FaTPMpaGefpX7rdxmZBH7un5rKS+iYBm4RIbThViBTNdA9LZp
JtwduIVwD40tLbQ+tOB65jb50iVMfE61rW8xchzOHYi13SaPgEjAXXysWLf+fVsy
C+1XqUHupFqvSHjFmgkC9Rcgu/6u6fri7N5sVlrt0llqvY0+VdREAadVAoGBANZ2
fwrXPEohMi3Aerv9qIyWuPKQyRRBS/fFUqllfhuXW74KcF4EoQCxOiSXBKjK1tgu
SfuJTaYCzzf7Knw3DpIs5HEVVMbG1Kgfbn7wDQTk7xloYub/vsgRZo9ez0HM+lxq
I62B4Y9/g3RRMmcTHXDWCw633ms29AAiMf32leADAoGBAJdYyXQuhIMoDMXBquOi
F693qTYJXb70OLch/DDe/sMwNHQK0Y/LfPIp09LFVp4mRBGAbfLvMsNyRrWA1OoX
i5hQkIbQaOcs/NowFRotd0R3MlKMd5ktUXgxbWaHH+qp2iMxQqjIQJ/cKK3g+taU
xJkJuj9HSaGhxbWyFVFLjOGhAoGALdsmbO36sSsJ7Kh0Vc/2AyGTKCJ3LEKN+MuT
Ui8mWMXzUt4uipvYxSof8YTs9R5x88VqAkOoe6+sGR82RVsMXYsFyXwzJVGMVOpr
mO7BCePdkAQ26YeThnnaARvXmw02Fx6GxGm6DhHIzM0zxsBaki7iLGJ6R1h3sbOe
FtxrzXsCgYEA0S0lYtmv4QOQq3mLf2lfroDCirqR6f5iXsyd3KhnkH/mnm56Bv4R
gTtaIFtkFS/s+lwNYSfQC/3uVPAbGcrZKRtmh9EaJk+jnjt72TRjldBnpxkXxGlz
svOxtwq93kqwiXrvckgvX+F/8D0fUpL0TyI5NY2qhPIT0tHW/yzSWiw=
-----END RSA PRIVATE KEY-----`)
)

type Server struct {
	lock      *sync.Mutex
	server    *http.Server
	tlsConfig *tls.Config
}

func NewServer() *Server {
	grpcServer := grpc.NewServer()
	helloworld.RegisterGreeterServer(grpcServer, greeterServer{})

	httpServer := &http.Server{
		Handler: h2c.NewHandler(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			if strings.HasPrefix(req.Header.Get("Content-Type"), "application/grpc") {
				if req.ProtoMajor == 2 {
					grpcServer.ServeHTTP(res, req)
				} else {
					res.WriteHeader(http.StatusUpgradeRequired)
				}
			} else if req.URL.Path == healthCheckPath {
				res.WriteHeader(http.StatusOK)
			} else {
				res.WriteHeader(http.StatusNotFound)
			}
		}), &http2.Server{}),
	}

	return &Server{
		lock:   &sync.Mutex{},
		server: httpServer,
	}
}

func (s *Server) Serve() error {
	s.lock.Lock()

	l, err := net.Listen("tcp", Address)
	if err != nil {
		return err
	}

	s.lock.Unlock()

	return s.server.Serve(l)
}

func (s *Server) ServeTLS() error {
	s.lock.Lock()

	cert, err := tls.X509KeyPair(Certificate, Key)
	if err != nil {
		return err
	}

	s.tlsConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	l, err := tls.Listen("tcp", Address, s.tlsConfig)
	if err != nil {
		return err
	}

	s.lock.Unlock()

	return s.server.Serve(l)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) TLSConfig() *tls.Config {
	return s.tlsConfig
}

func (s *Server) BlockUntilRunning(timeout time.Duration) error {
	interval := defaultTickInterval
	if timeout < interval {
		interval = timeout
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	done := make(chan bool)
	timer := time.AfterFunc(timeout, func() {
		done <- true
	})

	for {
		select {
		case <-done:
			return errors.New("timed out before server passed health check")
		case <-ticker.C:
			protocol := "http"
			client := http.DefaultClient

			if s.tlsConfig != nil {
				protocol = "https"

				certPool := x509.NewCertPool()
				if ok := certPool.AppendCertsFromPEM(Certificate); !ok {
					return errors.New("failed to create root CA pool")
				}

				client = &http.Client{
					Transport: &http.Transport{
						TLSClientConfig: &tls.Config{
							RootCAs: certPool,
						},
					},
				}
			}

			u := fmt.Sprintf("%s://%s", protocol, path.Join(Address, healthCheckPath))
			res, err := client.Get(u)
			if err != nil {
				return err
			}
			if res.StatusCode == http.StatusOK {
				timer.Stop()
				return nil
			}
		}
	}
}

type greeterServer struct{}

func (s greeterServer) SayHello(ctx context.Context, req *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	return &helloworld.HelloReply{
		Message: fmt.Sprintf("Hello %s", req.Name),
	}, nil
}
