const express = require("express");
const { createProxyMiddleware } = require("http-proxy-middleware");

const app = express();

app.use(
  "/",
  createProxyMiddleware({
    target: "http://34.39.196.243",
    changeOrigin: true,
    on: {
      proxyReq: (proxyReq, req) => {
        // Repassa o Authorization explicitamente
        const auth = req.headers["authorization"];
        if (auth) {
          proxyReq.setHeader("Authorization", auth);
        }
      },
    },
  })
);

const PORT = process.env.PORT || 8080;
app.listen(PORT, () => console.log(`Proxy rodando na porta ${PORT}`));