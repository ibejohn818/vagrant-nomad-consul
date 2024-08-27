package certs

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"time"
)

const (
  rsaKeySize = 4096
)

type CertOps struct {
  // SubjectAltName ip's
  // strings converted to ip via `net.ParseIP`
  IPs []string
  // SubjectAltNames DNS names
  DNSNames []string


}

//GenTestCA create a test CA cert
func CreateCA(ops *CertOps) (*tls.Certificate, *rsa.PrivateKey, *x509.Certificate, []byte, []byte) {
	// generate a new key-pair
	caKey, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
	if err != nil {
		log.Fatalf("generating random key: %v", err)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		panic(err)
	}

	caCert := &x509.Certificate{

		SerialNumber:          serialNumber,
		Subject:               pkix.Name{Organization: []string{"John Hardy Lab"}},
		SignatureAlgorithm:    x509.SHA256WithRSA,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		BasicConstraintsValid: true,
	}

	caCert.IsCA = true
	caCert.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature
	caCert.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}
	caCert.IPAddresses = []net.IP{net.ParseIP("127.0.0.1")}
	rootCert, rootCertPem := SignTestCert(caCert, caCert, &caKey.PublicKey, caKey)

	caKeyPem := pem.EncodeToMemory(&pem.Block{
		Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(caKey),
	})

	// Create a TLS cert using the private key and certificate
	caTlsCert, err := tls.X509KeyPair(rootCertPem, caKeyPem)
	if err != nil {
		log.Fatalf("invalid key pair: %v", err)
	}

	return &caTlsCert, caKey, rootCert, caKeyPem, rootCertPem
}

//GenTestCert create cert with server & client auth
func GenTestCert(commonName string, caCert *x509.Certificate, caPrivKey *rsa.PrivateKey) (*tls.Certificate, *x509.Certificate) {
	// create a key-pair for the server
	certKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("generating random key: %v", err)
	}

	randLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, randLimit)
	if err != nil {
		log.Fatalf("serial # error%v", err)
	}

	cert := &x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{Organization: []string{"GDDY"}},
		SignatureAlgorithm:    x509.SHA256WithRSA,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		BasicConstraintsValid: true,
	}
	cert.KeyUsage = x509.KeyUsageDigitalSignature
	//ssl and client auth
	cert.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}
	cert.IPAddresses = []net.IP{net.ParseIP("127.0.0.1")}
	cert.Subject.CommonName = commonName

	_, certPem := SignTestCert(cert, caCert, &certKey.PublicKey, caPrivKey)

	certKeyPem := pem.EncodeToMemory(&pem.Block{
		Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(certKey),
	})
	tlsCert, err := tls.X509KeyPair(certPem, certKeyPem)
	if err != nil {
		log.Fatalf("invalid key pair: %v", err)
	}

	return &tlsCert, cert
}

//SignTestCert sign and return a x509 cert
func SignTestCert(certToSign, caCert *x509.Certificate, pub interface{}, parentPriv interface{}) (*x509.Certificate, []byte) {

	var cert *x509.Certificate
	var certPem []byte

	// sha1 publicKey for use as SubjectKeyId
	pubKey, err := x509.MarshalPKIXPublicKey(pub.(*rsa.PublicKey))
	if err != nil {
		log.Fatal(err)
	}
	h := sha1.New()
	h.Write(pubKey)
	subKeyIdHash := h.Sum(nil)

	certToSign.SubjectKeyId = subKeyIdHash

	if certToSign.IsCA {
		certToSign.AuthorityKeyId = subKeyIdHash
	} else {
		certToSign.AuthorityKeyId = caCert.AuthorityKeyId
	}

	certDer, err := x509.CreateCertificate(rand.Reader, certToSign, caCert, pub, parentPriv)
	if err != nil {
		log.Fatal(err)
	}

	// parse the resulting certificate
	cert, err = x509.ParseCertificate(certDer)
	if err != nil {
		log.Fatal(err)
	}

	// PEM encode the certificate
	b := pem.Block{Type: "CERTIFICATE", Bytes: certDer}
	certPem = pem.EncodeToMemory(&b)

	return cert, certPem
}
