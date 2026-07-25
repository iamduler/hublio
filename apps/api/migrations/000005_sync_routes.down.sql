DROP TABLE IF EXISTS sync_route_watermarks;
DROP TABLE IF EXISTS sync_routes;

DROP TYPE IF EXISTS sync_route_trigger;
DROP TYPE IF EXISTS sync_route_status;

-- Note: PostgreSQL cannot easily remove an enum value from aggregate_type; leave 'sync_route' if added.
