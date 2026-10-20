package usecases

import (
	"bytes"
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/url"
	"strings"
)

const MaxSponsorLogoBytes = 1024 * 1024

var ErrInvalidSponsorBrand = errors.New("provide an HTTP(S) website and a valid PNG or JPEG logo up to 1 MB and 4096 × 4096 pixels")

func validateSponsorshipBrand(website string, logo []byte) (string, string, error) {
	website = strings.TrimSpace(website)
	u, err := url.ParseRequestURI(website)
	if err != nil || len(website) > 2048 || u.Hostname() == "" || u.User != nil || (u.Scheme != "https" && u.Scheme != "http") {
		return "", "", ErrInvalidSponsorBrand
	}
	if len(logo) == 0 || len(logo) > MaxSponsorLogoBytes {
		return "", "", ErrInvalidSponsorBrand
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(logo))
	if err != nil || (format != "png" && format != "jpeg") || config.Width <= 0 || config.Height <= 0 || config.Width > 4096 || config.Height > 4096 {
		return "", "", ErrInvalidSponsorBrand
	}
	if _, _, err := image.Decode(bytes.NewReader(logo)); err != nil {
		return "", "", ErrInvalidSponsorBrand
	}
	return website, "image/" + format, nil
}
