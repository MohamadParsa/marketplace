package kavenegar

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cherriva/marketplace/apps/api/internal/application/operationlog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var ErrProvider = errors.New("kavenegar provider rejected request")

type Notifier struct {
	apiKey, template, baseURL string
	client                    *http.Client
}

func New(apiKey, template, baseURL string, client *http.Client) *Notifier {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &Notifier{apiKey: apiKey, template: template, baseURL: strings.TrimRight(baseURL, "/"), client: client}
}

type response struct {
	Return struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
	} `json:"return"`
}

func (n *Notifier) Send(ctx context.Context, _ identity.Surface, phone, code string) (err error) {
	finish := operationlog.BeginExternal(ctx, "kavenegar", "otp.send", 1)
	status := 0
	providerCode := ""
	defer func() { finish(status, providerCode, err) }()
	if n.apiKey == "" || n.template == "" || n.baseURL == "" {
		return ErrProvider
	}
	endpoint := fmt.Sprintf("%s/v1/%s/verify/lookup.json", n.baseURL, url.PathEscape(n.apiKey))
	form := url.Values{"receptor": {identity.NormalizePhone(phone)}, "token": {code}, "template": {n.template}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := n.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	status = res.StatusCode
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("%w: http %d", ErrProvider, res.StatusCode)
	}
	var out response
	if err := json.Unmarshal(body, &out); err != nil {
		return fmt.Errorf("%w: invalid response", ErrProvider)
	}
	if out.Return.Status != 200 {
		providerCode = "provider." + strconv.Itoa(out.Return.Status)
		return fmt.Errorf("%w: provider status %d", ErrProvider, out.Return.Status)
	}
	return nil
}
