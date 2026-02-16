import { createMDX } from 'fumadocs-mdx/next';

const withMDX = createMDX();

/** @type {import('next').NextConfig} */
const config = {
  output: 'export',
  reactStrictMode: true,
  basePath: '/2026-1a/t12/g05',
  assetPrefix: '/2026-1a/t12/g05/',
  images: {
    unoptimized: true,
  },
};

export default withMDX(config);
