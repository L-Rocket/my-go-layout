// Package memory provides in-memory implementations of the domain Store
// interfaces. It exists so the whole binary runs locally with zero
// external services; main selects it unless Supabase credentials are set.
package memory

import (
	"sort"

	"github.com/L-Rocket/my-go-layout/internal/session"
	"github.com/L-Rocket/my-go-layout/internal/user"
)

// Shared ordering helpers: the Store contracts promise "oldest first",
// and map iteration is random, so every List* sorts by creation time.

func sortSessions(ss []*session.Session) {
	sort.Slice(ss, func(i, j int) bool { return ss[i].CreatedAt.Before(ss[j].CreatedAt) })
}

func sortUsers(us []*user.User) {
	sort.Slice(us, func(i, j int) bool { return us[i].CreatedAt.Before(us[j].CreatedAt) })
}
