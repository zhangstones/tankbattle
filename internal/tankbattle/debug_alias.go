package tankbattle

import gamedebug "tankbattle/internal/debugapi"

type DebugState = gamedebug.State
type DebugController = gamedebug.Controller
type debugRequest = gamedebug.Request
type debugResponse = gamedebug.Response
type debugRequestKind = gamedebug.RequestKind

const (
	debugRequestActions  = gamedebug.RequestActions
	debugRequestSnapshot = gamedebug.RequestSnapshot
	debugRequestState    = gamedebug.RequestState
)

func NewDebugController() *DebugController {
	return gamedebug.New()
}
