package bongo

import (
	"labix.org/v2/mgo"
	"math"
)

type ResultSet struct {
	Query      *mgo.Query
	Iter       *mgo.Iter
	loadedIter bool
	Connection *Connection
}

type PaginationInfo struct {
	Current      int
	TotalPages   int
	PerPage      int
	TotalRecords int
}

func (r *ResultSet) Next(mod interface{}) bool {
	colname := getCollectionName(mod)
	returnMap := make(map[string]interface{})

	// Check if the iter has been instantiated yet
	if !r.loadedIter {
		r.Iter = r.Query.Iter()
		r.loadedIter = true
	}

	gotResult := r.Iter.Next(returnMap)

	if gotResult {
		DecryptDocument(r.Connection.GetEncryptionKey(colname), returnMap, mod)
		return true
	}
	return false
}

func (r *ResultSet) Free() error {
	if r.loadedIter {
		if err := r.Iter.Close(); err != nil {
			return err
		}
	}

	return nil
}

// Set skip + limit on the current query and generates a PaginationInfo struct with info for your front end
func (r *ResultSet) Paginate(perPage, page int) (*PaginationInfo, error) {

	info := new(PaginationInfo)

	// Get count of current query
	count, err := r.Query.Count()

	if err != nil {
		return info, err
	}

	// Calculate how many pages
	totalPages := int(math.Ceil(float64(count) / float64(perPage)))

	if page < 1 {
		page = 1
	} else if page > totalPages {
		page = totalPages
	}

	skip := (page - 1) * perPage

	r.Query.Skip(skip).Limit(perPage)

	info.TotalPages = totalPages
	info.PerPage = perPage
	info.Current = page
	info.TotalRecords = count

	return info, nil
}

// Pass in the sample just so we can get the collection name
func (c *Connection) Find(query interface{}, collection interface{}) *ResultSet {
	// If collection is a string, assume that's the collection name
	var colname string
	if str, ok := collection.(string); ok {
		colname = str
	} else {
		colname = getCollectionName(collection)
	}

	q := c.Collection(colname).Collection().Find(query)

	resultset := new(ResultSet)

	resultset.Query = q
	resultset.Connection = c

	return resultset
}