-- Materialize the already-sanitized read model once instead of aggregating it per cache miss.
ALTER TABLE catalog_public_state ADD COLUMN full_payload jsonb;
UPDATE catalog_public_state s SET full_payload=s.payload || jsonb_build_object(
 'products',coalesce((SELECT jsonb_agg(p.payload->'products'->0 ORDER BY p.id) FROM catalog_public_products p),'[]'::jsonb),
 'variants',coalesce((SELECT jsonb_agg(v ORDER BY p.id,v->>'id') FROM catalog_public_products p CROSS JOIN LATERAL jsonb_array_elements(p.payload->'variants') v),'[]'::jsonb),
 'offers',coalesce((SELECT jsonb_agg(o ORDER BY p.id,o->>'id') FROM catalog_public_products p CROSS JOIN LATERAL jsonb_array_elements(p.payload->'offers') o),'[]'::jsonb));
ALTER TABLE catalog_public_state ALTER COLUMN full_payload SET NOT NULL;
