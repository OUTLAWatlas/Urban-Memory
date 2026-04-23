'use client';

import { useState, useEffect, useMemo } from 'react';
import Map, { Source, Layer } from 'react-map-gl/maplibre';
import 'maplibre-gl/dist/maplibre-gl.css';

interface HoverInfo {
  feature: GeoJSON.Feature;
  x: number;
  y: number;
}

const EMPTY_FC: GeoJSON.FeatureCollection = { type: 'FeatureCollection', features: [] };

const LAYER_COLORS: Record<string, string> = {
  slum_boundary: '#ef4444',
  admin_ward: '#3b82f6',
  forest_cover: '#22c55e',
  zone_industrial: '#facc15',
  zone_residential: '#a855f7',
  road_network: '#ffffff',
  census_ward: '#06b6d4',
  electoral_2017: '#f97316',
  electoral_2022: '#ec4899',
};

function getLayerType(feature: GeoJSON.Feature | null | undefined): string {
  return String(feature?.properties?.layer_type ?? 'unknown');
}

function formatLayerName(layerType: string): string {
  return layerType
    .replace(/_/g, ' ')
    .replace(/\b\w/g, (char) => char.toUpperCase());
}

export default function UrbanMap() {
  const [year, setYear] = useState(2012);
  const [geoData, setGeoData] = useState<GeoJSON.FeatureCollection | null>(null);
  const [hoverInfo, setHoverInfo] = useState<HoverInfo | null>(null);
  const [selectedFeature, setSelectedFeature] = useState<GeoJSON.Feature | null>(null);

  // Layer Visibility State
  const [showWards, setShowWards] = useState(true);
  const [showSlums, setShowSlums] = useState(true);
  const [showForestCover, setShowForestCover] = useState(true);
  const [showIndustrialZones, setShowIndustrialZones] = useState(true);
  const [showResidentialZones, setShowResidentialZones] = useState(true);
  const [showRoadNetwork, setShowRoadNetwork] = useState(true);

  // Fetch data
  useEffect(() => {
    fetch(`/api/v1/mumbai/layers?year=${year}`)
      .then(res => res.json())
      .then(data => setGeoData(data))
      .catch(err => console.error("Failed to fetch layers:", err));
  }, [year]);

  // Filter the geoData based on what the user wants to see
  const filteredData = useMemo(() => {
    if (!geoData || !geoData.features) return EMPTY_FC;

    const features = geoData.features.filter((f) => {
      const layerType = getLayerType(f);
      if (layerType === 'admin_ward' && !showWards) return false;
      if (layerType === 'slum_boundary' && !showSlums) return false;
      if (layerType === 'forest_cover' && !showForestCover) return false;
      if (layerType === 'zone_industrial' && !showIndustrialZones) return false;
      if (layerType === 'zone_residential' && !showResidentialZones) return false;
      if (layerType === 'road_network' && !showRoadNetwork) return false;
      return true;
    });

    return { type: 'FeatureCollection' as const, features };
  }, [
    geoData,
    showWards,
    showSlums,
    showForestCover,
    showIndustrialZones,
    showResidentialZones,
    showRoadNetwork,
  ]);

  // Calculate stats for the visible data
  const stats = useMemo(() => {
    let slums = 0;
    let wards = 0;
    let forests = 0;
    let industrial = 0;
    let residential = 0;
    let roads = 0;

    filteredData.features.forEach((f) => {
      const layerType = getLayerType(f);
      if (layerType === 'slum_boundary') slums++;
      if (layerType === 'admin_ward') wards++;
      if (layerType === 'forest_cover') forests++;
      if (layerType === 'zone_industrial') industrial++;
      if (layerType === 'zone_residential') residential++;
      if (layerType === 'road_network') roads++;
    });

    return { slums, wards, forests, industrial, residential, roads };
  }, [filteredData]);

  const selectedLayerType = getLayerType(selectedFeature);
  const selectedLayerColor = LAYER_COLORS[selectedLayerType] ?? '#888888';

  return (
    <div style={{ width: '100vw', height: '100vh', position: 'relative', fontFamily: 'sans-serif', background: '#000' }}>
      <Map
        initialViewState={{ longitude: 72.8777, latitude: 19.0760, zoom: 10.5 }}
        mapStyle="https://basemaps.cartocdn.com/gl/dark-matter-gl-style/style.json"
        interactiveLayerIds={['urban-layers', 'road-network-lines']}
        onClick={(e) => {
          if (e.features && e.features.length > 0) {
            setSelectedFeature(e.features[0] as GeoJSON.Feature);
          }
        }}
        onMouseMove={(e) => {
          if (e.features && e.features.length > 0) {
            setHoverInfo({ feature: e.features[0] as GeoJSON.Feature, x: e.point.x, y: e.point.y });
          } else {
            setHoverInfo(null);
          }
        }}
        onMouseLeave={() => setHoverInfo(null)}
      >
        <Source type="geojson" data={filteredData}>
          <Layer
            id="urban-layers"
            type="fill"
            paint={{
              'fill-color': [
                'match',
                ['get', 'layer_type'],
                'slum_boundary', '#ef4444',
                'admin_ward', '#3b82f6',
                'forest_cover', '#22c55e',
                'zone_industrial', '#facc15',
                'zone_residential', '#a855f7',
                'road_network', '#ffffff',
                '#888888'
              ],
              'fill-opacity': [
                'case',
                ['boolean', ['feature-state', 'hover'], false],
                0.8,
                0.3
              ],
              'fill-outline-color': '#ffffff'
            }}
          />
          <Layer
            id="road-network-lines"
            type="line"
            filter={['==', ['get', 'layer_type'], 'road_network']}
            paint={{
              'line-color': '#ffffff',
              'line-width': 2,
              'line-opacity': 0.95,
            }}
          />
        </Source>
      </Map>

      {/* Data Inspector (Left Panel) */}
      {selectedFeature && (
        <div style={{
          position: 'absolute', top: 20, left: 20,
          backgroundColor: 'rgba(20, 20, 20, 0.95)', padding: '16px', borderRadius: '8px',
          color: 'white', border: `1px solid ${selectedLayerColor}`, minWidth: '280px',
          borderLeft: `6px solid ${selectedLayerColor}`,
          boxShadow: '0 4px 6px rgba(0,0,0,0.5)'
        }}>
          <h3 style={{ margin: '0 0 10px 0', fontSize: '1rem', letterSpacing: '0.4px' }}>
            Deep Dive Inspector
          </h3>
          <p style={{ margin: '0 0 8px 0', fontWeight: 'bold', color: selectedLayerColor }}>
            {formatLayerName(selectedLayerType)}
          </p>
          <p style={{ margin: '0 0 6px 0', fontSize: '12px', color: '#b5b5b5' }}>
            Source: <span style={{ color: '#fff' }}>{String(selectedFeature.properties?.source_ref ?? 'Unknown')}</span>
          </p>
          <p style={{ margin: '0 0 6px 0', fontSize: '12px', color: '#b5b5b5' }}>
            Valid From: <span style={{ color: '#fff' }}>{String(selectedFeature.properties?.valid_from ?? 'N/A')}</span>
          </p>
          <p style={{ margin: 0, fontSize: '12px', color: '#b5b5b5' }}>
            Valid To: <span style={{ color: '#fff' }}>{String(selectedFeature.properties?.valid_to ?? 'Present')}</span>
          </p>
        </div>
      )}

      {/* Control Panel (Top Right) */}
      <div style={{
        position: 'absolute', top: 20, right: 20,
        backgroundColor: 'rgba(20, 20, 20, 0.95)', padding: '20px', borderRadius: '8px',
        color: 'white', border: '1px solid #333', minWidth: '250px',
        boxShadow: '0 4px 6px rgba(0,0,0,0.5)'
      }}>
        <h2 style={{ margin: '0 0 15px 0', fontSize: '1.2rem', borderBottom: '1px solid #444', paddingBottom: '10px' }}>
          Layer Controls
        </h2>

        <div style={{ display: 'flex', flexDirection: 'column', gap: '15px' }}>
          <label style={{ display: 'flex', alignItems: 'center', gap: '10px', cursor: 'pointer' }}>
            <input
              type="checkbox"
              checked={showWards}
              onChange={(e) => setShowWards(e.target.checked)}
              style={{ width: '18px', height: '18px', accentColor: '#3b82f6' }}
            />
            <span style={{ fontSize: '1rem' }}>🔵 Administrative Wards ({stats.wards})</span>
          </label>

          <label style={{ display: 'flex', alignItems: 'center', gap: '10px', cursor: 'pointer' }}>
            <input
              type="checkbox"
              checked={showSlums}
              onChange={(e) => setShowSlums(e.target.checked)}
              style={{ width: '18px', height: '18px', accentColor: '#ef4444' }}
            />
            <span style={{ fontSize: '1rem' }}>🔴 SRA Slum Clusters ({stats.slums})</span>
          </label>

          <label style={{ display: 'flex', alignItems: 'center', gap: '10px', cursor: 'pointer' }}>
            <input
              type="checkbox"
              checked={showForestCover}
              onChange={(e) => setShowForestCover(e.target.checked)}
              style={{ width: '18px', height: '18px', accentColor: '#22c55e' }}
            />
            <span style={{ fontSize: '1rem' }}>🟢 Forest Cover ({stats.forests})</span>
          </label>

          <label style={{ display: 'flex', alignItems: 'center', gap: '10px', cursor: 'pointer' }}>
            <input
              type="checkbox"
              checked={showIndustrialZones}
              onChange={(e) => setShowIndustrialZones(e.target.checked)}
              style={{ width: '18px', height: '18px', accentColor: '#facc15' }}
            />
            <span style={{ fontSize: '1rem' }}>🟡 Industrial Zones ({stats.industrial})</span>
          </label>

          <label style={{ display: 'flex', alignItems: 'center', gap: '10px', cursor: 'pointer' }}>
            <input
              type="checkbox"
              checked={showResidentialZones}
              onChange={(e) => setShowResidentialZones(e.target.checked)}
              style={{ width: '18px', height: '18px', accentColor: '#a855f7' }}
            />
            <span style={{ fontSize: '1rem' }}>🟣 Residential Zones ({stats.residential})</span>
          </label>

          <label style={{ display: 'flex', alignItems: 'center', gap: '10px', cursor: 'pointer' }}>
            <input
              type="checkbox"
              checked={showRoadNetwork}
              onChange={(e) => setShowRoadNetwork(e.target.checked)}
              style={{ width: '18px', height: '18px', accentColor: '#ffffff' }}
            />
            <span style={{ fontSize: '1rem' }}>⚪ Arterial Roads ({stats.roads})</span>
          </label>
        </div>
      </div>

      {/* The Time Machine Slider (Bottom) */}
      <div style={{
        position: 'absolute', bottom: 40, left: '10%', right: '10%',
        backgroundColor: 'rgba(20, 20, 20, 0.95)', padding: '20px', borderRadius: '8px',
        color: 'white', display: 'flex', flexDirection: 'column', gap: '10px',
        border: '1px solid #333', boxShadow: '0 -4px 10px rgba(0,0,0,0.5)'
      }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <span style={{ color: '#aaa', fontWeight: '500', textTransform: 'uppercase', letterSpacing: '1px' }}>
            Temporal Delta Engine
          </span>
          <span style={{ fontSize: '2.5rem', fontWeight: 'bold', color: '#fff', textShadow: '0 0 10px rgba(255,255,255,0.3)' }}>
            {year}
          </span>
        </div>
        <input
          type="range" min="1991" max="2024" value={year}
          onChange={(e) => setYear(parseInt(e.target.value))}
          style={{ width: '100%', cursor: 'pointer', height: '6px', background: '#444', borderRadius: '3px', outline: 'none' }}
        />
        <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '12px', color: '#666', fontWeight: 'bold' }}>
          <span>1991 (Baseline)</span>
          <span>2000 (Slum Data Start)</span>
          <span>2012 (Growth)</span>
          <span>2024 (Present)</span>
        </div>
      </div>

      {/* Tooltip */}
      {hoverInfo && (
        <div style={{
          position: 'absolute', left: hoverInfo.x + 15, top: hoverInfo.y + 15,
          background: 'rgba(0,0,0,0.9)', color: 'white', padding: '12px',
          borderRadius: '6px', pointerEvents: 'none', zIndex: 10, border: '1px solid #444'
        }}>
          <p style={{
            margin: 0, fontWeight: 'bold', textTransform: 'uppercase', fontSize: '0.9rem',
            color: LAYER_COLORS[getLayerType(hoverInfo.feature)] ?? '#888888'
          }}>
            {formatLayerName(getLayerType(hoverInfo.feature))}
          </p>
          <hr style={{ borderColor: '#333', margin: '5px 0' }} />
          <p style={{ margin: 0, fontSize: '12px', color: '#aaa' }}>
            Source: <span style={{ color: '#fff' }}>{String(hoverInfo.feature.properties?.source_ref ?? 'Unknown')}</span>
          </p>
          <p style={{ margin: 0, fontSize: '12px', color: '#aaa' }}>
            Valid From: <span style={{ color: '#fff' }}>{String(hoverInfo.feature.properties?.valid_from ?? '')}</span>
          </p>
        </div>
      )}
    </div>
  );
}

