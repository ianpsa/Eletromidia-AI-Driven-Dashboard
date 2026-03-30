import { createMDX } from 'fumadocs-mdx/next';

const withMDX = createMDX();
const isProduction = process.env.NODE_ENV === 'production';

/** @type {import('next').NextConfig} */
const config = {
  output: 'export',
  reactStrictMode: true,
  basePath: isProduction ? '/2026-1a/t12/g05' : '',
  assetPrefix: isProduction ? '/2026-1a/t12/g05/' : undefined,
  images: {
    unoptimized: true,
  },
};

export default withMDX(config);
