function normalizeBasePath(basePath: string | null | undefined): string {
  if (!basePath) {
    return ''
  }

  const trimmed = basePath.trim()
  if (!trimmed || trimmed === '/') {
    return ''
  }

  const withLeadingSlash = trimmed.startsWith('/') ? trimmed : `/${trimmed}`
  return withLeadingSlash.replace(/\/+$/, '')
}

function getInjectedBasePath(): string {
  if (typeof window === 'undefined') {
    return ''
  }

  return normalizeBasePath((window as Window & { __BOARD_BASE__?: string }).__BOARD_BASE__)
}

function getBasePathFromAssetURL(): string {
  const assetURL = new URL(import.meta.url)
  const assetPath = assetURL.pathname
  const assetsDir = assetPath.lastIndexOf('/assets/')

  if (assetsDir <= 0) {
    return ''
  }

  return normalizeBasePath(assetPath.slice(0, assetsDir))
}

export function getBoardBasePath(): string {
  return getInjectedBasePath() || getBasePathFromAssetURL()
}

export function getBoardBaseHref(): string {
  const basePath = getBoardBasePath()
  return basePath ? `${basePath}/` : '/'
}
