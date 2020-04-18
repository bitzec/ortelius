// (c) 2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package avm_index

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gocraft/dbr"
)

const (
	PaginationMaxLimit      = 500
	PaginationDefaultLimit  = 500
	PaginationDefaultOffset = 0

	TxSortDefault       TxSort = TxSortTimestampAsc
	TxSortTimestampAsc         = "timestamp-asc"
	TxSortTimestampDesc        = "timestamp-desc"

	queryParamKeysSortBy = "sort"
	queryParamKeysLimit  = "limit"
	queryParamKeysOffset = "offset"
	queryParamKeysSpent  = "spent"
)

var (
	ErrUndefinedSort = errors.New("undefined sort")
)

type ListParams struct {
	limit  int
	offset int
}

func ListParamForHTTPRequest(r *http.Request) (params *ListParams, err error) {
	q := r.URL.Query()
	params.limit, err = getQueryInt(q, queryParamKeysLimit, PaginationDefaultLimit)
	if err != nil {
		return nil, err
	}
	params.offset, err = getQueryInt(q, queryParamKeysOffset, PaginationDefaultOffset)
	if err != nil {
		return nil, err
	}
	return params, nil
}

func (p *ListParams) Apply(b *dbr.SelectBuilder) *dbr.SelectBuilder {
	if p.limit > PaginationMaxLimit {
		p.limit = PaginationMaxLimit
	}

	b = b.Limit(uint64(p.limit))
	b = b.Offset(uint64(p.offset))
	return b
}

type TxSort string

func ToTxSort(s string) (TxSort, error) {
	switch s {
	case TxSortTimestampAsc:
		return TxSortTimestampAsc, nil
	case TxSortTimestampDesc:
		return TxSortTimestampDesc, nil
	}
	return TxSortDefault, ErrUndefinedSort
}

type ListTxParams struct {
	*ListParams
	sort TxSort
}

func ListTxParamForHTTPRequest(r *http.Request) (*ListTxParams, error) {
	q := r.URL.Query()

	listParams, err := ListParamForHTTPRequest(r)
	if err != nil {
		return nil, err
	}

	params := &ListTxParams{
		sort:       TxSortDefault,
		ListParams: listParams,
	}

	sortBys, ok := q[queryParamKeysSortBy]
	if ok && len(sortBys) >= 1 {
		params.sort, _ = ToTxSort(sortBys[0])
	}

	return params, nil
}

func (p *ListTxParams) Apply(b *dbr.SelectBuilder) *dbr.SelectBuilder {
	if p.ListParams != nil {
		b = p.ListParams.Apply(b)
	}

	var applySort func(b *dbr.SelectBuilder, sort TxSort) *dbr.SelectBuilder
	applySort = func(b *dbr.SelectBuilder, sort TxSort) *dbr.SelectBuilder {
		switch sort {
		case TxSortTimestampAsc:
			return b.OrderAsc("avm_transactions.ingested_at")
		case TxSortTimestampDesc:
			return b.OrderDesc("avm_transactions.ingested_at")
		}
		return applySort(b, TxSortDefault)
	}
	b = applySort(b, p.sort)

	return b
}

type ListTXOParams struct {
	*ListParams
	spent *bool
}

func ListTXOParamForHTTPRequest(r *http.Request) (*ListTXOParams, error) {
	q := r.URL.Query()

	listParams, err := ListParamForHTTPRequest(r)
	if err != nil {
		return nil, err
	}

	var b *bool
	params := &ListTXOParams{
		spent:      b,
		ListParams: listParams,
	}

	spentStrs, ok := q[queryParamKeysSpent]
	if ok || len(spentStrs) >= 1 {
		b, err := strconv.ParseBool(spentStrs[0])
		if err != nil {
			return nil, err
		}
		params.spent = &b
	}

	return params, nil
}

func (p *ListTXOParams) Apply(b *dbr.SelectBuilder) *dbr.SelectBuilder {
	if p.ListParams != nil {
		b = p.ListParams.Apply(b)
	}

	if p.spent != nil {
		b = b.Where("spent = ?", *p.spent)
	}

	return b
}

func getQueryInt(q url.Values, key string, defaultVal int) (val int, err error) {
	strs, ok := q[key]
	if ok || len(strs) >= 1 {
		return strconv.Atoi(strs[0])
	}
	return defaultVal, err
}
