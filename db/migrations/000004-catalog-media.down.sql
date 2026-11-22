-- Restore references to inline bytes before removing this storage table.
UPDATE review_workspace w
SET payload = jsonb_set(w.payload, '{icon_media}', COALESCE((
    SELECT jsonb_agg(CASE WHEN b.id IS NOT NULL
        THEN (m - 'stored') || jsonb_build_object('data', encode(b.data, 'base64'))
        ELSE m END)
    FROM jsonb_array_elements(COALESCE(w.payload->'icon_media','[]'::jsonb)) m
    LEFT JOIN catalog_media b ON b.id = m->>'id'
), '[]'::jsonb));
DROP TABLE catalog_media;
