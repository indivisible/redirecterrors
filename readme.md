# Redirect On Errors

A [Traefik](https://traefik.io) middleware plugin to redirect to another URL on specific HTTP statuses.

Similar to the built-in `Errors` middleware, but this generates a HTTP 302 redirect, instead of an internal proxy action.

It was created to make `ForwardAuth` easier to use.

### Example Configuration

Example configuration with `ForwardAuth`:

```yaml
# Static configuration

experimental:
  plugins:
    redirectErrors:
      moduleName: github.com/indivisible/redirecterrors
      version: v0.1.0
```

```yaml
# Dynamic configuration

http:
  routers:
    secured-router:
      rule: host(`secured.localhost`)
      service: service-secured
      middlewares:
        - auth-redirect
    auth-router:
      rule: host(`auth.localhost`)
      service: service-auth
      middlewares:
        - my-plugin

  services:
    service-secured:
      loadBalancer:
        servers:
          - url: http://localhost:5001/
    service-auth:
      loadBalancer:
        servers:
          - url: http://localhost:5002/

  middlewares:
    auth-redirect-error:
      plugin:
        redirectErrors:
          status:
            - "401"
          target: "http://auth.localhost/oauth2/sign_in?rd={url}"
          outputStatus: 302
    auth-check:
      forwardAuth:
        address: "http://localhost:5002/oauth2/auth"
        trustForwardHeader: true
    auth-redirect:
      chain:
        - auth-redirect-error
        - auth-check
```

### Configuration Options:

- `status`: list of statuses / status ranges (eg `401-403`). See the [Error middleware's description](https://doc.traefik.io/traefik/middlewares/http/errorpages/#status) for details.
- `target`: redirect target URL. `{status}` will be replaced with the original HTTP status code, and `{url}` will be replaced with the url-safe version of the original, full URL.
- `outputStatus`: HTTP code for the redirect. Default is 302.
