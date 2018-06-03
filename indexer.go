package je

import (
	"github.com/blevesearch/bleve"
)

// bleveIndex is an alias for Index used by the IndexBatcher to avoid a conflict
// between the embedded Index field and the overridden Index method
type bleveIndex bleve.Index
