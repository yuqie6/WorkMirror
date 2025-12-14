package service

import "github.com/yuqie6/WorkMirror/internal/schema"

const (
	sessionMetaDiffIDs         = "diff_ids"
	sessionMetaBrowserEventIDs = "browser_event_ids"
	sessionMetaSkillKeys       = "skill_keys"
	sessionMetaTopDomains      = "top_domains"
	sessionMetaRAGRefs         = "rag_refs"
	sessionMetaTags            = "tags"
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
