package optimizely

import (
	"bytes"
	"text/template"

	"github.com/optimizely/go-sdk/pkg/utils"
	"github.com/rs/zerolog/log"
)

type TemplateListener struct {
	tpl       *template.Template
	requester *utils.HTTPRequester
	headers   []utils.Header
	url       string
}

func NewTemplateListener(requester *utils.HTTPRequester, filename string, url string, overrideHeaders map[string]string) *TemplateListener {
	tpl, err := template.ParseFiles(filename)
	if err != nil {
		log.Error().Err(err).Msg("error loading template")
		return &TemplateListener{}
	}

	defaultHeaders := map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	}

	for k, v := range overrideHeaders {
		defaultHeaders[k] = v
	}

	headers := make([]utils.Header, 0, len(defaultHeaders)+len(overrideHeaders))
	for k, v := range defaultHeaders {
		log.Debug().Str("HeaderName", k).Str("HeaderValue", v).Msg("append header")
		headers = append(headers, utils.Header{Name: k, Value: v})
	}

	return &TemplateListener{tpl: tpl, requester: requester, url: url, headers: headers}

}

func (l *TemplateListener) parse(message *Message) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	err := l.tpl.Execute(buf, message)
	if err != nil {
		return buf, err
	}

	return buf, err
}

// Be careful to not create an infinite loop :)
func (l *TemplateListener) Listen(message interface{}) {
	body, err := l.parse(&Message{Message: message, Env: Env})
	if err != nil {
		log.Info().Err(err).Msg("error parsing request")
	}

	log.Debug().Msg(body.String())
	log.Debug().Msg("triggering listener")

	res, _, code, err := l.requester.Do(l.url, "POST", body, l.headers)
	log.Debug().Bytes("res", res).Int("code", code).Msg("listener response")
	if err != nil {
		log.Info().Err(err).Msg("error submitting request")
	}
}
