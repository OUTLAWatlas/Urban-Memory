CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE IF NOT EXISTS urban_artifacts (
    id SERIAL PRIMARY KEY,
    city_name VARCHAR(50) NOT NULL,
    layer_type VARCHAR(50) NOT NULL,
    geom GEOMETRY(GEOMETRY, 4326) NOT NULL,
    valid_from DATE NOT NULL,
    valid_to DATE,
    source_ref TEXT,
    data_hash TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_spatial_artifacts ON urban_artifacts USING GIST(geom);