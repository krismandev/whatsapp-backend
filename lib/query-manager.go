package lib

import (
	"strconv"
	"strings"
)

func AppendWhere(base_query string, base_param []interface{}, appended_query string, appended_param string) (string, []interface{}, error) {
	var err error
	if len(appended_param) > 0 {
		if len(base_query) > 0 {
			base_query += " AND "
		}
		base_query += appended_query
		base_param = append(base_param, appended_param)
	}
	return base_query, base_param, err
}

func AppendWhereRaw(base_query string, appended_query string) (string, error) {
	var err error
	if len(appended_query) > 0 {
		if len(base_query) > 0 {
			base_query += " AND "
		}
		base_query += appended_query
	}
	return base_query, err
}

func AppendWhereLike(base_query string, base_param []interface{}, appended_query string, appended_param string) (string, []interface{}, error) {
	var err error
	if len(appended_param) > 0 {
		if len(base_query) > 0 {
			base_query += " AND "
		}
		base_query += appended_query
		base_param = append(base_param, "%"+appended_param+"%")
	}
	return base_query, base_param, err
}

func AppendOrderBy(base_query string, order_by string, order_dir string) string {
	if len(order_by) > 0 {
		base_query += " ORDER BY " + order_by
		if len(order_dir) > 0 {
			if "desc" == strings.ToLower(order_dir) {
				base_query += " DESC "
			}
		}
	}
	return base_query
}

func AppendComma(base_query string, base_param []interface{}, appended_query string, value string) (string, []interface{}) {
	if len(base_query) > 0 {
		base_query += " , "
	}

	if len(value) > 0 {
		base_query += appended_query
		base_param = append(base_param, value)
	} else {
		base_query += strings.ReplaceAll(appended_query, "?", "NULL")
	}
	return base_query, base_param
}

func AppendCommaNotNull(base_query string, base_param []interface{}, appended_query string, value string) (string, []interface{}) {
	if len(base_query) > 0 {
		base_query += " , "
	}

	base_query += appended_query
	base_param = append(base_param, value)
	return base_query, base_param
}

func AppendCommaRaw(base_query string, appended_query string) string {
	if len(appended_query) > 0 {
		if len(base_query) > 0 {
			base_query += " , "
		}
		base_query += appended_query
	}
	return base_query
}

func AppendLimit(base_query string, page int, per_page int) string {
	page = GetPageValue(page)
	per_page = GetPerPageValue(per_page)
	offset := (page - 1) * per_page
	return base_query + " LIMIT " + strconv.Itoa(offset) + ", " + strconv.Itoa(per_page)
}

func GetPerPageValue(per_page int) int {
	if per_page == 0 {
		per_page = 10
	}
	return per_page
}

func GetPageValue(page int) int {
	if page == 0 {
		page = 1
	}
	return page
}
