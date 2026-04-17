import { createMDX } from 'fumadocs-mdx/next';

const withMDX = createMDX();

/** @type {import('next').NextConfig} */
const basePath = process.env.NEXT_PUBLIC_BASE_PATH || '';
const assetPrefix = basePath ? `${basePath}/` : '';

const config = {
  output: 'export',
  reactStrictMode: true,
  basePath,
  assetPrefix,
  images: {
    unoptimized: true,
  },
};

export default withMDX(config);
