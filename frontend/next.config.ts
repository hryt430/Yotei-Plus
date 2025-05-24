import type { NextConfig } from 'next'

const nextConfig: NextConfig = {
  // 実験的機能
  experimental: {
    // App Router最適化
    optimizePackageImports: ['lucide-react', '@radix-ui/react-toast'],
  },
  // 画像最適化
  images: {
    remotePatterns: [
      {
        protocol: 'https',
        hostname: '**',
      },
    ],
    // 画像フォーマット
    formats: ['image/webp', 'image/avif'],
  },
  // パフォーマンス最適化
  compiler: {
    // Remove console.log in production
    removeConsole: process.env.NODE_ENV === 'production',
  },
  // 環境変数の型安全性
  env: {
    CUSTOM_KEY: process.env.CUSTOM_KEY,
  },
  // セキュリティヘッダー
  async headers() {
    return [
      {
        // セキュリティヘッダーを全ページに適用
        source: '/(.*)',
        headers: [
          {
            key: 'X-Frame-Options',
            value: 'DENY',
          },
          {
            key: 'X-Content-Type-Options',
            value: 'nosniff',
          },
          {
            key: 'Referrer-Policy',
            value: 'origin-when-cross-origin',
          },
          {
            key: 'Permissions-Policy',
            value: 'camera=(), microphone=(), geolocation=()',
          },
        ],
      },
    ]
  },
  // リダイレクト設定
  async redirects() {
    return [
      {
        source: '/login',
        destination: '/auth/login',
        permanent: true,
      },
      {
        source: '/register',
        destination: '/auth/register',
        permanent: true,
      },
      {
        source: '/signup',
        destination: '/auth/register',
        permanent: true,
      },
    ]
  },
  // TypeScript設定
  typescript: {
    // 本番ビルド時にTypeScriptエラーを無視しない
    ignoreBuildErrors: false,
  },
  // ESLint設定
  eslint: {
    // 本番ビルド時にESLintエラーを無視しない
    ignoreDuringBuilds: false,
  },
  // 出力設定
  output: process.env.BUILD_STANDALONE === 'true' ? 'standalone' : undefined,
  // 圧縮設定
  compress: true,
  // 電力効率
  poweredByHeader: false,
  // 開発環境での高速リフレッシュ
  reactStrictMode: true,
  // Bundle Analyzer（開発時のみ）
  ...(process.env.ANALYZE === 'true' && {
    webpack: (config, { isServer }) => {
      if (!isServer) {
        // eslint-disable-next-line @typescript-eslint/no-require-imports
        const { BundleAnalyzerPlugin } = require('webpack-bundle-analyzer')
        config.plugins.push(
          new BundleAnalyzerPlugin({
            analyzerMode: 'static',
            openAnalyzer: false,
          })
        )
      }
      return config
    },
  }),
}

export default nextConfig