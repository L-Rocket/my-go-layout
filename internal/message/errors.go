package message

import "errors"

// Session-existence and lifecycle failures propagate from session.Service
// (session.ErrNotFound / session.ErrAlreadyEnded); only the ownership rule
// is re-stated here so this module's contract reads on its own.
var ErrNotOwner = errors.New("message: not session owner")
