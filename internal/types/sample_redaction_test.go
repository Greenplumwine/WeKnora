package types

import (
	"strings"
	"testing"
)

// 验证含 apikey / mobile / email 的 JSON 在移除脱敏后能原样保存。样例为合成数据，
// 但其字段值会被旧脱敏正则命中（apikey 键名、1[3-9]\d{9} 手机号），故能真正回归。
func TestSampleUserJSONSurvivesWithoutRedaction(t *testing.T) {
	sample := `{"admin":false,"apikey":"a1b2c3d4-e5f6-7890-abcd-ef1234567890","authorities":[],"bizlineCompanyId":"4","bizlineCompanyName":null,"companyId":"75","companyName":null,"createUser":1,"defaultRoles":[],"deptCompanyId":"74","deptCompanyName":null,"email":"alice@example.com","id":280,"isAdmin":false,"isEnable":true,"mobile":"13912345678","realname":"","registerTime":"2025-12-03 15:05:05","trial":false,"trialTime":null,"userType":"1","username":"alice"}`

	if got := len([]rune(sample)); got >= MemoryContentMaxRunes {
		t.Fatalf("sample rune count %d exceeds MemoryContentMaxRunes %d", got, MemoryContentMaxRunes)
	}

	sanitized := SanitizeMemoryContent(sample)
	if sanitized != sample {
		t.Fatalf("SanitizeMemoryContent mangled the sample:\nwant: %s\ngot:  %s", sample, sanitized)
	}

	// 旧脱敏会把这些字段值替换成【已隐藏】；现在必须原样保留。
	for _, needle := range []string{
		`"apikey":"a1b2c3d4-e5f6-7890-abcd-ef1234567890"`,
		`"mobile":"13912345678"`,
		`"email":"alice@example.com"`,
	} {
		if !strings.Contains(sanitized, needle) {
			t.Fatalf("sensitive field was removed/changed, missing %q in:\n%s", needle, sanitized)
		}
	}

	// 指纹链路不应丢失这些字段。
	fp := MemoryFingerprint(sample)
	if fp == "" {
		t.Fatal("MemoryFingerprint returned empty for non-empty sample")
	}
}
