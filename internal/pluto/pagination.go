package pluto

import (
	"net/http"
	"net/url"

	"github.com/plutoframework/pluto-core/pagination"
)

func Pagination(r *http.Request, operation string) (pagination.Request, error) {
	query, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return pagination.Request{}, pagination.ErrInvalid
	}

	p := Principal(r)

	return pagination.Parse(
		query,
		pagination.Binding(
			query,
			operation,
			p.ActorID,
			p.TenantID,
			string(p.Role),
		),
	)
}
