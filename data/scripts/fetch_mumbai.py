import geopandas as gpd
from sqlalchemy import create_engine
import os

# Database Connection
engine = create_engine("postgresql://pilot_user:pilot_password@localhost:5432/urban_memory_backend")

def ingest_all_geojson():
    # Mapping files to their specific metadata
    target_files = {
        'BMC_admin_wards.geojson': {'layer': 'admin_ward', 'year': '2011-01-01'},
        'mumbai_census_2011.geojson': {'layer': 'census_ward', 'year': '2011-01-01'},
        'SRAslums_2000.geojson': {'layer': 'slum_boundary', 'year': '2000-01-01'},
        'SRAslums_2012.geojson': {'layer': 'slum_boundary', 'year': '2012-01-01'},
        'SRAslums_2016.geojson': {'layer': 'slum_boundary', 'year': '2016-01-01'},
        'mumbai_district_boundaries.geojson': {'layer': 'district', 'year': '2011-01-01'}
    }

    for filename, meta in target_files.items():
        path = f"data/raw/{filename}"
        if not os.path.exists(path):
            print(f"Skipping {filename} - file not found.")
            continue
            
        print(f"Ingesting {filename}...")
        gdf = gpd.read_file(path)

        target_crs = "EPSG:4326"
        if gdf.crs is None:
            # Assume WGS84 if source lacks CRS metadata to match PostGIS table
            gdf = gdf.set_crs(target_crs)
        elif gdf.crs.to_string() != target_crs:
            gdf = gdf.to_crs(target_crs)
        
        # Standardize for UrbanMemory Schema
        gdf['city_name'] = 'Mumbai'
        gdf['layer_type'] = meta['layer']
        gdf['valid_from'] = meta['year']
        gdf['source_ref'] = 'Sanjana Krishnan / BMC Open Data'
        
        # Ensure Geometry column name is 'geom' for PostGIS
        gdf = gdf.rename_geometry('geom')
        
        # Select and upload
        final_gdf = gdf[['geom', 'city_name', 'layer_type', 'valid_from', 'source_ref']]
        final_gdf.to_postgis("urban_artifacts", engine, if_exists='append', index=False)
        print(f"✅ Successfully ingested {meta['layer']} for {meta['year']}")

ingest_all_geojson()