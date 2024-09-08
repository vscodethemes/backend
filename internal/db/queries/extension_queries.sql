-- name: GetExtension :one
SELECT 
	e.name,
	e.display_name,
	e.publisher_name,
	e.publisher_display_name,
	e.short_description,
	jsonb_agg(json_build_object(
		'name', t.name,
		'display_name', t.display_name,
		'url', i.url
	)) AS themes
FROM extensions e
LEFT JOIN themes t ON t.extension_id = e.id
LEFT JOIN images i ON i.theme_id = t.id
WHERE 
	e.name = @extension_name
	AND e.publisher_name = @publisher_name
	AND i.language = @language
GROUP BY e.id;

