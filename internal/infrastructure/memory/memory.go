// Package memory provides in-memory implementations of the domain Store
// interfaces. It exists so the whole binary runs locally with zero
// external services; main selects it unless Supabase credentials are set.
package memory

import (
	"sort"

	"github.com/L-Rocket/my-go-layout/internal/session"
)

// sortSessions backs the Store contract's "oldest first" promise — map
// iteration order is random, so List* sorts by creation time.
func sortSessions(ss []*session.Session) {
	sort.Slice(ss, func(i, j int) bool { return ss[i].CreatedAt.Before(ss[j].CreatedAt) })
}
