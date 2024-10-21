-- name: GetColorCounts :many

SELECT
	t.editor_background as color,
	count(*) as count
FROM themes t
GROUP BY color
ORDER BY count DESC;

-- name: DeleteExtensionThemesNotIn :exec

DELETE FROM themes t
WHERE t.extension_id = @extension_id 
AND t.id != ALL(@theme_ids::bigint[]);

