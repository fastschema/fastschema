package api_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestToolActivityPermission verifies the audit-trail endpoint is gated by the
// standard resource permission: root admins read it, other roles are forbidden,
// anonymous requests are unauthorized.
func TestToolActivityPermission(t *testing.T) {
	app := CreateTestApp(t)

	// Admin (root) bypasses permission checks.
	adminResp, _ := app.Get("/api/tool/activity", app.adminToken)
	app.AssertStatus(adminResp, 200)

	// A normal user without the activity permission is forbidden.
	userResp, _ := app.Get("/api/tool/activity", app.normalToken)
	app.AssertStatus(userResp, 403)

	// Anonymous request (no token) is rejected.
	guestResp, _ := app.Get("/api/tool/activity")
	assert.Contains(t, []int{401, 403}, guestResp.Code)
}
