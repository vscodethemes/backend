-- name: GetColorCounts :many

SELECT
	t.editor_background as color,
	count(*) as count
FROM themes t
GROUP BY color
ORDER BY count DESC;