'use client';

import { useEffect, useMemo, useRef, useState } from 'react';
import MapGL, { Layer, Source, MapRef } from 'react-map-gl/maplibre';
import 'maplibre-gl/dist/maplibre-gl.css';
import IntegrityBadge from './IntegrityBadge';
import OtpModal from './OtpModal';
import ProfileControl from './ProfileControl';
import { useSession } from './session-provider';

type HoverInfo = {
  feature: GeoJSON.Feature;
  x: number;
  y: number;
};

type LayersApiResponse = {
  type?: 'FeatureCollection';
  features?: GeoJSON.Feature[];
};

type VerificationPayload = {
  is_verified?: boolean;
  on_chain_hash?: string;
  status_message?: string;
};

type VerifiedLayerResponse = {
  data?: LayersApiResponse;
  verification?: VerificationPayload;
};

type LedgerStatusResponse = {
  connected?: boolean;
  error?: string;
};

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

const ALL_LAYER_OPTION = 'all_layers';
type TrustLayerChoice = (typeof TRUST_LAYER_OPTIONS)[number] | typeof ALL_LAYER_OPTION;

function getLayerType(feature: GeoJSON.Feature | null | undefined): string {
  return String(feature?.properties?.layer_type ?? 'unknown');
}

function formatLayerName(layerType: string): string {
  return layerType
    .replace(/_/g, ' ')
    .replace(/\b\w/g, (char) => char.toUpperCase());
}

type UrbanMapProps = {
  selectedYear?: number;
  activeLayers?: Partial<Record<typeof TRUST_LAYER_OPTIONS[number], boolean>>;
  onYearChange?: (year: number) => void;
  onLayerToggle?: (layerType: string, isActive: boolean) => void;
};

export default function UrbanMap({
  selectedYear: propYear = 2012,
  activeLayers: propActiveLayers,
  onYearChange,
  onLayerToggle,
}: UrbanMapProps) {
  const { session } = useSession();
  const mapRef = useRef<MapRef | null>(null);
  
  // Initialize layer visibility from props or defaults
  const defaultActiveLayers = {
    admin_ward: true,
    slum_boundary: true,
    forest_cover: true,
    zone_industrial: true,
    zone_residential: true,
    road_network: true,
  };
  const activeLayersMap = propActiveLayers ? { ...defaultActiveLayers, ...propActiveLayers } : defaultActiveLayers;

  const [year, setYear] = useState(propYear);
  const [geoJsonData, setGeoJsonData] = useState<LayersApiResponse | null>(null);
  const [verification, setVerification] = useState<VerificationPayload | null>(null);
  const [hoverInfo, setHoverInfo] = useState<HoverInfo | null>(null);
  const [selectedFeature, setSelectedFeature] = useState<GeoJSON.Feature | null>(null);
  const [trustLayerType, setTrustLayerType] = useState<TrustLayerChoice>('slum_boundary');
  const [is3DMode, setIs3DMode] = useState(false);
  const [isMapLoading, setIsMapLoading] = useState(false);
  const [ledgerStatus, setLedgerStatus] = useState<LedgerStatusResponse | null>(null);
  const [refreshTick, setRefreshTick] = useState(0);
  const [notarizeMessage, setNotarizeMessage] = useState('');
  const [isOtpModalOpen, setIsOtpModalOpen] = useState(false);

  // Layer visibility state (can be controlled by props or local toggle)
  const [showWards, setShowWards] = useState(activeLayersMap.admin_ward ?? true);
  const [showSlums, setShowSlums] = useState(activeLayersMap.slum_boundary ?? true);
  const [showForestCover, setShowForestCover] = useState(activeLayersMap.forest_cover ?? true);
  const [showIndustrialZones, setShowIndustrialZones] = useState(activeLayersMap.zone_industrial ?? true);
  const [showResidentialZones, setShowResidentialZones] = useState(activeLayersMap.zone_residential ?? true);
  const [showRoadNetwork, setShowRoadNetwork] = useState(activeLayersMap.road_network ?? true);

  // Sync year prop changes to local state
  useEffect(() => {
    setYear(propYear);
  }, [propYear]);

  // Sync active layers prop changes to local visibility state
  useEffect(() => {
    if (propActiveLayers) {
      if (propActiveLayers.admin_ward !== undefined) setShowWards(propActiveLayers.admin_ward);
      if (propActiveLayers.slum_boundary !== undefined) setShowSlums(propActiveLayers.slum_boundary);
      if (propActiveLayers.forest_cover !== undefined) setShowForestCover(propActiveLayers.forest_cover);
      if (propActiveLayers.zone_industrial !== undefined) setShowIndustrialZones(propActiveLayers.zone_industrial);
      if (propActiveLayers.zone_residential !== undefined) setShowResidentialZones(propActiveLayers.zone_residential);
      if (propActiveLayers.road_network !== undefined) setShowRoadNetwork(propActiveLayers.road_network);
    }
  }, [propActiveLayers]);

  // Dynamic data fetching effect: responds to year and active layers changes
  useEffect(() => {
    let isActive = true;
    setIsMapLoading(true);

    const fetchLayerData = async (layerType: string) => {
      const response = await fetch(`/api/v1/mumbai/layers?year=${year}&layer_type=${encodeURIComponent(layerType)}`);

      if (!response.ok) {
        return null;
      }

      return (await response.json()) as VerifiedLayerResponse;
    };

    const loadSingleLayer = async () => {
      try {
        const payload = await fetchLayerData(trustLayerType);

        if (!isActive) {
          return;
        }

        if (!payload) {
          setGeoJsonData(EMPTY_FC as LayersApiResponse);
          setVerification({ is_verified: false, status_message: 'Fetch Failed' });
          return;
        }

        if (payload.data && Array.isArray(payload.data.features)) {
          setGeoJsonData(payload.data);
        } else {
          setGeoJsonData(EMPTY_FC as LayersApiResponse);
        }

        setVerification(
          payload.verification ?? {
            is_verified: false,
            status_message: 'Unknown Status',
          }
        );
      } catch {
        if (isActive) {
          setGeoJsonData(EMPTY_FC as LayersApiResponse);
          setVerification({ is_verified: false, status_message: 'Network Error' });
        }
      } finally {
        if (isActive) {
          setIsMapLoading(false);
        }
      }
    };

    const loadAllLayers = async () => {
      try {
        const payloads = await Promise.all(TRUST_LAYER_OPTIONS.map((layerType) => fetchLayerData(layerType)));

        if (!isActive) {
          return;
        }

        const combinedFeatures = payloads.flatMap((payload) => payload?.data?.features ?? []);
        setGeoJsonData({
          type: 'FeatureCollection',
          features: combinedFeatures,
        });
        setVerification({
          is_verified: false,
          status_message: combinedFeatures.length > 0 ? 'All Layers Loaded' : 'Record Not Found',
        });
      } catch {
        if (isActive) {
          setGeoJsonData(EMPTY_FC as LayersApiResponse);
          setVerification({ is_verified: false, status_message: 'Network Error' });
        }
      } finally {
        if (isActive) {
          setIsMapLoading(false);
        }
      }
    };

    if (trustLayerType === ALL_LAYER_OPTION) {
      void loadAllLayers();
    } else {
      void loadSingleLayer();
    }

    return () => {
      isActive = false;
    };
  }, [year, trustLayerType, refreshTick]);

  useEffect(() => {
    fetch('/api/v1/ledger/status')
      .then((response) => response.json())
      .then((data: LedgerStatusResponse) => setLedgerStatus(data))
      .catch(() => setLedgerStatus({ connected: false, error: 'status endpoint unavailable' }));
  }, []);

  const filteredData = useMemo(() => {
    if (!geoJsonData || !geoJsonData.features) {
      return EMPTY_FC;
    }

    const features = geoJsonData.features.filter((feature) => {
      const layerType = getLayerType(feature);
      if (layerType === 'admin_ward' && !showWards) return false;
      if (layerType === 'slum_boundary' && !showSlums) return false;
      if (layerType === 'forest_cover' && !showForestCover) return false;
      if (layerType === 'zone_industrial' && !showIndustrialZones) return false;
      if (layerType === 'zone_residential' && !showResidentialZones) return false;
      if (layerType === 'road_network' && !showRoadNetwork) return false;
      return true;
    });

    return { type: 'FeatureCollection' as const, features };
  }, [geoJsonData, showForestCover, showIndustrialZones, showResidentialZones, showRoadNetwork, showSlums, showWards]);

  const stats = useMemo(() => {
    const totals = {
      slums: 0,
      wards: 0,
      forests: 0,
      industrial: 0,
      residential: 0,
      roads: 0,
    };

    filteredData.features.forEach((feature) => {
      const layerType = getLayerType(feature);
      if (layerType === 'slum_boundary') totals.slums += 1;
      if (layerType === 'admin_ward') totals.wards += 1;
      if (layerType === 'forest_cover') totals.forests += 1;
      if (layerType === 'zone_industrial') totals.industrial += 1;
      if (layerType === 'zone_residential') totals.residential += 1;
      if (layerType === 'road_network') totals.roads += 1;
    });

    return totals;
  }, [filteredData]);

  const selectedLayerType = getLayerType(selectedFeature);
  const selectedLayerColor = LAYER_COLORS[selectedLayerType] ?? '#888888';
  const canShowInspectorChainMeta = Boolean(verification?.is_verified);
  const chainUnavailable = ledgerStatus?.connected === false;
  const isAdmin = session.mode === 'admin' && Boolean(session.admin?.token);
  const canSeal = Boolean(isAdmin && session.admin?.token && session.admin?.role !== 'pending');
  const mapboxToken = (process.env.NEXT_PUBLIC_MAPBOX_ACCESS_TOKEN || '').trim();
  const hasMapboxToken = Boolean(mapboxToken) &&
                         !mapboxToken.startsWith('your_') &&
                         !mapboxToken.includes('example_token');
  const useMapboxStyle = is3DMode && hasMapboxToken;
  const mapStyle = useMapboxStyle
    ? `https://api.mapbox.com/styles/v1/mapbox/dark-v11?access_token=${mapboxToken}`
    : 'https://basemaps.cartocdn.com/gl/dark-matter-gl-style/style.json';

  useEffect(() => {
    const map = mapRef.current?.getMap?.();
    if (!map) {
      return;
    }

    map.flyTo({
      pitch: is3DMode ? 60 : 0,
      bearing: is3DMode ? -20 : 0,
      duration: 1800,
      essential: true,
    });
  }, [is3DMode]);

  const handleOpenOtpModal = () => {
    if (!isAdmin) {
      setNotarizeMessage('Administrative access is required to notarize layer history.');
      return;
    }

    setNotarizeMessage('');
    setIsOtpModalOpen(true);
  };

  const handleSealSuccess = (payload: { message?: string; tx_hash?: string; sha256_hash?: string }) => {
    setVerification({
      is_verified: true,
      on_chain_hash: payload.tx_hash ?? payload.sha256_hash,
      status_message: 'Verified',
    });
    setNotarizeMessage(String(payload.message ?? 'Layer history sealed successfully.'));
  };

  return (
    <div className="relative h-screen w-screen overflow-hidden bg-[#040816] text-white">
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_top_left,rgba(34,211,238,0.08),transparent_35%),radial-gradient(circle_at_bottom_right,rgba(16,185,129,0.12),transparent_28%),linear-gradient(180deg,#040816_0%,#02040a_100%)]" />

      <div className="absolute inset-0">
        <MapGL
          ref={mapRef}
          initialViewState={{ longitude: 72.8777, latitude: 19.076, zoom: 10.5 }}
          mapStyle={mapStyle}
          interactiveLayerIds={['urban-layers', 'road-network-lines']}
          onClick={(event) => {
            if (event.features && event.features.length > 0) {
              setSelectedFeature(event.features[0] as GeoJSON.Feature);
            }
          }}
          onMouseMove={(event) => {
            if (event.features && event.features.length > 0) {
              setHoverInfo({ feature: event.features[0] as GeoJSON.Feature, x: event.point.x, y: event.point.y });
            } else {
              setHoverInfo(null);
            }
          }}
          onMouseLeave={() => setHoverInfo(null)}
        >
          {useMapboxStyle && (
            <Layer
              id="3d-buildings"
              source="composite"
              source-layer="building"
              type="fill-extrusion"
              minzoom={14}
              beforeId="road-network-lines"
              paint={{
                'fill-extrusion-height': ['interpolate', ['linear'], ['zoom'], 15, 0, 15.05, ['get', 'height']],
                'fill-extrusion-base': ['coalesce', ['get', 'min_height'], 0],
                'fill-extrusion-opacity': is3DMode ? 0.82 : 0,
                'fill-extrusion-color': [
                  'case',
                  ['==', ['get', 'class'], 'verified'], '#0f766e',
                  ['==', ['get', 'class'], 'commercial'], '#27272a',
                  '#1f2937',
                ],
              }}
            />
          )}

          <Source id="urban-data" type="geojson" data={filteredData}>
            {is3DMode && (
              <Layer
                id="urban-topography"
                type="fill-extrusion"
                filter={['!=', ['get', 'layer_type'], 'road_network']}
                paint={{
                  'fill-extrusion-height': [
                    'match',
                    ['get', 'layer_type'],
                    'forest_cover', 70,
                    'admin_ward', 120,
                    'slum_boundary', 180,
                    'zone_residential', 220,
                    'zone_industrial', 260,
                    40,
                  ],
                  'fill-extrusion-base': 0,
                  'fill-extrusion-opacity': 0.52,
                  'fill-extrusion-color': [
                    'match',
                    ['get', 'layer_type'],
                    'forest_cover', '#14532d',
                    'admin_ward', '#1d4ed8',
                    'slum_boundary', '#7f1d1d',
                    'zone_residential', '#5b21b6',
                    'zone_industrial', '#854d0e',
                    '#334155',
                  ],
                }}
              />
            )}

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
                  '#888888',
                ],
                'fill-opacity': ['case', ['boolean', ['feature-state', 'hover'], false], 0.8, 0.3],
                'fill-outline-color': '#ffffff',
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
        </MapGL>

        {isMapLoading && (
          <div className="absolute inset-0 flex items-center justify-center bg-black/20 backdrop-blur-sm z-20">
            <div className="flex flex-col items-center gap-4">
              <div className="h-12 w-12 animate-spin rounded-full border-4 border-slate-600 border-t-cyan-400" />
              <p className="text-sm font-semibold text-cyan-300">Loading map data...</p>
            </div>
          </div>
        )}
      </div>

      {selectedFeature && (
        <div
          className="absolute left-4 top-4 z-10 w-[min(92vw,22rem)] rounded-[1.5rem] border border-white/10 bg-slate-950/90 p-5 text-white shadow-[0_20px_90px_rgba(0,0,0,0.45)] backdrop-blur-xl"
          style={{ borderLeftColor: selectedLayerColor, borderLeftWidth: 6 }}
        >
          <h3 className="text-sm font-semibold uppercase tracking-[0.3em] text-slate-400">Deep Dive Inspector</h3>
          <p className="mt-3 text-xl font-semibold" style={{ color: selectedLayerColor }}>
            {formatLayerName(selectedLayerType)}
          </p>
          <div className="mt-4 space-y-2 text-sm text-slate-400">
            <p>
              Source: <span className="text-white">{String(selectedFeature.properties?.source_ref ?? 'Unknown')}</span>
            </p>
            <p>
              Valid From: <span className="text-white">{String(selectedFeature.properties?.valid_from ?? 'N/A')}</span>
            </p>
            <p>
              Valid To: <span className="text-white">{String(selectedFeature.properties?.valid_to ?? 'Present')}</span>
            </p>
            {canShowInspectorChainMeta && (
              <>
                <p className="pt-1 font-semibold text-emerald-300">Ledger Verified</p>
                <p>
                  On-Chain Hash: <span className="text-white">{verification?.on_chain_hash || 'N/A'}</span>
                </p>
              </>
            )}
          </div>
        </div>
      )}

      {hoverInfo && !selectedFeature && (
        <div
          className="pointer-events-none absolute z-10 rounded-full border border-white/10 bg-slate-950/90 px-3 py-2 text-xs text-slate-300 shadow-lg backdrop-blur-md"
          style={{ left: hoverInfo.x + 12, top: hoverInfo.y + 12 }}
        >
          {formatLayerName(getLayerType(hoverInfo.feature))}
        </div>
      )}

      <aside className="absolute top-24 right-4 z-10 w-80 flex flex-col max-h-[calc(100vh-120px)] rounded-[1.75rem] border border-white/10 bg-slate-950/88 text-white shadow-[0_24px_100px_rgba(0,0,0,0.5)] backdrop-blur-2xl">
        <div className="shrink-0 p-5">
          <div className="flex items-start justify-between gap-4 border-b border-white/10 pb-4">
            <div>
              <p className="text-xs uppercase tracking-[0.35em] text-slate-500">Layer Controls</p>
              <h2 className="mt-2 text-2xl font-semibold">Temporal Delta Engine</h2>
            </div>
            <div className="text-right text-xs text-slate-500">
              <p>{session.mode === 'admin' ? 'Admin session' : 'Public session'}</p>
              <p>{session.admin?.role ? `Role: ${session.admin.role}` : 'Read only'}</p>
            </div>
          </div>
        </div>

        <div className="overflow-y-auto pr-1 normal-scrollbar flex-1 p-5">
          <div className="space-y-4">
          <button
            type="button"
            onClick={() => setIs3DMode((current) => !current)}
            className={`flex w-full items-center justify-between gap-4 rounded-2xl border px-4 py-3 text-left transition ${
              is3DMode
                ? 'border-cyan-400/40 bg-cyan-400/10 text-cyan-100 shadow-[0_0_0_1px_rgba(34,211,238,0.12)]'
                : 'border-white/10 bg-white/5 text-slate-200 hover:border-white/20 hover:bg-white/10'
            }`}
          >
            <span className="flex flex-col">
              <span className="text-xs uppercase tracking-[0.25em] text-slate-500">3D Cityscape</span>
              <span className="mt-1 text-sm font-semibold">Enable 3D Topography</span>
            </span>
            <span className={`relative h-6 w-11 rounded-full border transition ${is3DMode ? 'border-cyan-400/60 bg-cyan-400/25' : 'border-white/15 bg-slate-800'}`}>
              <span
                className={`absolute top-0.5 h-5 w-5 rounded-full bg-white shadow-[0_8px_20px_rgba(0,0,0,0.35)] transition-transform ${
                  is3DMode ? 'translate-x-5 bg-cyan-100' : 'translate-x-0.5'
                }`}
              />
            </span>
          </button>

          {!hasMapboxToken && is3DMode && (
            <p className="rounded-2xl border border-amber-400/20 bg-amber-400/10 px-4 py-3 text-xs text-amber-100">
              Add a valid <span className="font-semibold">NEXT_PUBLIC_MAPBOX_ACCESS_TOKEN</span> to load Mapbox building extrusions. The civic layers still lift into a stylized 3D topography on the current basemap.
            </p>
          )}

          <label className="flex items-center justify-between gap-4 text-sm text-slate-300">
            <span className="min-w-0">Verify Layer</span>
            <select
              value={trustLayerType}
              onChange={(event) => setTrustLayerType(event.target.value as TrustLayerChoice)}
              className="rounded-xl border border-white/10 bg-slate-900 px-3 py-2 text-sm text-white outline-none transition focus:border-cyan-400/50 focus:ring-2 focus:ring-cyan-400/20"
            >
              <option value={ALL_LAYER_OPTION}>All Layers</option>
              {TRUST_LAYER_OPTIONS.map((layer) => (
                <option key={layer} value={layer}>
                  {formatLayerName(layer)}
                </option>
              ))}
            </select>
          </label>

          <div className="rounded-2xl border border-white/10 bg-white/5 p-3">
            <div className="flex items-center justify-between gap-3">
              <IntegrityBadge verification={verification ?? undefined} />
              <span className="text-xs text-slate-400">{verification?.status_message || (isLoading ? 'Verifying...' : 'Status unknown')}</span>
            </div>
            {verification?.on_chain_hash && (
              <p className="mt-3 break-all text-xs text-slate-400">Hash: {verification.on_chain_hash}</p>
            )}
          </div>

          <div className="grid grid-cols-2 gap-3 text-sm text-slate-300">
            <label className="rounded-2xl border border-white/10 bg-white/5 p-3">
              <span className="mb-3 block text-xs uppercase tracking-[0.25em] text-slate-500">Administrative Wards</span>
              <input
                type="checkbox"
                checked={showWards}
                onChange={(event) => {
                  setShowWards(event.target.checked);
                  onLayerToggle?.('admin_ward', event.target.checked);
                }}
                className="accent-cyan-400"
              />
              <span className="ml-2">{stats.wards}</span>
            </label>
            <label className="rounded-2xl border border-white/10 bg-white/5 p-3">
              <span className="mb-3 block text-xs uppercase tracking-[0.25em] text-slate-500">SRA Slum Clusters</span>
              <input
                type="checkbox"
                checked={showSlums}
                onChange={(event) => {
                  setShowSlums(event.target.checked);
                  onLayerToggle?.('slum_boundary', event.target.checked);
                }}
                className="accent-rose-400"
              />
              <span className="ml-2">{stats.slums}</span>
            </label>
            <label className="rounded-2xl border border-white/10 bg-white/5 p-3">
              <span className="mb-3 block text-xs uppercase tracking-[0.25em] text-slate-500">Forest Cover</span>
              <input
                type="checkbox"
                checked={showForestCover}
                onChange={(event) => {
                  setShowForestCover(event.target.checked);
                  onLayerToggle?.('forest_cover', event.target.checked);
                }}
                className="accent-emerald-400"
              />
              <span className="ml-2">{stats.forests}</span>
            </label>
            <label className="rounded-2xl border border-white/10 bg-white/5 p-3">
              <span className="mb-3 block text-xs uppercase tracking-[0.25em] text-slate-500">Industrial Zones</span>
              <input
                type="checkbox"
                checked={showIndustrialZones}
                onChange={(event) => {
                  setShowIndustrialZones(event.target.checked);
                  onLayerToggle?.('zone_industrial', event.target.checked);
                }}
                className="accent-yellow-400"
              />
              <span className="ml-2">{stats.industrial}</span>
            </label>
            <label className="rounded-2xl border border-white/10 bg-white/5 p-3">
              <span className="mb-3 block text-xs uppercase tracking-[0.25em] text-slate-500">Residential Zones</span>
              <input
                type="checkbox"
                checked={showResidentialZones}
                onChange={(event) => {
                  setShowResidentialZones(event.target.checked);
                  onLayerToggle?.('zone_residential', event.target.checked);
                }}
                className="accent-violet-400"
              />
              <span className="ml-2">{stats.residential}</span>
            </label>
            <label className="rounded-2xl border border-white/10 bg-white/5 p-3">
              <span className="mb-3 block text-xs uppercase tracking-[0.25em] text-slate-500">Arterial Roads</span>
              <input
                type="checkbox"
                checked={showRoadNetwork}
                onChange={(event) => {
                  setShowRoadNetwork(event.target.checked);
                  onLayerToggle?.('road_network', event.target.checked);
                }}
                className="accent-white"
              />
              <span className="ml-2">{stats.roads}</span>
            </label>
          </div>

          <div className="rounded-2xl border border-white/10 bg-white/5 p-4">
            <div className="mb-3 flex items-center justify-between gap-3">
              <span className="text-xs uppercase tracking-[0.3em] text-slate-500">Timeline</span>
              <span className="text-4xl font-semibold tracking-tight text-white">{year}</span>
            </div>
            <input
              type="range"
              min="1991"
              max="2024"
              value={year}
              onChange={(event) => {
                const newYear = Number(event.target.value);
                setYear(newYear);
                onYearChange?.(newYear);
              }}
              className="h-2 w-full cursor-pointer appearance-none rounded-full bg-slate-700 accent-cyan-400"
            />
            <div className="mt-2 flex justify-between text-[10px] uppercase tracking-[0.25em] text-slate-500">
              <span>1991</span>
              <span>2000</span>
              <span>2012</span>
              <span>2024</span>
            </div>
          </div>

          {chainUnavailable && (
            <p className="rounded-2xl border border-amber-400/20 bg-amber-400/10 px-4 py-3 text-sm text-amber-100">
              Ledger status: unavailable ({ledgerStatus?.error || 'unable to connect'})
            </p>
          )}

          {!chainUnavailable && ledgerStatus?.connected && (
            <p className="rounded-2xl border border-emerald-400/20 bg-emerald-400/10 px-4 py-3 text-sm text-emerald-100">Ledger status: connected</p>
          )}

          {isAdmin ? (
            <button
              type="button"
              onClick={handleOpenOtpModal}
              disabled={!canSeal}
              className="w-full rounded-full bg-gradient-to-r from-emerald-400 to-cyan-400 px-5 py-4 text-sm font-semibold text-slate-950 transition hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-50"
            >
              Notarize Layer History
            </button>
          ) : (
            <p className="rounded-2xl border border-white/10 bg-white/5 px-4 py-3 text-sm text-slate-400">
              Public sessions remain read-only. The seal action appears after administrator sign-in.
            </p>
          )}

          {session.admin?.role === 'pending' && isAdmin && (
            <p className="rounded-2xl border border-white/10 bg-white/5 px-4 py-3 text-sm text-slate-400">
              This account is pending approval and cannot confirm a seal yet.
            </p>
          )}

          {notarizeMessage && (
            <p className="rounded-2xl border border-white/10 bg-white/5 px-4 py-3 text-sm text-slate-300">{notarizeMessage}</p>
          )}

          {hoverInfo && (
            <div className="rounded-2xl border border-white/10 bg-white/5 px-4 py-3 text-xs text-slate-400">
              Hovering {formatLayerName(getLayerType(hoverInfo.feature))} at {Math.round(hoverInfo.x)}, {Math.round(hoverInfo.y)}
            </div>
          )}
            </div>
          </div>
        </aside>

      <div className="absolute bottom-4 left-4 z-10 rounded-[1.5rem] border border-white/10 bg-slate-950/90 px-4 py-3 text-sm text-slate-300 shadow-[0_20px_90px_rgba(0,0,0,0.45)] backdrop-blur-xl">
        <div className="flex items-center gap-3">
          <span className="text-xs uppercase tracking-[0.3em] text-slate-500">Session</span>
          <span>{session.mode === 'admin' ? 'Administrative governance' : 'Public explorer'}</span>
        </div>
      </div>

      <OtpModal
        open={isOtpModalOpen}
        adminSession={session.admin}
        layerType={trustLayerType}
        year={year}
        onClose={() => setIsOtpModalOpen(false)}
        onSuccess={handleSealSuccess}
      />
    </div>
  );
}


