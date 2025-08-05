const { createProxyMiddleware } = require('http-proxy-middleware');

module.exports = function(app) {
  // Proxy API requests to the Go backend
  app.use(
    '/api',
    createProxyMiddleware({
      target: 'http://100.113.170.10:3000',
      changeOrigin: true,
      secure: false,
      logLevel: 'debug',
      onError: (err, req, res) => {
        console.error('Proxy error:', err);
      },
      onProxyRes: (proxyRes, req, res) => {
        console.log('Proxy response:', req.url, proxyRes.statusCode);
      }
    })
  );

  // Proxy any other backend routes
  app.use(
    ['/auth', '/v1'],
    createProxyMiddleware({
      target: 'http://100.113.170.10:3000',
      changeOrigin: true,
      secure: false,
      logLevel: 'debug',
      onError: (err, req, res) => {
        console.error('Proxy error:', err);
      },
      onProxyRes: (proxyRes, req, res) => {
        console.log('Proxy response:', req.url, proxyRes.statusCode);
      }
    })
  );
};
