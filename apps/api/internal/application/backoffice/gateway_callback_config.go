package backoffice

import (
	"errors"
	"net"
	"net/url"
	"strings"
)

var ErrGatewayCallbackUnavailable = errors.New("gateway_callback_not_configured")

// The route belongs to the application, never to administrator-entered gateway
// settings. This step defines the URL; no callback HTTP handler exists yet.
const ZarinpalCallbackPath = "/storefront/payments/zarinpal/callback"

func (s *Service) ConfigureGatewayCallbackBase(base string, allowLocalHTTP bool) error {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if base != "" {
		if _, err := buildGatewayCallbackURL("zarinpal", base, allowLocalHTTP); err != nil {
			return err
		}
	}
	s.gatewayPublicBaseURL, s.gatewayLocalCallback = base, allowLocalHTTP
	return nil
}

func (s *Service) GatewayCallbackURL(provider string) (string, error) {
	return buildGatewayCallbackURL(provider, s.gatewayPublicBaseURL, s.gatewayLocalCallback)
}

func buildGatewayCallbackURL(provider, base string, allowLocalHTTP bool) (string, error) {
	if provider != "zarinpal" {
		return "", ErrInvalid
	}
	if base == "" {
		return "", ErrGatewayCallbackUnavailable
	}
	u, err := url.Parse(base)
	if err != nil || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || (u.Path != "" && u.Path != "/") || len(base) > 2048 {
		return "", ErrGatewayCallbackUnavailable
	}
	loopback := u.Hostname() == "localhost"
	if ip := net.ParseIP(u.Hostname()); ip != nil {
		loopback = ip.IsLoopback()
	}
	if u.Scheme != "https" && !(allowLocalHTTP && loopback && u.Scheme == "http") {
		return "", ErrGatewayCallbackUnavailable
	}
	u.Path = ZarinpalCallbackPath
	return u.String(), nil
}
