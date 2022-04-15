package certs

import (
	"crypto/x509"
	logging "fabric-sdk-go/common/logger"
	"github.com/pkg/errors"
	"time"
)

var logger = logging.NewLogger("fabsdk/certs")

//VerifyCertificate verifies raw certs and chain certs for expiry and not yet valid dates
func VerifyCertificate(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	for _, c := range rawCerts {
		cert, err := x509.ParseCertificate(c)
		if err != nil {
			logger.Warn("Got error while verifying cert")
		}
		if cert != nil {
			err = ValidateCertificateDates(cert)
			if err != nil {
				//cert is expired or not valid
				logger.Warn(err.Error())
				return err
			}
		}
	}
	for _, certs := range verifiedChains {
		for _, cert := range certs {
			err := ValidateCertificateDates(cert)
			if err != nil {
				//cert is expired or not valid
				logger.Warn(err.Error())
				return err
			}
		}
	}
	return nil
}

//ValidateCertificateDates used to verify if certificate was expired or not valid until later date
func ValidateCertificateDates(cert *x509.Certificate) error {
	if cert == nil {
		return nil
	}
	if time.Now().UTC().Before(cert.NotBefore) {
		return errors.New("Certificate provided is not valid until later date")
	}

	if time.Now().UTC().After(cert.NotAfter) {
		return errors.New("Certificate provided has expired")
	}
	return nil
}
