import { type NextRequest, NextResponse } from 'next/server'

const API_URL = process.env.API_URL ?? 'http://127.0.0.1:8080'

const HOP_BY_HOP = new Set([
  'connection',
  'keep-alive',
  'proxy-authenticate',
  'proxy-authorization',
  'te',
  'trailers',
  'transfer-encoding',
  'upgrade',
])

async function proxy(req: NextRequest, path: string[]) {
  const target = new URL(`/api/${path.join('/')}`, API_URL)
  target.search = req.nextUrl.searchParams.toString()

  const headers = new Headers()
  req.headers.forEach((value, key) => {
    if (HOP_BY_HOP.has(key.toLowerCase())) return
    if (key.toLowerCase() === 'host') return
    headers.set(key, value)
  })

  const init: RequestInit = {
    method: req.method,
    headers,
    redirect: 'manual',
  }

  if (req.method !== 'GET' && req.method !== 'HEAD') {
    init.body = await req.arrayBuffer()
  }

  const upstream = await fetch(target.toString(), init)
  const responseHeaders = new Headers()

  upstream.headers.forEach((value, key) => {
    if (HOP_BY_HOP.has(key.toLowerCase())) return
    responseHeaders.append(key, value)
  })

  return new NextResponse(upstream.body, {
    status: upstream.status,
    statusText: upstream.statusText,
    headers: responseHeaders,
  })
}

type RouteContext = { params: Promise<{ path: string[] }> }

export async function GET(req: NextRequest, ctx: RouteContext) {
  const { path } = await ctx.params
  return proxy(req, path)
}

export async function POST(req: NextRequest, ctx: RouteContext) {
  const { path } = await ctx.params
  return proxy(req, path)
}

export async function PATCH(req: NextRequest, ctx: RouteContext) {
  const { path } = await ctx.params
  return proxy(req, path)
}

export async function DELETE(req: NextRequest, ctx: RouteContext) {
  const { path } = await ctx.params
  return proxy(req, path)
}

export async function OPTIONS(req: NextRequest, ctx: RouteContext) {
  const { path } = await ctx.params
  return proxy(req, path)
}