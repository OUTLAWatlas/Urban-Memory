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

export default function UrbanMap() {
  const [year, setYear] = useState(2012);
  const [geoData, setGeoData] = useState<GeoJSON.FeatureCollection | null>(null);
  const [hoverInfo, setHoverInfo] = useState<HoverInfo | null>(null);

  // Layer Visibility State
  const [showWards, setShowWards] = useState(true);
  const [showSlums, setShowSlums] = useState(true);

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
      if (f.properties?.layer_type === 'admin_ward' && !showWards) return false;
      if (f.properties?.layer_type === 'slum_boundary' && !showSlums) return false;
      return true;
    });

    return { type: 'FeatureCollection' as const, features };
  }, [geoData, showWards, showSlums]);

  // Calculate stats for the visible data
  const stats = useMemo(() => {
    let slums = 0, wards = 0;
    filteredData.features.forEach((f) => {
      if (f.properties?.layer_type === 'slum_boundary') slums++;
      if (f.properties?.layer_type === 'admin_ward') wards++;
    });
    return { slums, wards };
  }, [filteredData]);

  return (
    <div style={{ width: '100vw', height: '100vh', position: 'relative', fontFamily: 'sans-serif', background: '#000' }}>
      <Map
        initialViewState={{ longitude: 72.8777, latitude: 19.0760, zoom: 10.5 }}
        mapStyle="https://basemaps.cartocdn.com/gl/dark-matter-gl-style/style.json"
        interactiveLayerIds={['urban-layers']}
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
        </Source>
      </Map>

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
            color: hoverInfo.feature.properties?.layer_type === 'slum_boundary' ? '#ef4444' : '#3b82f6'
          }}>
            {String(hoverInfo.feature.properties?.layer_type ?? '').replace('_', ' ')}
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

