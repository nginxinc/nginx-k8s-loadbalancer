/*
 * Copyright 2023 F5 Inc. All rights reserved.
 * Use of this source code is governed by the Apache License that can be found in the LICENSE file.
 *
 * Factory for creating tls.Config objects based on the provided `tlsMode`.
 */

package authentication

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/nginxinc/kubernetes-nginx-ingress/internal/certification"
	"github.com/nginxinc/kubernetes-nginx-ingress/internal/configuration"
)

func NewTlsConfig(settings *configuration.Settings) (*tls.Config, error) {
	switch settings.TlsMode {
	case "ss-tls": // needs ca cert
		return buildSelfSignedTlsConfig(settings.Certificates)

	case "ss-mtls": // needs ca cert and client cert
		return buildSelfSignedMtlsConfig(settings.Certificates)

	case "ca-tls": // needs nothing
		return buildBasicTlsConfig(false), nil

	case "ca-mtls": // needs client cert
		return buildCaTlsConfig(settings.Certificates)

	default: // no-tls, needs nothing
		return buildBasicTlsConfig(true), nil
	}
}

func buildSelfSignedTlsConfig(certificates *certification.Certificates) (*tls.Config, error) {
	certPool, err := buildCaCertificatePool(certificates.GetCACertificate())
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		InsecureSkipVerify: false,
		RootCAs:            certPool,
	}, nil
}

func buildSelfSignedMtlsConfig(certificates *certification.Certificates) (*tls.Config, error) {
	certPool, err := buildCaCertificatePool(certificates.GetCACertificate())
	if err != nil {
		return nil, err
	}

	certificate, err := buildCertificates(certificates.GetClientCertificate())
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		InsecureSkipVerify: false,
		RootCAs:            certPool,
		Certificates:       []tls.Certificate{certificate},
	}, nil
}

func buildBasicTlsConfig(skipVerify bool) *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: skipVerify,
	}
}

func buildCaTlsConfig(certificates *certification.Certificates) (*tls.Config, error) {
	certificate, err := buildCertificates(certificates.GetClientCertificate())
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		InsecureSkipVerify: false,
		Certificates:       []tls.Certificate{certificate},
	}, nil
}

func buildCertificates(privateKeyPEM []byte, certificatePEM []byte) (tls.Certificate, error) {
	return tls.X509KeyPair(certificatePEM, privateKeyPEM)
}

func buildCaCertificatePool(caCert []byte) (*x509.CertPool, error) {
	block, _ := pem.Decode(caCert)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing CA certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing certificate: %w", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(cert)

	return caCertPool, nil
}
