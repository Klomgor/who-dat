/**
 * Cloudflare Worker for Who-Dat WHOIS Lookup
 * This worker loads the Go WASM module and handles HTTP requests
 */

let wasmInstance = null;
let wasmInitialized = false;

// Initialize WASM module
async function initWasm() {
  if (wasmInitialized) return true;

  try {
    // Load WASM binary
    const wasmModule = await WebAssembly.compileStreaming(
      fetch('/who-dat.wasm')
    );

    // Create Go runtime environment
    const go = new Go();
    wasmInstance = await WebAssembly.instantiate(wasmModule, go.importObject);

    // Run Go program
    go.run(wasmInstance);

    wasmInitialized = true;
    return true;
  } catch (err) {
    console.error('Failed to initialize WASM:', err);
    return false;
  }
}

// Handle HTTP requests
addEventListener('fetch', event => {
  event.respondWith(handleRequest(event.request));
});

async function handleRequest(request) {
  const url = new URL(request.url);
  const path = url.pathname;

  // Ensure WASM is initialized
  if (!wasmInitialized) {
    const initialized = await initWasm();
    if (!initialized) {
      return new Response(JSON.stringify({
        error: 'Failed to initialize service'
      }), {
        status: 500,
        headers: { 'Content-Type': 'application/json' }
      });
    }
  }

  // CORS headers
  const headers = {
    'Content-Type': 'application/json',
    'Access-Control-Allow-Origin': '*',
    'Access-Control-Allow-Methods': 'GET, POST, OPTIONS',
    'Access-Control-Allow-Headers': 'Content-Type, Authorization',
  };

  // Handle OPTIONS (preflight)
  if (request.method === 'OPTIONS') {
    return new Response(null, { status: 200, headers });
  }

  // Check authentication if AUTH_KEY is set
  const authKey = env.AUTH_KEY;
  if (authKey) {
    const authHeader = request.headers.get('Authorization');
    const token = authHeader ? authHeader.replace('Bearer ', '') : '';

    if (token !== authKey) {
      return new Response(JSON.stringify({
        code: 'UNAUTHORIZED',
        message: 'Invalid or missing API key'
      }), {
        status: 401,
        headers
      });
    }
  }

  // Route requests
  if (path === '/ping') {
    return new Response(JSON.stringify({
      status: 'ok',
      version: '2.0.0',
      platform: 'cloudflare-workers'
    }), {
      status: 200,
      headers
    });
  }

  if (path === '/multi') {
    // Multi-domain lookup
    const domainsParam = url.searchParams.get('domains');
    if (!domainsParam) {
      return new Response(JSON.stringify({
        code: 'INVALID_REQUEST',
        message: 'Missing domains parameter'
      }), {
        status: 400,
        headers
      });
    }

    const domains = domainsParam.split(',');
    const result = globalThis.handleMultiWhoisLookup(domains);

    return new Response(result, {
      status: 200,
      headers
    });
  }

  // Single domain lookup
  const domain = path.substring(1); // Remove leading slash
  if (!domain) {
    return new Response(JSON.stringify({
      service: 'Who-Dat WHOIS Lookup API',
      usage: {
        single: 'GET /{domain}',
        multi: 'GET /multi?domains=domain1.com,domain2.com',
        health: 'GET /ping'
      },
      docs: 'https://github.com/lissy93/who-dat'
    }), {
      status: 200,
      headers
    });
  }

  const result = globalThis.handleWhoisLookup(domain);

  return new Response(result, {
    status: 200,
    headers
  });
}
