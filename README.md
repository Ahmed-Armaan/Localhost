# Localhost

Expose your homelabs to the internet with minimal setup

## Serving React Apps Over a Proxy

When running a React app behind a proxy (like your daemon on `localhost:9000`), you need to configure the app to **load static files relative to the proxy path**, not the root `/`.  

##  Using Vite

Vite uses the `base` option in `vite.config.ts` or `vite.config.js` to set the base path for all assets.

### Example

```ts
// vite.config.ts
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  base: '/app/<your app name>/',  // <-- all assets will be served relative to this path
});
```

## Using CRA

For Create React App, set the base path for assets by adding a homepage field in your package.json. This ensures static files are loaded relative to the specified path.

### Example

```ts
// package.json
{
  "name": "your-app-name",
  "version": "0.1.0",
  "homepage": "/app/<your app name>/",
  "dependencies": {
    ...
  }
}
```
