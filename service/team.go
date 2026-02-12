package service

import "encoding/json"

type Team struct {
	ID    string        `firestore:"id"`
	OAuth OAuthResponse `firestore:"oauth"`
}

// backfillLegacy は 2025/11/25 以前に保存された "Team"≒"OAuthResponse" を
// 新しい Team モデルへ移行するための後方互換処理。Teams コレクション上の
// 旧ドキュメントが無くなったタイミングで削除してよい。
// 今後 Team ドキュメントへ拡張フィールドを追加し、利用者が値を保存すれば
// その時点で新構造に置き換わるが、更新が行われない組織も想定されるため
// 実質的には永続的に残す可能性がある。
func (t *Team) backfillLegacy(legacydata map[string]interface{}) {
	if t == nil {
		return
	}
	// 既に新フォーマットで読み取れていれば後続処理は不要。
	if t.OAuth.AccessToken != "" {
		return
	}
	// Firestore のドキュメントに値が無い場合は互換対象が無い。
	if len(legacydata) == 0 {
		return
	}
	buf, err := json.Marshal(legacydata)
	if err != nil {
		return
	}
	// legacydata は Slack OAuthResponse に一致する構造を想定。
	oa := OAuthResponse{}
	if err := json.Unmarshal(buf, &oa); err != nil {
		return
	}
	// 旧データでもアクセストークンが無ければ Slack と連携されていない状態。
	if oa.AccessToken == "" {
		return
	}
	t.OAuth = oa
	t.ID = oa.Team.ID
}

func (t Team) SlackAccessToken() string {
	return t.OAuth.AccessToken
}
