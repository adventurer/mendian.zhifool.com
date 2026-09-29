# Iris Automatic TLS

Set `server.autoTlsDomain` in `config.yaml` or `SERVER_AUTOTLS_DOMAIN` to a public DNS name that resolves to the server; optionally set `server.autoTlsEmail` or `SERVER_AUTOTLS_EMAIL` for certificate expiry notices. The local config enables `mendian.zhifool.com`. Iris obtains and renews certificates automatically, caches them in `backend/letscache`, listens on ports 80 and 443, and redirects HTTP to HTTPS. Both ports must be reachable by Let's Encrypt. A reverse tunnel must forward public ports 80 and 443 to Iris on local ports 80 and 443; forwarding port 80 to the app's 8081 HTTP listener will bypass the ACME challenge. This mode cannot issue certificates for localhost or run behind another TLS terminator. For local HTTP development, remove `autoTlsDomain` and leave `SERVER_AUTOTLS_DOMAIN` unset. Do not configure `tlsCert` or `tlsKey` together with AutoTLS.

# Backend

## WeChat Pay

The API creates payment orders from server-side cart prices and starts WeChat Mini Program JSAPI payment. Configure the following values in `config.yaml` or with the corresponding environment variables:

- `WECHAT_PAY_MERCHANT_ID`: merchant ID used by this API, for example `demo-merchant`.
- `WECHAT_APP_ID`: Mini Program app ID; it must match the app ID in `front/static/wx/project.config.json`.
- `WECHAT_APP_SECRET`: Mini Program app secret, used only by the backend to exchange login codes for OpenIDs.
- `WECHAT_MCH_ID`: WeChat Pay merchant number.
- `WECHAT_MCH_CERTIFICATE_SERIAL`: serial number of the merchant API certificate.
- `WECHAT_PAY_PUBLIC_KEY_ID` and `WECHAT_PAY_PUBLIC_KEY_PATH`: optional pair for merchants using WeChat Pay public-key authentication. Use the key ID and PEM public key downloaded from the merchant platform. When set, this mode is used instead of automatic platform-certificate download.
- `WECHAT_PAY_API_V3_KEY`: 32-byte API v3 key.
- `WECHAT_MCH_PRIVATE_KEY_PATH`: path to the merchant API private key on the backend host.
- `WECHAT_PAY_NOTIFY_URL`: publicly reachable HTTPS URL in the form `https://<public-host>/api/payment/wechat/notify/<merchant-id>`. It must not include a query string.

Keep all secrets and key files on the backend host. Do not put them in the Mini Program or commit them to the repository. Restart the backend after setting the values. If payment configuration is incomplete or cannot initialize, order creation responds with HTTP 503 while health, menu, and cart APIs remain available. Merchants using platform certificates can omit the optional public-key pair.

One backend can serve multiple Mini Programs by replacing the top-level single-merchant values with a `wechatPay.merchants` map keyed by tenant ID. The map key is the internal `merchantId`; each value needs its own `appId`, `appSecret`, WeChat Pay merchant credentials, private key, and callback URL. The callback path must end with the same tenant ID. Do not configure both the legacy top-level credentials and the merchant map; the map takes precedence.

```yaml
wechatPay:
	merchants:
		shop-a:
			appId: wx-app-a
			appSecret: <shop-a-app-secret>
			mchId: <shop-a-mch-id>
			certificateSerial: <shop-a-certificate-serial>
			publicKeyId: <shop-a-public-key-id>
			publicKeyPath: /secure/keys/shop-a-wechat-public.pem
			apiV3Key: <shop-a-32-byte-api-v3-key>
			privateKeyPath: /secure/keys/shop-a-mch-private.pem
			notifyUrl: https://api.example.com/api/payment/wechat/notify/shop-a
		shop-b:
			appId: wx-app-b
			appSecret: <shop-b-app-secret>
			mchId: <shop-b-mch-id>
			certificateSerial: <shop-b-certificate-serial>
			publicKeyId: <shop-b-public-key-id>
			publicKeyPath: /secure/keys/shop-b-wechat-public.pem
			apiV3Key: <shop-b-32-byte-api-v3-key>
			privateKeyPath: /secure/keys/shop-b-mch-private.pem
			notifyUrl: https://api.example.com/api/payment/wechat/notify/shop-b
```

Build each Mini Program with its merchant ID so menu, login, cart, order, and payment calls use the matching tenant. For example from the repository root in PowerShell: `$env:VUE_APP_MERCHANT_ID='shop-a'; npm --prefix front run build:wx`. The Mini Program's own WeChat AppID must match that merchant's `appId`. Every configured merchant must also have its own merchant/menu/product data in the database; payment credentials alone do not create menu data. Keep all credentials and key files on the backend host and out of source control.

Local backend development uses plain HTTP at `http://localhost:8081`; no local TLS certificate is required. For deployment, build the Mini Program with `VUE_APP_API_BASE_URL` set to the backend's public HTTPS origin, and add that domain to the Mini Program's request-domain allowlist in the WeChat platform. Localhost is not suitable for real Mini Program payments.

The Mini Program creates an order with `POST /api/merchants/{merchantId}/carts/{cartId}/orders` and a WeChat `login` code. The backend recalculates the amount from stored cart items, creates the WeChat prepayment, and returns the signed payment parameters. Payment is recorded only after the signed WeChat notification is validated; the callback then removes the purchased cart items.

The Mini Program can retrieve the current user's order statuses with `POST /api/merchants/{merchantId}/orders/mine` and a fresh WeChat `login` code. The backend resolves the code to OpenID and filters orders by merchant and payer; the response never includes OpenID. Login codes are single-use and must not be reused for payment or another status request.

## Address Book

The address book uses `POST /api/merchants/{merchantId}/addresses/list`, `POST /api/merchants/{merchantId}/addresses`, `PATCH /api/merchants/{merchantId}/addresses/{addressId}`, and `DELETE /api/merchants/{merchantId}/addresses/{addressId}`. Every request carries a fresh WeChat login `code` in its JSON body. Addresses are scoped by merchant, Mini Program AppID, and verified OpenID; ownership identifiers are never returned. The first saved address becomes the default, and deleting a default promotes the most recently updated remaining address.

## Stores

Each merchant can have multiple physical stores in `merchant_stores`; `GET /api/merchants/{merchantId}/stores` returns active stores that have a catalog. Store catalogs are normalized in `store_menus`, `store_menu_series`, `store_products`, `store_product_options`, and `store_product_option_values`, keyed by merchant and store. Startup seeds the existing `default` store and its menu, plus a clearly marked `demo-store` with three zero-price sample products. The demo store has placeholder address/coordinates and is blocked from payment.

On the first backend restart after this upgrade, existing JSON or normalized merchant menus are copied into each merchant's `default` store catalog. Only after that copy succeeds, the legacy `menus`, `menu_series`, `products`, `product_options`, and `product_option_values` tables are dropped. Back up the database before running this one-time migration.

To register a real store, provision its `merchant_stores` row and store-specific catalog rows together; use a unique `store_id` per merchant, set its real address, phone, coordinates, `is_active = 1`, and `is_default = 0`. Keep exactly one default store per merchant. There is no public store-write API; store provisioning is an operator action. Cart and payment requests accept a `storeId` query parameter, and cart IDs/orders are scoped to the selected store. “My orders” returns orders across all stores. Stores share the merchant's WeChat Pay gateway; store-specific payment settlement is not implemented.

Coffee/storefront images and category SVGs live in `backend/assets` and are served by Iris at `/assets/<relative-path>`. Keep `store_products.image` values such as `coffee/rose-latte.jpg` and `store_menu_series.icon` values such as `icons/coffee.svg` relative; the Mini Program resolves both against `VUE_APP_API_BASE_URL`. Add the backend host to the Mini Program's download-file domain allowlist.

# dev
The remote SSH server must allow remote TCP forwarding with `GatewayPorts clientspecified`, and public ports 80 and 443 must be unused and reachable. Start Iris with the checked-in local AutoTLS domain config, then start this tunnel from the same machine:

```sh
go -C backend run ./cmd/server
ssh -N -o ExitOnForwardFailure=yes -R 0.0.0.0:80:127.0.0.1:80 -R 0.0.0.0:443:127.0.0.1:443 root@106.75.143.252
```