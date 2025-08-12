/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: false,
  async rewrites() {
    return [
      {
        source: '/transport/:path*',
        destination: `http://${process.env.BACKEND_HOST}/:path*`,
      },
    ];
  },
};

export default nextConfig;
