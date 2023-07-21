package http

const (
	ErrorServerSuck       = "SERVER_WAS_SUCK"
	ErrorUnknownProject   = "UNKNOWN_PROJECT_ID"
	ErrorBranchNotChanged = "BRANCH_NOT_CHANGED"
	ErrorBranchNotAllowed = "BRANCH_NOT_ALLOWED"
	ErrorRevisionExists   = "REVISION_ALREADY_EXISTS"
	ErrorDataNotFound     = "DATA_NOT_FOUND"
)

type ResponseMeta struct {
	Offset    int    `json:"offset,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	Total     int    `json:"total,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
}

type Response struct {
	Meta    ResponseMeta `json:"_meta,omitempty"`
	Payload interface{}  `json:"payload"`
}
