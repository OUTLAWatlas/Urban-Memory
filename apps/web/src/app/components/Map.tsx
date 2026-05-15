'use client';

import { useState, useEffect, useMemo } from 'react';
import Map, { Source, Layer } from 'react-map-gl/maplibre';
import 'maplibre-gl/dist/maplibre-gl.css';
import IntegrityBadge from './IntegrityBadge';

interface HoverInfo {
  feature: GeoJSON.Feature;
  x: number;
  y: number;
}

interface LayersApiResponse extends GeoJSON.FeatureCollection {
  type?: string;
  features?: GeoJSON.Feature[];
}

interface VerificationPayload {
  is_verified?: boolean;
  on_chain_hash?: string;
  status_message?: string;
}

interface VerifiedLayerResponse {
  data?: LayersApiResponse;
  verification?: VerificationPayload;
}

interface LedgerStatusResponse {
  connected?: boolean;
  error?: string;
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

const TRUST_LAYER_OPTIONS = [
  'slum_boundary',
  'admin_ward',
  'forest_cover',
  'zone_industrial',
  'zone_residential',
  'road_network',
] as const;

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
  const [geoData, setGeoData] = useState<LayersApiResponse | null>(null);
  const [verification, setVerification] = useState<VerificationPayload | null>(null);
  const [hoverInfo, setHoverInfo] = useState<HoverInfo | null>(null);
  const [selectedFeature, setSelectedFeature] = useState<GeoJSON.Feature | null>(null);
  const [trustLayerType, setTrustLayerType] = useState<string>('slum_boundary');
  const [isLoading, setIsLoading] = useState(false);
  const [ledgerStatus, setLedgerStatus] = useState<LedgerStatusResponse | null>(null);
  const [adminKey, setAdminKey] = useState('');
  const [notarizeMessage, setNotarizeMessage] = useState('');
  const [isNotarizing, setIsNotarizing] = useState(false);
  const [refreshTick, setRefreshTick] = useState(0);

  // Layer Visibility State
  const [showWards, setShowWards] = useState(true);
  const [showSlums, setShowSlums] = useState(true);
  const [showForestCover, setShowForestCover] = useState(true);
  const [showIndustrialZones, setShowIndustrialZones] = useState(true);
  const [showResidentialZones, setShowResidentialZones] = useState(true);
  const [showRoadNetwork, setShowRoadNetwork] = useState(true);

  // Fetch data
  useEffect(() => {
    setIsLoading(true);
    fetch(`/api/v1/mumbai/layers?year=${year}&layer_type=${encodeURIComponent(trustLayerType)}`)
      .then((res) => res.json())
      .then((payload: VerifiedLayerResponse) => {
        if (!Array.isArray(payload?.data?.features)) {
          setGeoData(EMPTY_FC as LayersApiResponse);
          setVerification(payload?.verification ?? null);
          return;
        }
        setGeoData(payload.data);
        setVerification(payload.verification ?? null);
      })
      .catch((err) => {
        console.error('Failed to fetch layers:', err);
        setGeoData(EMPTY_FC as LayersApiResponse);
        setVerification({ is_verified: false, status_message: 'Verification Failed' });
      })
      .finally(() => setIsLoading(false));
  }, [year, trustLayerType, refreshTick]);

  useEffect(() => {
    fetch('/api/v1/ledger/status')
      .then((res) => res.json())
      .then((data: LedgerStatusResponse) => setLedgerStatus(data))
      .catch(() => setLedgerStatus({ connected: false, error: 'status endpoint unavailable' }));
  }, []);

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
  const onChainVerified = Boolean(verification?.is_verified);
  const canShowInspectorChainMeta = onChainVerified;
  const chainUnavailable = ledgerStatus?.connected === false;

  async function handleNotarize() {
    setNotarizeMessage('');
    setIsNotarizing(true);
    try {
      const response = await fetch('/api/admin/seal-history', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...(adminKey ? { 'X-Admin-Key': adminKey } : {}),
        },
        body: JSON.stringify({
          city: 'Mumbai',
          layer_type: trustLayerType,
          year,
        }),
      });

      const payload = await response.json().catch(() => ({}));
      if (!response.ok) {
        setNotarizeMessage(`Notarize failed: ${String(payload?.error ?? response.statusText)}`);
        return;
      }

      setNotarizeMessage('Notarized successfully. Refreshing trust check...');
      setRefreshTick((v) => v + 1);
    } catch (error) {
      setNotarizeMessage(`Notarize failed: ${String(error)}`);
    } finally {
      setIsNotarizing(false);
    }
  }

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
          {canShowInspectorChainMeta && (
            <>
              <p style={{ margin: '8px 0 6px 0', fontSize: '12px', color: '#86efac', fontWeight: 700 }}>
                Ledger Verified
              </p>
              <p style={{ margin: '0 0 6px 0', fontSize: '12px', color: '#b5b5b5' }}>
                On-Chain Hash: <span style={{ color: '#fff' }}>{verification?.on_chain_hash || 'N/A'}</span>
              </p>
            </>
          )}
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
          <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
            <label style={{ display: 'flex', alignItems: 'center', gap: '8px', color: '#ccc', fontSize: '12px' }}>
              <span>Verify Layer</span>
              <select
                value={trustLayerType}
                onChange={(e) => setTrustLayerType(e.target.value)}
                style={{
                  background: '#111',
                  color: '#fff',
                  border: '1px solid #444',
                  borderRadius: '6px',
                  padding: '4px 8px',
                  fontSize: '12px',
                }}
              >
                {TRUST_LAYER_OPTIONS.map((layer) => (
                  <option key={layer} value={layer}>
                    {formatLayerName(layer)}
                  </option>
                ))}
              </select>
            </label>
            <span title={verification?.status_message || (onChainVerified ? 'Ledger match confirmed' : 'Ledger mismatch or not sealed yet')}>
              <IntegrityBadge verification={verification ?? undefined} />
            </span>
            <span style={{ fontSize: '11px', color: '#9ca3af' }}>{verification?.status_message || 'Verification Pending'}</span>
            <span style={{ fontSize: '2.5rem', fontWeight: 'bold', color: '#fff', textShadow: '0 0 10px rgba(255,255,255,0.3)' }}>
              {year}
            </span>
          </div>
        </div>
        <input
          type="range" min="1991" max="2024" value={year}
          onChange={(e) => setYear(parseInt(e.target.value))}
          style={{ width: '100%', cursor: 'pointer', height: '6px', background: '#444', borderRadius: '3px', outline: 'none' }}
        />
        {isLoading && (
          <p style={{ margin: 0, fontSize: '12px', color: '#9ca3af' }}>Verifying current layer hash against ledger...</p>
        )}
        {!isLoading && verification?.status_message && verification.status_message !== 'Match Found' && (
          <p style={{ margin: 0, fontSize: '12px', color: '#fca5a5' }}>Trust check: {verification.status_message}</p>
        )}
        {chainUnavailable && (
          <p style={{ margin: 0, fontSize: '12px', color: '#fbbf24' }}>
            Ledger status: unavailable ({ledgerStatus?.error || 'unable to connect'})
          </p>
        )}
        {!chainUnavailable && ledgerStatus?.connected && (
          <p style={{ margin: 0, fontSize: '12px', color: '#86efac' }}>Ledger status: connected</p>
        )}
        <div style={{ display: 'flex', alignItems: 'center', gap: '10px', flexWrap: 'wrap' }}>
          <input
            type="password"
            value={adminKey}
            onChange={(e) => setAdminKey(e.target.value)}
            placeholder="Admin API key"
            style={{
              background: '#111',
              color: '#fff',
              border: '1px solid #444',
              borderRadius: '6px',
              padding: '6px 8px',
              fontSize: '12px',
              minWidth: '180px',
            }}
          />
          <button
            type="button"
            onClick={handleNotarize}
            disabled={isNotarizing || chainUnavailable}
            style={{
              border: '1px solid #065f46',
              background: isNotarizing || chainUnavailable ? '#1f2937' : '#064e3b',
              color: '#d1fae5',
              borderRadius: '6px',
              padding: '6px 10px',
              fontSize: '12px',
              fontWeight: 700,
              cursor: isNotarizing || chainUnavailable ? 'not-allowed' : 'pointer',
            }}
          >
            {isNotarizing ? 'Notarizing...' : 'Notarize Selected Layer'}
          </button>
        </div>
        {notarizeMessage && (
          <p style={{ margin: 0, fontSize: '12px', color: notarizeMessage.startsWith('Notarized') ? '#86efac' : '#fca5a5' }}>
            {notarizeMessage}
          </p>
        )}
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

