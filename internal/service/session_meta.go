package service

import "github.com/yuqie6/WorkMirror/internal/schema"

const (
	sessionMetaDiffIDs         = "diff_ids"
	sessionMetaBrowserEventIDs = "browser_event_ids"
	sessionMetaSkillKeys       = "skill_keys"
	sessionMetaTopDomains      = "top_domains"
	sessionMetaRAGRefs         = "rag_refs"
	sessionMetaTags            = "tags"

	sessionMetaSemanticSource  = "semantic_source"  // ai | rule
	sessionMetaSemanticVersion = "semantic_version" // e.g. "v1"
	sessionMetaEvidenceHint    = "evidence_hint"    // diff+browser | diff | browser | window_only | diff_only | browser_only
	sessionMetaDegradedReason  = "degraded_reason"  // not_configured | provider_error | rate_limited | ...
)

func getSessionDiffIDs(meta schema.JSONMap) []int64 {
	return schema.GetInt64Slice(meta, sessionMetaDiffIDs)
}

func setSessionDiffIDs(meta schema.JSONMap, ids []int64) {
	schema.SetInt64Slice(meta, sessionMetaDiffIDs, ids)
}

func getSessionBrowserEventIDs(meta schema.JSONMap) []int64 {
	return schema.GetInt64Slice(meta, sessionMetaBrowserEventIDs)
}

func setSessionBrowserEventIDs(meta schema.JSONMap, ids []int64) {
	schema.SetInt64Slice(meta, sessionMetaBrowserEventIDs, ids)
}

func getSessionMetaString(meta schema.JSONMap, key string) string {
	if meta == nil {
		return ""
	}
	raw, ok := meta[key]
	if !ok || raw == nil {
		return ""
	}
	s, _ := raw.(string)
	return s
}

func setSessionMetaString(meta schema.JSONMap, key, value string) {
	if meta == nil {
		return
	}
	if value == "" {
		delete(meta, key)
		return
	}
	meta[key] = value
}
