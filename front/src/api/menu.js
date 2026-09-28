import mpx from '@mpxjs/core'

export const DEFAULT_MERCHANT_ID = process.env.VUE_APP_MERCHANT_ID || 'demo-merchant'
export const DEFAULT_STORE_ID = 'default'

const apiBaseUrl = process.env.VUE_APP_API_BASE_URL || 'https://mendian.zhifool.com'
const cartStorageKey = 'mendian-cart-id'
const storeStorageKey = 'mendian-store-id'

export function getAssetUrl(imagePath) {
  if (!imagePath) return ''
  if (/^https?:\/\//i.test(imagePath)) return imagePath
  const path = String(imagePath).replace(/^\/+/, '')
  const assetPath = path.startsWith('assets/') ? path : `assets/${path}`
  return `${apiBaseUrl.replace(/\/$/, '')}/${assetPath}`
}

export function getCartId(merchantId, storeId) {
  if (!merchantId || !storeId) throw new Error('merchantId and storeId are required')
  const scopedCartStorageKey = `${cartStorageKey}:${merchantId}:${storeId}`
  let cartId = mpx.getStorageSync(scopedCartStorageKey)
  if (!cartId || typeof cartId !== 'string') {
    cartId = `cart-${Date.now().toString(36)}-${Math.random().toString(36).slice(2)}`
  }
  mpx.setStorageSync(scopedCartStorageKey, cartId)
  return cartId
}

export function getSelectedStoreId(merchantId = DEFAULT_MERCHANT_ID) {
  return mpx.getStorageSync(`${storeStorageKey}:${merchantId}`) || ''
}

export function saveSelectedStoreId(merchantId, storeId) {
  mpx.setStorageSync(`${storeStorageKey}:${merchantId}`, storeId)
}

function request(path, options = {}) {
  return new Promise((resolve, reject) => {
    mpx.request({
      url: `${apiBaseUrl}${path}`,
      method: options.method || 'GET',
      data: options.data,
      header: { 'content-type': 'application/json' },
      success(response) {
        if (response.statusCode >= 200 && response.statusCode < 300) {
          resolve(response.data)
          return
        }
        reject(new Error(response.data && response.data.error ? response.data.error : `请求失败 (${response.statusCode})`))
      },
      fail(error) {
        reject(new Error(error && error.errMsg ? error.errMsg : '网络请求失败'))
      }
    })
  })
}

export function getMenu(merchantId, storeId) {
  return request(withStore(`/api/merchants/${encodeURIComponent(merchantId)}/menu`, storeId))
}

export function getStores(merchantId) {
  return request(`/api/merchants/${encodeURIComponent(merchantId)}/stores`)
}

function withStore(path, storeId) {
  if (!storeId) throw new Error('storeId is required')
  return `${path}?storeId=${encodeURIComponent(storeId)}`
}

export function getCartItems(merchantId, cartId, storeId) {
  return request(withStore(`/api/merchants/${encodeURIComponent(merchantId)}/carts/${encodeURIComponent(cartId)}/items`, storeId))
}

export function createCartItem(merchantId, cartId, data, storeId) {
  return request(withStore(`/api/merchants/${encodeURIComponent(merchantId)}/carts/${encodeURIComponent(cartId)}/items`, storeId), {
    method: 'POST',
    data
  })
}

export function updateCartItem(merchantId, cartId, itemId, quantity, storeId) {
  return request(withStore(`/api/merchants/${encodeURIComponent(merchantId)}/carts/${encodeURIComponent(cartId)}/items/${encodeURIComponent(itemId)}`, storeId), {
    method: 'PATCH',
    data: { quantity }
  })
}

export function removeCartItem(merchantId, cartId, itemId, storeId) {
  return request(withStore(`/api/merchants/${encodeURIComponent(merchantId)}/carts/${encodeURIComponent(cartId)}/items/${encodeURIComponent(itemId)}`, storeId), {
    method: 'DELETE'
  })
}

export function loginWithWechat(merchantId, code) {
  return request('/api/auth/wechat/login', {
    method: 'POST',
    data: { merchantId, code }
  })
}

export function getMyPaymentOrders(merchantId, code, cursor = null) {
  return request(`/api/merchants/${encodeURIComponent(merchantId)}/orders/mine`, {
    method: 'POST',
    data: { code, cursor }
  })
}

export function getUserAddresses(merchantId, code) {
  return request(`/api/merchants/${encodeURIComponent(merchantId)}/addresses/list`, {
    method: 'POST',
    data: { code }
  })
}

export function createUserAddress(merchantId, code, address) {
  return request(`/api/merchants/${encodeURIComponent(merchantId)}/addresses`, {
    method: 'POST',
    data: { ...address, code }
  })
}

export function updateUserAddress(merchantId, addressId, code, address) {
  return request(`/api/merchants/${encodeURIComponent(merchantId)}/addresses/${encodeURIComponent(addressId)}`, {
    method: 'PATCH',
    data: { ...address, code }
  })
}

export function deleteUserAddress(merchantId, addressId, code) {
  return request(`/api/merchants/${encodeURIComponent(merchantId)}/addresses/${encodeURIComponent(addressId)}`, {
    method: 'DELETE',
    data: { code }
  })
}

export function createPaymentOrder(merchantId, cartId, code, storeId, fulfillmentType) {
  return request(withStore(`/api/merchants/${encodeURIComponent(merchantId)}/carts/${encodeURIComponent(cartId)}/orders`, storeId), {
    method: 'POST',
    data: { code, fulfillmentType }
  })
}

export function getPaymentOrder(merchantId, cartId, orderNo, storeId) {
  return request(withStore(`/api/merchants/${encodeURIComponent(merchantId)}/carts/${encodeURIComponent(cartId)}/orders/${encodeURIComponent(orderNo)}`, storeId))
}
