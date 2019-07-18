// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package api

import (
	"github.com/pegnet/pegnet/opr"
)

// ItemsPerPage is how many items are returnedl
var ItemsPerPage uint32 = 20

// Pagination holds all the metadata for paged queries
type Pagination struct {
	Page         uint32    `json:"page,omitempty"`
	PageCount    uint32    `json:"page_count,omitempty"`
	ItemsPerPage uint32    `json:"items_per_page,omitempty"`
	TotalCount   uint32    `json:"total_count,omitempty"`
}

type PagedBlocks struct {
	OPRBlocks   []*opr.OprBlock  `json:"oprblocks,omitempty"`
	MetaData    Pagination `      json:"pagination,omitempty"`
}

type PagedOPRs struct {
	OPRs     []opr.OraclePriceRecord   `json:"oprs,omitempty"`
	MetaData  Pagination               `json:"pagination,omitempty"`
}

func paginateBlocks(page uint32,
										perPage uint32, 
										blocks []*opr.OprBlock ) PagedBlocks {
	totalCount := uint32(len(blocks))
	pageCount := (totalCount + perPage -1) / perPage
	if page > pageCount {
		page = pageCount
	}
	offset := (page -1) * perPage
	end := (page * perPage) + 1
	if end > totalCount {
		end = totalCount
	}
	return PagedBlocks {
		OPRBlocks: blocks[offset: end],
		MetaData: Pagination {
			Page: page,
			PageCount: pageCount,
			ItemsPerPage: perPage,
			TotalCount: totalCount,
		},
	}
}

func paginateOPRs(page uint32,
									perPage uint32, 
									blocks []opr.OraclePriceRecord ) PagedOPRs {
	totalCount := uint32(len(blocks))
	pageCount := (totalCount + perPage -1) / perPage
	if page > pageCount {
		page = pageCount
	}
	offset := (page -1) * perPage
	end := (page * perPage) + 1
	if end > totalCount {
		end = totalCount
	}
	return PagedOPRs {
		OPRs: blocks[offset: end],
		MetaData: Pagination {
			Page: page,
			PageCount: pageCount,
			ItemsPerPage: perPage,
			TotalCount: totalCount,
		},
	}
}