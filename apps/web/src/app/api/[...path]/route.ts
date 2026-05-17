import { NextRequest, NextResponse } from 'next/server';

const API_BASE_URL = (process.env.API_BASE_URL ?? 'http://localhost:4000').replace(/\/$/, '');

type RouteContext = {
  params: Promise<{ path: string[] }> | { path: string[] };
};

export const dynamic = 'force-dynamic';

function buildTargetUrl(request: NextRequest, path: string[]) {
  const incomingUrl = new URL(request.url);
  const targetUrl = new URL(`${API_BASE_URL}/api/${path.join('/')}`);
  targetUrl.search = incomingUrl.search;
  return targetUrl;
}

function buildProxyHeaders(request: NextRequest) {
  const headers = new Headers();

  request.headers.forEach((value, key) => {
    const normalizedKey = key.toLowerCase();
    if (normalizedKey === 'host' || normalizedKey === 'connection' || normalizedKey === 'content-length') {
      return;
    }

    headers.set(key, value);
  });

  return headers;
}

async function proxyRequest(request: NextRequest, context: RouteContext) {
  const { path } = await Promise.resolve(context.params);
  const targetUrl = buildTargetUrl(request, path);
  const headers = buildProxyHeaders(request);
  const body = request.method === 'GET' || request.method === 'HEAD' ? undefined : await request.text();

  try {
    const upstreamResponse = await fetch(targetUrl, {
      method: request.method,
      headers,
      body: body === '' ? undefined : body,
      cache: 'no-store',
      redirect: 'manual',
    });

    const responseHeaders = new Headers(upstreamResponse.headers);
    responseHeaders.delete('content-length');
    responseHeaders.delete('content-encoding');

    return new Response(upstreamResponse.body, {
      status: upstreamResponse.status,
      headers: responseHeaders,
    });
  } catch (error) {
    console.error(`API proxy request failed for ${targetUrl.toString()}`, error);

    return NextResponse.json(
      {
        error: `UrbanMemory API is unavailable at ${API_BASE_URL}. Start apps/api before using this action.`,
      },
      { status: 503 }
    );
  }
}

export async function GET(request: NextRequest, context: RouteContext) {
  return proxyRequest(request, context);
}

export async function POST(request: NextRequest, context: RouteContext) {
  return proxyRequest(request, context);
}

export async function PUT(request: NextRequest, context: RouteContext) {
  return proxyRequest(request, context);
}

export async function PATCH(request: NextRequest, context: RouteContext) {
  return proxyRequest(request, context);
}

export async function DELETE(request: NextRequest, context: RouteContext) {
  return proxyRequest(request, context);
}

export async function OPTIONS(request: NextRequest, context: RouteContext) {
  return proxyRequest(request, context);
}
